#!/usr/bin/env bash
set -euo pipefail

# ────────────────────────────────────────────────────────────────────
# MongoDB Import & Verification Script
# Imports all 5 collections and runs 17 verification queries.
# Usage: ./import_and_verify_mongodb.sh
# Must be run from: dataset-generation-scripts/
# ────────────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="$SCRIPT_DIR/output/mongodb"
CONN_STRING="mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin"
DB_NAME="vn_provinces"
MONGOSH="mongosh $CONN_STRING/$DB_NAME --quiet"
MONGOIMPORT="mongoimport --uri=$CONN_STRING --db=$DB_NAME --jsonArray"

TOTAL_PASS=0
TOTAL_FAIL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info()  { echo -e "${YELLOW}[INFO]${NC} $*"; }
log_pass()  { echo -e "${GREEN}[PASS]${NC} $*"; }
log_fail()  { echo -e "${RED}[FAIL]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# ──────────────────────────────────────────────────
# Phase 0: Pre-flight
# ──────────────────────────────────────────────────

echo "=========================================="
echo " Phase 0: Pre-flight Checks"
echo "=========================================="

# Check binaries
log_info "Checking for mongoimport..."
command -v mongoimport >/dev/null 2>&1 || { log_error "mongoimport not found in PATH"; exit 1; }

log_info "Checking for mongosh..."
command -v mongosh >/dev/null 2>&1 || { log_error "mongosh not found in PATH"; exit 1; }

# Check tunnel / reachability
log_info "Checking MongoDB reachability on localhost:27017..."
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" localhost:27017 2>/dev/null || echo "000")
if [[ "$HTTP_STATUS" == "000" ]]; then
  log_info "MongoDB not reachable. Starting SSH tunnel..."
  ssh -f -N -L 27017:localhost:27017 -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz 2>/dev/null || {
    log_error "SSH tunnel failed. Check: ssh -v thanglequoc@machine.thanglequoc.xyz"
    exit 1
  }
  sleep 2
  log_info "Tunnel started. Verifying..."
fi

# Verify mongosh can connect
log_info "Verifying mongosh connection..."
$MONGOSH --eval "db.version()" >/dev/null 2>&1 || {
  log_error "mongosh cannot connect to MongoDB. Check tunnel and credentials."
  exit 1
}
log_info "MongoDB connection OK"

# Verify source files exist
log_info "Verifying source files..."
MANIFEST_FILE=$(ls "$DATA_DIR"/*ward_gis*.json.manifest 2>/dev/null | head -1)
if [[ -z "$MANIFEST_FILE" ]]; then
  log_error "No ward GIS manifest file found in $DATA_DIR"
  exit 1
fi

FILES_TO_CHECK=(
  "$DATA_DIR"/administrative_regions_*.json
  "$DATA_DIR"/administrative_units_*.json
  "$DATA_DIR"/mongo_data_vn_unit_*.json
  "$DATA_DIR"/mongo_data_vn_province_gis_*.json
  "$MANIFEST_FILE"
  "$DATA_DIR/create_indexes.js"
)

ADMIN_REG_FILE=$(ls "$DATA_DIR"/administrative_regions_*.json 2>/dev/null | head -1)
ADMIN_UNIT_FILE=$(ls "$DATA_DIR"/administrative_units_*.json 2>/dev/null | head -1)
VN_UNIT_FILE=$(ls "$DATA_DIR"/mongo_data_vn_unit_*.json 2>/dev/null | head -1)
PROV_GIS_FILE=$(ls "$DATA_DIR"/mongo_data_vn_province_gis_*.json 2>/dev/null | head -1)

for file in "$ADMIN_REG_FILE" "$ADMIN_UNIT_FILE" "$VN_UNIT_FILE" "$PROV_GIS_FILE" "$MANIFEST_FILE" "$DATA_DIR/create_indexes.js"; do
  if [[ ! -f "$file" ]]; then
    log_error "Missing: $file"
    exit 1
  fi
done

# Load ward part files from manifest
WARD_PARTS=()
while IFS= read -r line; do
  [[ -n "$line" ]] && WARD_PARTS+=("$DATA_DIR/$line")
done < "$MANIFEST_FILE"

for part_file in "${WARD_PARTS[@]}"; do
  if [[ ! -f "$part_file" ]]; then
    log_error "Missing ward part: $part_file"
    exit 1
  fi
done

log_info "All source files verified (${#WARD_PARTS[@]} ward parts)"
log_info "Source files OK"

# ──────────────────────────────────────────────────
# Phase 1: Collection Preparation
# ──────────────────────────────────────────────────

echo ""
echo "=========================================="
echo " Phase 1: Collection Preparation"
echo "=========================================="

COLLECTIONS=("provinces" "provinces-gis" "wards-gis" "administrative_regions" "administrative_units")

for coll in "${COLLECTIONS[@]}"; do
  log_info "Dropping collection: $coll"
  $MONGOSH --eval "db.getCollection('$coll').drop()" >/dev/null 2>&1 || true
done

log_info "Verifying all collections are empty..."
for coll in "${COLLECTIONS[@]}"; do
  COUNT=$($MONGOSH --eval "db.getCollection('$coll').countDocuments()" 2>/dev/null | tail -1)
  if [[ "$COUNT" != "0" ]]; then
    log_error "Collection '$coll' still has $COUNT docs after drop"
    exit 1
  fi
done
log_info "Database is clean"

# ──────────────────────────────────────────────────
# Phase 2: Import
# ──────────────────────────────────────────────────

echo ""
echo "=========================================="
echo " Phase 2: Import"
echo "=========================================="

import_collection() {
  local coll_name="$1"
  local file="$2"
  local expected_count="$3"

  log_info "Importing $coll_name from $(basename "$file")..."
  $MONGOIMPORT --collection="$coll_name" --file="$file" 2>&1 | tail -1

  local actual_count
  actual_count=$($MONGOSH --eval "db.getCollection('$coll_name').countDocuments()" 2>/dev/null | tail -1)
  if [[ "$actual_count" != "$expected_count" ]]; then
    log_error "$coll_name: expected $expected_count docs, got $actual_count"
    exit 1
  fi
  log_info "$coll_name: $actual_count documents imported"
}

import_collection "administrative_regions" "$ADMIN_REG_FILE" 8
import_collection "administrative_units" "$ADMIN_UNIT_FILE" 5
import_collection "provinces" "$VN_UNIT_FILE" 34
import_collection "provinces-gis" "$PROV_GIS_FILE" 34

# Import ward parts sequentially
log_info "Importing wards-gis (${#WARD_PARTS[@]} parts)..."
for i in "${!WARD_PARTS[@]}"; do
  log_info "  Part $((i+1))/${#WARD_PARTS[@]}: $(basename "${WARD_PARTS[$i]}")"
  $MONGOIMPORT --collection="wards-gis" --file="${WARD_PARTS[$i]}" 2>&1 | tail -1
done

WARD_COUNT=$($MONGOSH --eval "db.getCollection('wards-gis').countDocuments()" 2>/dev/null | tail -1)
if [[ "$WARD_COUNT" != "3321" ]]; then
  log_error "wards-gis: expected 3321 docs, got $WARD_COUNT"
  exit 1
fi
log_info "wards-gis: $WARD_COUNT documents imported"

log_info "Import phase complete"

# ──────────────────────────────────────────────────
# Phase 3: Index Creation
# ──────────────────────────────────────────────────

echo ""
echo "=========================================="
echo " Phase 3: Index Creation"
echo "=========================================="

log_info "Running create_indexes.js..."
$MONGOSH --file "$DATA_DIR/create_indexes.js"
log_info "Index creation complete"