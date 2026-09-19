#!/usr/bin/env bash
set -euo pipefail

# Uploads the packaged archives and downloads.json manifest to Cloudflare R2,
# then verifies that the public CDN URLs return HTTP 200 (warns only).
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
echo "Verifying CDN URLs"
fail=0
while read -r url; do
  code="$(curl -s -o /dev/null -w '%{http_code}' "$url")"
  echo "  ${code} ${url}"
  [[ "$code" == "200" ]] || fail=1
done < <(jq -r '.datasets[].url' "$MANIFEST")
if [[ "$fail" -ne 0 ]]; then
  echo "::warning::One or more CDN URLs did not return HTTP 200. Check the bucket's public access / custom domain configuration."
fi
