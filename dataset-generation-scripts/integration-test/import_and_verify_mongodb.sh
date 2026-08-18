#!/usr/bin/env bash
set -euo pipefail

# ────────────────────────────────────────────────────────────────────
# MongoDB Import & Verification Script
# Imports all 5 collections and runs 17 verification queries.
# Usage: ./integration-test/import_and_verify_mongodb.sh
# Must be run from: dataset-generation-scripts/
# ────────────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="$SCRIPT_DIR/output/mongodb"
GIS_DATA_DIR="$DATA_DIR/gis"

# Load MongoDB credentials from the git-ignored .env.agent file (repo root)
set -a; source "$(git rev-parse --show-toplevel)/.env.agent"; set +a
CONN_STRING="mongodb://${MONGO_ROOT_USER}:${MONGO_ROOT_PASSWORD}@${MONGO_HOST}:${MONGO_PORT}/${MONGO_DB_NAME}?authSource=admin"
DB_NAME="$MONGO_DB_NAME"
run_mongosh() {
  mongosh "$CONN_STRING" --quiet --eval "$1" 2>/dev/null || true
}

run_mongoimport() {
  mongoimport --uri="$CONN_STRING" --jsonArray "$@"
}

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
mongosh "$CONN_STRING" --quiet --eval "db.version()" >/dev/null 2>&1 || {
  log_error "mongosh cannot connect to MongoDB. Check tunnel and credentials."
  exit 1
}
log_info "MongoDB connection OK"

