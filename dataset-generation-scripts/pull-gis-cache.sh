#!/usr/bin/env bash
set -euo pipefail

# One-time pull of the government GIS responses into the local mock cache.
#
# The captured files are large (~63 MB gzipped, ~6,700 files), so
# resources/gis/gis_server_cache/ is gitignored. Run this script whenever the
# upstream GIS data changes, then serve it with cmd/mockgis.
#
# Usage:
#   ./pull-gis-cache.sh                 # use the committed malk list (no DB needed)
#   ./pull-gis-cache.sh --source db     # use sapnhap_geojson_objects (authoritative; needs the PostGIS container)
#   ./pull-gis-cache.sh --workers 20    # any extra flags are forwarded to cmd/giscapture
#
# Then run the generator offline:
#   go run ./cmd/mockgis --data ./resources/gis/gis_server_cache --addr 127.0.0.1:18080
#   GIS_SERVER_BASE_URL=http://127.0.0.1:18080 go run main.go

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Defaults first so any user-supplied flag (e.g. --source db) overrides them,
# since Go's flag package takes the last occurrence.
exec go run ./cmd/giscapture --source json --force "$@"
