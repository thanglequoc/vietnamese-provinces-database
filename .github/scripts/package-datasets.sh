#!/usr/bin/env bash
set -euo pipefail

# Packages each published dataset folder into a ZIP archive and writes a
# `downloads.json` manifest (archive name, size, sha256, public URL).
#
# Usage:
#   package-datasets.sh <version> <output-dir> [--repo-root DIR]
#
# Options:
#   --repo-root DIR   Repository root containing the dataset folders
#                     (default: this repository's root)
#
# Environment:
#   R2_PUBLIC_BASE_URL  Public CDN base URL
#                       (default: https://vn-provinces-ds.thanglequoc.xyz)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Published dataset folders. The folder name doubles as the R2 key prefix.
DATASETS=(
  postgresql
  mysql
  sqlserver
  oracle
  json
  mongodb
  redis
  elasticsearch
)

usage() {
  cat <<'EOF'
Packages each published dataset folder into a ZIP archive and writes a
`downloads.json` manifest (archive name, size, sha256, public URL).

Usage:
  package-datasets.sh <version> <output-dir> [--repo-root DIR]

Options:
  --repo-root DIR   Repository root containing the dataset folders
                    (default: this repository's root)

Environment:
  R2_PUBLIC_BASE_URL  Public CDN base URL
                      (default: https://vn-provinces-ds.thanglequoc.xyz)
EOF
}

VERSION=""
OUT_DIR=""
REPO_ROOT="$DEFAULT_REPO_ROOT"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-root) REPO_ROOT="${2:?--repo-root needs a value}"; shift 2 ;;
    -h|--help)   usage; exit 0 ;;
    -*)          echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
    *)
      if [[ -z "$VERSION" ]]; then VERSION="$1"
      elif [[ -z "$OUT_DIR" ]]; then OUT_DIR="$1"
      else echo "error: unexpected argument: $1" >&2; exit 2
      fi
      shift
      ;;
  esac
done

if [[ -z "$VERSION" || -z "$OUT_DIR" ]]; then
  echo "error: <version> and <output-dir> are required" >&2
  usage >&2
  exit 2
fi

[[ "$VERSION" == v* ]] || VERSION="v$VERSION"
R2_PUBLIC_BASE_URL="${R2_PUBLIC_BASE_URL:-https://vn-provinces-ds.thanglequoc.xyz}"

REPO_ROOT="$(cd "$REPO_ROOT" && pwd)"
mkdir -p "$OUT_DIR"
OUT_DIR="$(cd "$OUT_DIR" && pwd)"
GENERATED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

command -v zip >/dev/null 2>&1 || { echo "error: 'zip' is not installed" >&2; exit 1; }

sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

human_size() {
  awk -v b="$1" 'BEGIN {
    n = split("B KB MB GB TB", unit, " ");
    i = 1;
    while (b >= 1024 && i < n) { b /= 1024; i++ }
    if (i == 1) printf "%d %s", b, unit[i];
    else printf "%.2f %s", b, unit[i];
  }'
}

echo "Packaging datasets for ${VERSION}"
echo "  repo root : ${REPO_ROOT}"
echo "  output dir: ${OUT_DIR}"
echo "  base url  : ${R2_PUBLIC_BASE_URL}"
echo

json="$OUT_DIR/downloads.json"
{
  echo "{"
  printf '  "version": "%s",\n' "$VERSION"
  printf '  "generated_at": "%s",\n' "$GENERATED_AT"
  printf '  "base_url": "%s",\n' "${R2_PUBLIC_BASE_URL%/}"
  echo '  "datasets": ['
  first=1
  for id in "${DATASETS[@]}"; do
    if [[ ! -d "$REPO_ROOT/$id" ]]; then
      echo "error: dataset folder not found: $REPO_ROOT/$id" >&2
      exit 1
    fi

    archive="vn_provinces_${id}_dataset_${VERSION}.zip"
    path="$OUT_DIR/$archive"

    rm -f "$path"
    ( cd "$REPO_ROOT" && zip -r -X -9 -q "$path" "$id" \
        -x '*.DS_Store' -x '*/__MACOSX/*' -x '*/Thumbs.db' -x 'Thumbs.db' )
    bytes="$(wc -c < "$path" | tr -d ' ')"
    human="$(human_size "$bytes")"
    sha="$(sha256 "$path")"
    printf '  %-42s %10s  %s\n' "$archive" "$human" "$sha" >&2

    url="${R2_PUBLIC_BASE_URL%/}/$VERSION/$id/$archive"

    [[ "$first" -eq 1 ]] || echo ","
    first=0
    {
      echo "    {"
      printf '      "id": "%s",\n' "$id"
      printf '      "archive": "%s",\n' "$archive"
      printf '      "size_bytes": %s,\n' "$bytes"
      printf '      "size_human": "%s",\n' "$human"
      printf '      "sha256": "%s",\n' "$sha"
      printf '      "url": "%s"\n' "$url"
      echo -n "    }"
    }
  done
  echo
  echo "  ]"
  echo "}"
} > "$json"

echo
echo "Wrote ${json}"
