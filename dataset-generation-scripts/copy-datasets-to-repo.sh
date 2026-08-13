#!/usr/bin/env bash
set -euo pipefail

# One-shot copy of generated dataset outputs (dataset-generation-scripts/output/)
# into the repository's published folders (json/, postgresql/, ...).
#
# Run this AFTER regenerating the dataset (go run main.go) so the output
# contains the freshly generated files.
#
# Usage:
#   ./copy-datasets-to-repo.sh            # copy everything
#   ./copy-datasets-to-repo.sh --dry-run  # preview what would be copied
#   ./copy-datasets-to-repo.sh json       # copy only one dataset

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"

DRY_RUN=0
DATASETS=()
for arg in "$@"; do
  case "$arg" in
    --dry-run|-n) DRY_RUN=1 ;;
    *) DATASETS+=("$arg") ;;
  esac
done

# Matches the datetime suffix segment used by generated files,
# e.g. "_2026-08-13__22_43_48"
readonly DATETIME_SUFFIX_RE='_2[0-9]{3}-[0-9]{2}-[0-9]{2}__[0-9]{2}_[0-9]{2}_[0-9]{2}'

run() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "[dry-run] $*"
  else
    "$@"
  fi
}

# copy_datasets <output_subdir> <repo_target_dir> [--strip-datetime]
#   Copies every top-level entry (files + subdirs) from output/<output_subdir>
#   into <repo_target_dir>, overwriting existing files.
#   With --strip-datetime, the datetime suffix is removed from copied filenames
#   (used where the published root files have no suffix).
copy_datasets() {
  local src="$OUTPUT_DIR/$1"
  local dst="$REPO_ROOT/$2"
  local strip="${3:-}"

  if [[ ! -d "$src" ]]; then
    echo "skip: $1 not found in output/"
    return 0
  fi

  echo "== $1 -> $2 =="
  run mkdir -p "$dst"

  if [[ "$strip" == "--strip-datetime" ]]; then
    while IFS= read -r -d '' entry; do
      local name new_name
      name="$(basename "$entry")"
      new_name="$(printf '%s' "$name" | sed -E "s/${DATETIME_SUFFIX_RE}//")"
      if [[ "$new_name" == "$name" ]]; then
        run cp -R "$entry" "$dst/"
      else
        run cp -R "$entry" "$dst/$new_name"
      fi
    done < <(find "$src" -mindepth 1 -maxdepth 1 -print0)
  else
    run cp -R "$src/." "$dst/"
  fi
}

# Prune stale timestamped geojson archives from json/ so only the fixed-name
# zip (vn_provinces_wards_geojson.zip) remains after copying.
prune_stale_geojson_zips() {
  local stale_zips=("$REPO_ROOT"/json/vn_provinces_wards_geojson_*.zip)
  if [[ -f "${stale_zips[0]:-}" ]]; then
    for z in "${stale_zips[@]}"; do
      [[ "$(basename "$z")" == "vn_provinces_wards_geojson.zip" ]] && continue
      echo "remove stale: $z"
      run rm "$z"
    done
  fi
}

copy_all() {
  # JSON: filenames are deterministic (no suffix). Also prune old zips.
  copy_datasets json json
  prune_stale_geojson_zips

  # SQL engines: root import files are deterministic; GIS subfolders keep
  # timestamped part files (matches the published convention).
  copy_datasets postgresql postgresql
  copy_datasets mysql mysql
  copy_datasets sqlserver sqlserver
  copy_datasets oracle oracle

  # MongoDB / Redis: output root files carry a datetime suffix that the
  # published root files do not, so strip it.
  copy_datasets mongodb mongodb --strip-datetime
  copy_datasets redis redis --strip-datetime

  # Elasticsearch: output mirrors the published folder exactly.
  copy_datasets elasticsearch elasticsearch
}

if [[ "${#DATASETS[@]}" -eq 0 ]]; then
  copy_all
else
  case "${DATASETS[0]}" in
    json)            copy_datasets json json; prune_stale_geojson_zips ;;
    postgresql)      copy_datasets postgresql postgresql ;;
    mysql)           copy_datasets mysql mysql ;;
    sqlserver)       copy_datasets sqlserver sqlserver ;;
    oracle)          copy_datasets oracle oracle ;;
    mongodb)         copy_datasets mongodb mongodb --strip-datetime ;;
    redis)           copy_datasets redis redis --strip-datetime ;;
    elasticsearch)   copy_datasets elasticsearch elasticsearch ;;
    *) echo "unknown dataset: ${DATASETS[0]}" >&2; exit 1 ;;
  esac
fi

echo
echo "Done."
