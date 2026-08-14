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

run() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "[dry-run] $*"
  else
    "$@"
  fi
}

# copy_datasets <output_subdir> <repo_target_dir>
#   Copies every top-level entry (files + subdirs) from output/<output_subdir>
#   into <repo_target_dir>, overwriting existing files.
copy_datasets() {
  local src="$OUTPUT_DIR/$1"
  local dst="$REPO_ROOT/$2"
  if [[ ! -d "$src" ]]; then
    echo "skip: $1 not found in output/"
    return 0
  fi
  echo "== $1 -> $2 =="
  run mkdir -p "$dst"
  run cp -R "$src/." "$dst/"
}

# prune_datetime_variants <repo_target_dir> <glob>
#   Removes timestamped artifact variants (matching <glob>) from <repo_target_dir>.
prune_datetime_variants() {
  local target="$REPO_ROOT/$1"
  local stale=("$target"/$2)
  if [[ -f "${stale[0]:-}" ]]; then
    for f in "${stale[@]}"; do
      echo "remove stale: $f"
      run rm "$f"
    done
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
  copy_datasets json json
  prune_stale_geojson_zips

  copy_datasets postgresql postgresql
  prune_datetime_variants postgresql/gis 'postgresql_ImportData_gis_*.sql*'

  copy_datasets mysql mysql
  prune_datetime_variants mysql/gis 'mysql_ImportData_gis_*.sql*'

  copy_datasets sqlserver sqlserver
  prune_datetime_variants sqlserver/gis 'mssql_ImportData_gis_*.sql*'

  copy_datasets oracle oracle

  copy_datasets mongodb mongodb
  prune_datetime_variants mongodb 'administrative_units_*.json'
  prune_datetime_variants mongodb 'administrative_regions_*.json'
  prune_datetime_variants mongodb 'mongo_data_vn_unit_*.json'
  prune_datetime_variants mongodb/gis 'mongo_data_vn_province_gis_*.json'
  prune_datetime_variants mongodb/gis 'mongo_data_vn_ward_gis_2*.json*'

  copy_datasets redis redis
  prune_datetime_variants redis 'redis_vn_provinces_dataset_*.redis'

  copy_datasets elasticsearch elasticsearch
}

if [[ "${#DATASETS[@]}" -eq 0 ]]; then
  copy_all
else
  case "${DATASETS[0]}" in
    json)          copy_datasets json json; prune_stale_geojson_zips ;;
    postgresql)    copy_datasets postgresql postgresql; prune_datetime_variants postgresql/gis 'postgresql_ImportData_gis_*.sql*' ;;
    mysql)         copy_datasets mysql mysql; prune_datetime_variants mysql/gis 'mysql_ImportData_gis_*.sql*' ;;
    sqlserver)     copy_datasets sqlserver sqlserver; prune_datetime_variants sqlserver/gis 'mssql_ImportData_gis_*.sql*' ;;
    oracle)        copy_datasets oracle oracle ;;
    mongodb)       copy_datasets mongodb mongodb; prune_datetime_variants mongodb 'administrative_units_*.json'; prune_datetime_variants mongodb 'administrative_regions_*.json'; prune_datetime_variants mongodb 'mongo_data_vn_unit_*.json'; prune_datetime_variants mongodb/gis 'mongo_data_vn_province_gis_*.json'; prune_datetime_variants mongodb/gis 'mongo_data_vn_ward_gis_2*.json*' ;;
    redis)         copy_datasets redis redis; prune_datetime_variants redis 'redis_vn_provinces_dataset_*.redis' ;;
    elasticsearch) copy_datasets elasticsearch elasticsearch ;;
    *) echo "unknown dataset: ${DATASETS[0]}" >&2; exit 1 ;;
  esac
fi

echo
echo "Done."