# Verify source files exist
log_info "Verifying source files..."
MANIFEST_FILE=$(ls "$GIS_DATA_DIR"/*ward_gis*.json.manifest 2>/dev/null | head -1)
if [[ -z "$MANIFEST_FILE" ]]; then
  log_error "No ward GIS manifest file found in $GIS_DATA_DIR"
  exit 1
fi

FILES_TO_CHECK=(
  "$DATA_DIR"/administrative_regions_*.json
  "$DATA_DIR"/administrative_units_*.json
  "$DATA_DIR"/mongo_data_vn_unit_*.json
  "$GIS_DATA_DIR"/mongo_data_vn_province_gis_*.json
  "$MANIFEST_FILE"
  "$GIS_DATA_DIR/create_indexes.js"
)

ADMIN_REG_FILE=$(ls "$DATA_DIR"/administrative_regions_*.json 2>/dev/null | head -1)
ADMIN_UNIT_FILE=$(ls "$DATA_DIR"/administrative_units_*.json 2>/dev/null | head -1)
VN_UNIT_FILE=$(ls "$DATA_DIR"/mongo_data_vn_unit_*.json 2>/dev/null | head -1)
PROV_GIS_FILE=$(ls "$GIS_DATA_DIR"/mongo_data_vn_province_gis_*.json 2>/dev/null | head -1)

for file in "$ADMIN_REG_FILE" "$ADMIN_UNIT_FILE" "$VN_UNIT_FILE" "$PROV_GIS_FILE" "$MANIFEST_FILE" "$GIS_DATA_DIR/create_indexes.js"; do
  if [[ ! -f "$file" ]]; then
    log_error "Missing: $file"
    exit 1
  fi
done

# Load ward part files from manifest
WARD_PARTS=()
while IFS= read -r line; do
  [[ -n "$line" ]] && WARD_PARTS+=("$GIS_DATA_DIR/$line")
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
  mongosh "$CONN_STRING" --quiet --eval "db.getCollection('$coll').drop()" >/dev/null 2>&1 || true
done

log_info "Verifying all collections are empty..."
for coll in "${COLLECTIONS[@]}"; do
  COUNT=$(mongosh "$CONN_STRING" --quiet --eval "db.getCollection('$coll').countDocuments()" 2>/dev/null | tail -1)
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
  run_mongoimport --collection="$coll_name" --file="$file" 2>&1 | tail -1

  local actual_count
  actual_count=$(mongosh "$CONN_STRING" --quiet --eval "db.getCollection('$coll_name').countDocuments()" 2>/dev/null | tail -1)
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
  run_mongoimport --collection="wards-gis" --file="${WARD_PARTS[$i]}" 2>&1 | tail -1
done

WARD_COUNT=$(mongosh "$CONN_STRING" --quiet --eval "db.getCollection('wards-gis').countDocuments()" 2>/dev/null | tail -1)
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
mongosh "$CONN_STRING" --quiet --file "$GIS_DATA_DIR/create_indexes.js"
log_info "Index creation complete"

# ──────────────────────────────────────────────────
# Phase 4: Verification
# ──────────────────────────────────────────────────

echo ""
echo "=========================================="
echo " Phase 4: Verification"
echo "=========================================="

verify_count() {
  local check_name="$1"
  local collection="$2"
  local expected="$3"
  local actual
  actual=$(run_mongosh "db.getCollection('$collection').countDocuments()" | tail -1)
  if [[ "$actual" == "$expected" ]]; then
    log_pass "$check_name: $actual (expected $expected)"
    ((TOTAL_PASS++))
  else
    log_fail "$check_name: got $actual, expected $expected"
    log_fail "  Re-run: mongosh '$CONN_STRING' --quiet --eval \"db.getCollection('$collection').countDocuments()\""
    ((TOTAL_FAIL++))
  fi
}

verify_empty() {
  local check_name="$1"
  local eval_expr="$2"
  local actual
  actual=$(run_mongosh "$eval_expr" | tail -1)
  # mongosh may return empty string or "null" for empty results
  if [[ "$actual" == "0" || "$actual" == "null" || -z "$actual" ]]; then
    log_pass "$check_name: no issues found"
    ((TOTAL_PASS++))
  else
    log_fail "$check_name: got $actual, expected 0"
    log_fail "  Re-run: mongosh '$CONN_STRING' --quiet --eval '$eval_expr'"
    ((TOTAL_FAIL++))
  fi
}

verify_equals() {
  local check_name="$1"
  local eval_expr="$2"
  local expected="$3"
  local actual
  actual=$(run_mongosh "$eval_expr" | tail -1)
  if [[ "$actual" == "$expected" ]]; then
    log_pass "$check_name: $actual"
    ((TOTAL_PASS++))
  else
    log_fail "$check_name: got $actual, expected $expected"
    log_fail "  Re-run: mongosh '$CONN_STRING' --quiet --eval '$eval_expr'"
    ((TOTAL_FAIL++))
  fi
}

verify_mongosh_result() {
  local check_name="$1"
  local eval_expr="$2"
  local expected_result="$3"
  local actual
  actual=$(run_mongosh "$eval_expr" | tail -1)
  if [[ "$actual" == "$expected_result" ]]; then
    log_pass "$check_name: $actual"
    ((TOTAL_PASS++))
  else
    log_fail "$check_name: got $actual, expected $expected_result"
    log_fail "  Re-run: mongosh '$CONN_STRING' --quiet --eval '$eval_expr'"
    ((TOTAL_FAIL++))
  fi
}

# 4.1 Document Counts
echo ""
log_info "--- 4.1 Document Counts ---"
verify_count "4.1.1 Province count" "provinces" "34"
verify_count "4.1.2 Province GIS count" "provinces-gis" "34"
verify_count "4.1.3 Ward GIS count" "wards-gis" "3321"
verify_count "4.1.4 Admin regions count" "administrative_regions" "8"
verify_count "4.1.5 Admin units count" "administrative_units" "5"

# 4.2 Data Integrity
echo ""
log_info "--- 4.2 Data Integrity ---"

verify_empty "4.2.1 No duplicate province codes" \
  "JSON.stringify(db.getCollection('provinces-gis').aggregate([{\$group:{_id:'\$Code',count:{\$sum:1}}},{\$match:{count:{\$gt:1}}},{\$count:'dupes'}]).toArray()[0]?.dupes || 0)"

verify_empty "4.2.2 No duplicate ward codes" \
  "JSON.stringify(db.getCollection('wards-gis').aggregate([{\$group:{_id:'\$Code',count:{\$sum:1}}},{\$match:{count:{\$gt:1}}},{\$count:'dupes'}]).toArray()[0]?.dupes || 0)"

verify_equals "4.2.3 Every province-gis has GIS" \
  "db.getCollection('provinces-gis').countDocuments({GIS: {\$exists: false}})" "0"

verify_equals "4.2.4 Every ward-gis has GIS" \
  "db.getCollection('wards-gis').countDocuments({GIS: {\$exists: false}})" "0"

# 4.3 GIS Geometry Validity
echo ""
log_info "--- 4.3 GIS Geometry Validity ---"

verify_equals "4.3.1 Province geometry is MultiPolygon" \
  "db.getCollection('provinces-gis').countDocuments({'GIS.Geometry.type': {\$ne: 'MultiPolygon'}})" "0"

verify_equals "4.3.2 Ward geometry is Polygon/MultiPolygon" \
  "db.getCollection('wards-gis').countDocuments({'GIS.Geometry.type': {\$nin: ['Polygon', 'MultiPolygon']}})" "0"

verify_equals "4.3.3 Province center points exist" \
  "db.getCollection('provinces-gis').countDocuments({'GIS.Center': {\$exists: true}})" "34"

# 4.4 Cross-Collection Referential Integrity
echo ""
log_info "--- 4.4 Cross-Collection Referential Integrity ---"

verify_empty "4.4.1 Every ward refs valid province" \
  "JSON.stringify(db.getCollection('wards-gis').aggregate([{\$lookup:{from:'provinces-gis',localField:'ProvinceCode',foreignField:'Code',as:'province'}},{\$match:{province:{\$size:0}}},{\$count:'orphaned'}]).toArray()[0]?.orphaned || 0)"

verify_empty "4.4.2 Every province has wards" \
  "JSON.stringify(db.getCollection('provinces-gis').aggregate([{\$lookup:{from:'wards-gis',localField:'Code',foreignField:'ProvinceCode',as:'wards'}},{\$match:{wards:{\$size:0}}},{\$count:'empty'}]).toArray()[0]?.empty || 0)"

# 4.5 Spatial Query Samples
echo ""
log_info "--- 4.5 Spatial Query Samples ---"

verify_mongosh_result "4.5.1 Point-in-Hà Nội (105.8542, 21.0285)" \
  "JSON.stringify(db.getCollection('provinces-gis').findOne({'GIS.Geometry':{\$geoIntersects:{\$geometry:{type:'Point',coordinates:[105.8542,21.0285]}}}},{Code:1}).Code)" '"01"'

verify_mongosh_result "4.5.2 Point-in-Ba Đình ward (105.8435, 21.0366)" \
  "JSON.stringify(db.getCollection('wards-gis').findOne({'GIS.Geometry':{\$geoIntersects:{\$geometry:{type:'Point',coordinates:[105.8435,21.0366]}}}},{ProvinceCode:1}).ProvinceCode)" '"01"'

verify_equals "4.5.3 Point-in-ocean matches nothing" \
  "db.getCollection('provinces-gis').countDocuments({'GIS.Geometry':{\$geoIntersects:{\$geometry:{type:'Point',coordinates:[112.0,16.0]}}}})" "0"

# ──────────────────────────────────────────────────
# Phase 5: Report
# ──────────────────────────────────────────────────

echo ""
echo "=========================================="
echo " Phase 5: Report"
echo "=========================================="
echo ""
echo "Results: ${GREEN}${TOTAL_PASS} passed${NC}, ${RED}${TOTAL_FAIL} failed${NC}"
echo ""

if [[ "$TOTAL_FAIL" -gt 0 ]]; then
  log_error "Verification FAILED — ${TOTAL_FAIL} check(s) did not pass"
  exit 1
else
  log_pass "All ${TOTAL_PASS} verification checks passed"
  exit 0
fi
