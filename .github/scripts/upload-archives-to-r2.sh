#!/usr/bin/env bash
set -euo pipefail

# Uploads the packaged archives and downloads.json manifest to Cloudflare R2,
# then verifies that the public CDN URLs are reachable (warns only).
#
# Verification issues a 1-byte range request per archive so it does not download
# the whole file (a plain GET pulled 40-100 MB per URL). The CDN answers with
# HTTP 206; a server that ignores Range still answers HTTP 200, which also passes.
#
# Usage:
#   upload-archives-to-r2.sh <version> <archives-dir>
#
# Required env:
#   R2_BUCKET_NAME
#   AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
#   AWS_ENDPOINT_URL_S3   e.g. https://<account>.r2.cloudflarestorage.com
#   AWS_DEFAULT_REGION    must be "auto" for R2

VERSION="${1:?usage: upload-archives-to-r2.sh <version> <archives-dir>}"
ARCHIVES_DIR="${2:?usage: upload-archives-to-r2.sh <version> <archives-dir>}"
: "${R2_BUCKET_NAME:?R2_BUCKET_NAME is required}"

MANIFEST="${ARCHIVES_DIR}/downloads.json"
[[ -f "$MANIFEST" ]] || { echo "error: manifest not found: $MANIFEST" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "error: jq is required" >&2; exit 1; }
command -v aws >/dev/null 2>&1 || { echo "error: aws-cli is required" >&2; exit 1; }

echo "Uploading archives for ${VERSION} to s3://${R2_BUCKET_NAME}/"
while IFS=$'\t' read -r archive dataset_id; do
  echo "::group::Uploading ${dataset_id}/${archive}"
  aws s3 cp "${ARCHIVES_DIR}/${archive}" \
    "s3://${R2_BUCKET_NAME}/${VERSION}/${dataset_id}/${archive}" \
    --no-progress
  echo "::endgroup::"
done < <(jq -r '.datasets[] | [.archive, .id] | @tsv' "$MANIFEST")

aws s3 cp "$MANIFEST" "s3://${R2_BUCKET_NAME}/${VERSION}/downloads.json" --no-progress

echo
echo "Verifying CDN URLs (1-byte range request)"
fail=0
while read -r url; do
  # Request a single byte so the check does not download the entire archive.
  # 206 = range honoured, 200 = range ignored (full body) — both mean reachable.
  code="$(curl -s -o /dev/null -w '%{http_code}' -r 0-0 --max-time 30 "$url")"
  echo "  ${code} ${url}"
  [[ "$code" == "200" || "$code" == "206" ]] || fail=1
done < <(jq -r '.datasets[].url' "$MANIFEST")
if [[ "$fail" -ne 0 ]]; then
  echo "::warning::One or more CDN URLs did not return HTTP 200/206. Check the bucket's public access / custom domain configuration."
fi
