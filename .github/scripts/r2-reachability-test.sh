#!/usr/bin/env bash
set -euo pipefail

# Verifies Cloudflare R2 credentials and that uploaded objects are reachable
# through the public custom domain. Uploads a tiny probe object, checks the
# CDN URL, then deletes the probe.
#
# Usage:
#   cp .env.r2.example .env.r2   # then fill in real values
#   ./.github/scripts/r2-reachability-test.sh
#
# Required env (loaded from .env.r2 or the ambient environment):
#   R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_NAME
# Optional:
#   R2_PUBLIC_BASE_URL (default: https://vn-provinces-ds.thanglequoc.xyz)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ENV_FILE="${REPO_ROOT}/.env.r2"

if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  set -a; source "$ENV_FILE"; set +a
  echo "Loaded credentials from $ENV_FILE"
else
  echo "No $ENV_FILE found — using ambient environment variables"
fi

: "${R2_ACCOUNT_ID:?R2_ACCOUNT_ID is required}"
: "${R2_ACCESS_KEY_ID:?R2_ACCESS_KEY_ID is required}"
: "${R2_SECRET_ACCESS_KEY:?R2_SECRET_ACCESS_KEY is required}"
: "${R2_BUCKET_NAME:?R2_BUCKET_NAME is required}"
: "${R2_PUBLIC_BASE_URL:=https://vn-provinces-ds.thanglequoc.xyz}"

export AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID"
export AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY"
export AWS_DEFAULT_REGION="auto"
export AWS_ENDPOINT_URL_S3="https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"

TEST_KEY="_reachability/hello.txt"
TMP_FILE="$(mktemp -t r2-probe.XXXXXX)"
trap 'rm -f "$TMP_FILE"' EXIT

printf 'vn-provinces-database R2 reachability probe %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$TMP_FILE"

echo
echo "== 1/4: credentials (list bucket) =="
aws s3 ls "s3://${R2_BUCKET_NAME}/" --no-cli-pager || {
  echo "FAIL: could not list bucket s3://${R2_BUCKET_NAME}/" >&2
  exit 1
}

echo
echo "== 2/4: upload probe object =="
aws s3 cp "$TMP_FILE" "s3://${R2_BUCKET_NAME}/${TEST_KEY}" --no-cli-pager

echo
echo "== 3/4: fetch probe via custom domain =="
PROBE_URL="${R2_PUBLIC_BASE_URL%/}/${TEST_KEY}"
HTTP_CODE="$(curl -s -o /dev/null -w '%{http_code}' "$PROBE_URL" || true)"
echo "GET $PROBE_URL -> HTTP $HTTP_CODE"

echo
echo "== 4/4: cleanup =="
aws s3 rm "s3://${R2_BUCKET_NAME}/${TEST_KEY}" --no-cli-pager

echo
if [[ "$HTTP_CODE" == "200" ]]; then
  echo "PASS: R2 credentials work and the custom domain serves objects."
else
  echo "WARN: credentials/upload work, but the custom domain returned HTTP ${HTTP_CODE}."
  echo "      Check the bucket's public access / custom domain configuration in Cloudflare."
  exit 2
fi
