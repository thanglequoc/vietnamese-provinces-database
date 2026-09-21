# Release & Generation Guide (Maintainers)

> This document is for **maintainers** who need to cut a new dataset release.
> If you are here to *use* the dataset, see the [root README](../README.md).

This is the **only supported release route**. It is:

```
1. Produce the dataset update      (decree workflow, or manual local run)
2. Review & merge the dataset PR
3. Create the upgrade patch        (patch/<version>/)
4. Write the release notes         (docs/release_notes/<version>.md)
5. Run "Publish Dataset Archives"  (upload to R2 + open README download-table PR)
6. Merge the README PR             (download links now point at <version>)
7. Tag the release                 ("Tag Release"; version.txt must match)
8. Verify
```

> The tag is deliberately created **last**, after the download links are merged.
> That way `git show vX.Y.Z:README.md` — and the GitHub release page — point at
> the archives for that same version, not the previous release's links.

---

## 1. Prerequisites

### 1.1 Repository secrets and variables

Configure under **GitHub → Settings → Secrets and variables → Actions**.

**Secrets** (never printed, write-only):

| Name | Purpose |
|------|---------|
| `R2_ACCESS_KEY_ID` | Cloudflare R2 API token — Access Key ID |
| `R2_SECRET_ACCESS_KEY` | Cloudflare R2 API token — Secret Access Key |
| `DECREE_AUTOMATION_PAT` | PAT used by the automated PRs so they trigger `Test Go Code` (classic: `repo` + `workflow`; fine-grained: Contents RW + Pull requests RW) |

**Variables** (non-sensitive):

| Name | Example / value | Purpose |
|------|-----------------|---------|
| `R2_ACCOUNT_ID` | `380b8416d2155bd5d1531460190d3533` | Builds the S3 endpoint `https://<id>.r2.cloudflarestorage.com` |
| `R2_BUCKET_NAME` | `vietnamese-provinces-database` | Target bucket |
| `R2_PUBLIC_BASE_URL` | `https://vn-provinces-ds.thanglequoc.xyz` | Public CDN base used in README links and CDN verification |

### 1.2 Cloudflare R2 setup

1. **Bucket** — create an R2 bucket named `vietnamese-provinces-database`.
2. **API token** — Cloudflare dashboard → **R2 → API → Manage API Tokens → Create API Token**, permission **Object Read & Write**, scoped to that bucket. Copy the **Access Key ID** and **Secret Access Key** into `R2_ACCESS_KEY_ID` / `R2_SECRET_ACCESS_KEY`.
3. **Public custom domain** — bucket → **Settings → Custom Domains → Connect Domain**, add `vn-provinces-ds.thanglequoc.xyz`. Objects are then served at `https://vn-provinces-ds.thanglequoc.xyz/<key>`.
4. **Account ID** — R2 → Overview (right sidebar), or the Cloudflare account home.

> R2 requires `AWS_DEFAULT_REGION=auto` and the S3-compatible endpoint. The workflows already set both.

### 1.3 Local tooling (only needed for the manual generation route)

- **Go 1.26+** (matches `dataset-generation-scripts/go.mod`)
- **Docker** (temporary Postgres/PostGIS container)
- `psql`/`aws`/`gh` are optional conveniences

```bash
cd dataset-generation-scripts
docker compose -f docker/docker-compose.yaml up -d   # PostGIS on localhost:15432
cp .env.example .env
```

The `.env` must match the bundled Docker container:

```dotenv
POSTGRES_DB_USERNAME=postgres
POSTGRES_DB_PSWD=root
POSTGRES_DB_HOST=localhost
POSTGRES_DB_PORT=15432
POSTGRES_TMP_DB_NAME=vn_provinces_tmp
```

---

## 2. Repository configuration reference

| Path | Role |
|------|------|
| `dataset-generation-scripts/version.txt` | **Source of truth**: `dataset_version` + `latest_decree`. Must match the git tag at release time. |
| `dataset-generation-scripts/main.go` | Entry point. `INCLUDE_GIS` const (`true` by default) gates the GIS pipeline. |
| `dataset-generation-scripts/copy-datasets-to-repo.sh` | Copies `output/` into the published folders (`json/`, `postgresql/`, …). |
| `.github/workflows/new-decree-data-patch.yml` | **New Decree Data Patch** — scheduled decree detection + regeneration, opens the dataset PR. |
| `.github/workflows/publish-dataset-archives.yml` | **Publish Dataset Archives** — manual; packages `master` + uploads to R2 and opens the README download-table PR (run *before* tagging). |
| `.github/workflows/tag-release.yml` | **Tag Release** — manual; creates the release tag + GitHub release from `version.txt` (run *after* the download PR merges). |
| `.github/workflows/test-go.yml` | **Test Go Code** — runs on PRs. |
| `.github/scripts/package-datasets.sh` | Zips each dataset folder + writes `downloads.json`. |
| `.github/scripts/upload-archives-to-r2.sh` | Uploads archives + manifest to R2, verifies CDN URLs. |
| `.github/scripts/update_download_tables.py` | Regenerates the root README download tables from `downloads.json`. |
| `.opencode/skills/vn-provinces-patch` | Generates the data-only SQL upgrade patch. |
| `.opencode/skills/vn-provinces-release-notes` | Generates the bilingual release note + README version tables. |

### R2 key layout

```
<version>/<format>/vn_provinces_<format>_dataset_<version>.zip
<version>/downloads.json
```

- `<version>` is `dataset_version` from `version.txt` (e.g. `v5.2.0`).
- `<format>` is the lowercase dataset folder id: `postgresql`, `mysql`, `sqlserver`,
  `oracle`, `json`, `mongodb`, `redis`, `elasticsearch`.

Example:
`https://vn-provinces-ds.thanglequoc.xyz/v5.2.0/postgresql/vn_provinces_postgresql_dataset_v5.2.0.zip`

### README download tables

The download tables are generated between markers and must not be hand-edited:

```
<!-- DOWNLOAD_TABLE:START -->   … <!-- DOWNLOAD_TABLE:END -->        (README.md, README_vi.md)
```

---

## 3. Step-by-step release process

### Step 1 — Produce the dataset update

Pick **one** of the two routes.

#### Route A — Automated (new government decree)

The **New Decree Data Patch** workflow runs daily at 08:00 ICT. When a newly
effective decree differs from `version.txt`, it:

1. minor-bumps `dataset_version` (e.g. `v5.1.0 → v5.2.0`) and sets `latest_decree`;
2. regenerates every format (`go run main.go`) and copies them into the repo;
3. opens a PR `auto/decree-<decree-slug>` against `master`.

Manual trigger (e.g. to force a regeneration):

```bash
# Force generation even when no new decree is detected.
# WARNING: --apply always bumps the minor version, so force=true creates a new version.
gh workflow run "New Decree Data Patch" --ref master -f force=true

# Rehearse without opening a PR (generation only; no PR).
gh workflow run "New Decree Data Patch" --ref master -f force=true -f dry_run=true
```

#### Route B — Manual (data/code correction, no new decree)

```bash
cd dataset-generation-scripts
docker compose -f docker/docker-compose.yaml up -d

# Choose the new version (skip if you are keeping the current one):
#   edit version.txt → dataset_version=v5.2.1, latest_decree=<decree>

go run main.go                  # set INCLUDE_GIS=false in main.go for a fast admin-only run
./copy-datasets-to-repo.sh

# Verify the temporary DB
docker exec vn_provinces_postgres_container psql -U postgres -d vn_provinces_tmp \
  -c "SELECT (SELECT COUNT(*) FROM provinces_tmp) AS provinces, (SELECT COUNT(*) FROM wards_tmp) AS wards;"

go test -v ./...                # Docker must be running

# Commit from the repository root
cd ..
git add json postgresql mysql sqlserver oracle mongodb redis elasticsearch dataset-generation-scripts/version.txt
git commit -m "feat: regenerate dataset v5.2.1"
git push
```

### Step 2 — Review and merge the dataset PR

Check the diff and the counts (currently **34 provinces / 3,321 wards**, and
**5** `administrative_units`). Merge to `master`.

### Step 3 — Create the upgrade patch

Use the `vn-provinces-patch` skill with the **previous release tag** as the baseline
(e.g. `v5.1.0`). It writes a data-only PostgreSQL patch under `patch/<version>/` and
verifies it in the Docker container. Commit it.

### Step 4 — Write the release notes

Use the `vn-provinces-release-notes` skill against the previous tag. It writes
`docs/release_notes/<version>.md`, updates the index, and adds the version row to the
release tables in `README.md` / `README_vi.md`. Commit it.

> Steps 3 and 4 are independent of the R2 publish; both can go in a small PR or a
> direct commit.

### Step 5 — Publish the archives

**Actions → Publish Dataset Archives → Run workflow** (branch `master`).

> The tag does **not** exist yet at this point — that is intentional. The archives
> are packaged from the merged `master` data, and the tag is created in Step 7 so
> that it includes the updated download links.

The workflow:
1. reads `dataset_version` from `version.txt` (used for the R2 keys, the manifest
   and the PR branch);
2. checks out `master` and zips each dataset folder → `downloads.json`
   (recording the source commit);
3. uploads to R2 at `<version>/<format>/…` and verifies the CDN returns HTTP 200;
4. opens a PR `auto/downloads-<version>` that refreshes the README download tables.

CLI equivalent:

```bash
gh workflow run "Publish Dataset Archives" --ref master
gh run watch "$(gh run list --workflow 'Publish Dataset Archives' --limit 1 --json databaseId -q '.[0].databaseId')"
```

### Step 6 — Merge the README PR

Merge `auto/downloads-<version>`. This makes the published links live in the root
READMEs. Re-running the workflow for the same version is idempotent (no PR when the
tables already match).

Confirm the links are live before tagging:

```bash
grep -c 'dataset_v5.2.0.zip' README.md README_vi.md   # expect a non-zero count
```

### Step 7 — Tag the release

**Actions → Tag Release → Run workflow** (branch `master`).

The workflow:
1. reads `dataset_version` from `version.txt` and uses it as the tag name;
2. fails if the READMEs do not yet link to the published archives, or if a tag
   with the same name already exists at a *different* commit;
3. creates and pushes the tag on the current `master` HEAD (safe to re-run — an
   existing tag at the same commit is reused);
4. creates the GitHub release, using `docs/release_notes/<version>.md` as the body
   when present, and GitHub-generated notes otherwise.

> The release-note file is **not** required to tag; it only provides a curated
> release body.

CLI equivalent (the workflow is a thin wrapper around this):

```bash
# confirm the version
grep '^dataset_version=' dataset-generation-scripts/version.txt

git tag v5.2.0
git push origin v5.2.0

gh release create v5.2.0 \
  --title "<decree description or release summary>" \
  --notes-file docs/release_notes/v5.2.0.md
```

Because the tag is created after Step 6, `git show v5.2.0:README.md` points at the
v5.2.0 archives.

### Step 8 — Verify

```bash
V=v5.2.0
for f in postgresql mysql sqlserver oracle json mongodb redis elasticsearch; do
  curl -s -o /dev/null -w "%{http_code}  $f\n" \
    "https://vn-provinces-ds.thanglequoc.xyz/$V/$f/vn_provinces_${f}_dataset_$V.zip"
done
curl -s -o /dev/null -w "manifest %{http_code}\n" \
  "https://vn-provinces-ds.thanglequoc.xyz/$V/downloads.json"
```

Expect HTTP `200` for all eight archives and the manifest.

Optional deeper check — download and import the PostgreSQL archive into the local
PostGIS container (see [§4](#4-verifying-a-published-archive)).

---

## 4. Verifying a published archive

```bash
TMP=$(mktemp -d) && cd "$TMP"
curl -sSLO "https://vn-provinces-ds.thanglequoc.xyz/v5.2.0/postgresql/vn_provinces_postgresql_dataset_v5.2.0.zip"
unzip -q vn_provinces_postgresql_dataset_v5.2.0.zip

C=vn_provinces_postgres_container
docker exec "$C" psql -U postgres -q -c "DROP DATABASE IF EXISTS vn_provinces_verify;"
docker exec "$C" psql -U postgres -q -c "CREATE DATABASE vn_provinces_verify;"
docker exec -i "$C" psql -U postgres -d vn_provinces_verify -q -v ON_ERROR_STOP=1 \
  < postgresql/postgres_CreateTables_vn_units.sql
docker exec -i "$C" psql -U postgres -d vn_provinces_verify -q -v ON_ERROR_STOP=1 \
  < postgresql/postgres_ImportData_vn_units.sql

docker exec "$C" psql -U postgres -d vn_provinces_verify -c \
  "SELECT (SELECT COUNT(*) FROM provinces) AS provinces, (SELECT COUNT(*) FROM wards) AS wards;"

# GIS add-on (optional)
docker exec "$C" psql -U postgres -d vn_provinces_verify -q -c "CREATE EXTENSION IF NOT EXISTS postgis;"
docker exec -i "$C" psql -U postgres -d vn_provinces_verify -q -v ON_ERROR_STOP=1 \
  < postgresql/gis/postgresql_CreateGISTables.sql
while read -r p; do docker exec -i "$C" psql -U postgres -d vn_provinces_verify -q -v ON_ERROR_STOP=1 < "postgresql/gis/$p"; done \
  < postgresql/gis/postgresql_ImportData_gis.sql.manifest
```

Expected: **34** provinces, **3,321** wards, **34** `gis_provinces`, **3,321** `gis_wards`.

---

## 5. Version numbering

| Trigger | Rule | Example |
|---------|------|---------|
| Decree (automated) | minor bump, patch reset to 0 | `v5.1.0 → v5.2.0` |
| Manual correction | maintainer's choice (patch bump typical) | `v5.2.0 → v5.2.1` |
| Git tag | must equal `version.txt` `dataset_version` | `v5.2.0` |

---

## 6. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| Publish workflow packages the wrong data | `master` moved / `version.txt` bumped without merging the dataset | Check the run summary's `master@<sha>` and re-run after fixing `master` |
| "could not read dataset_version from version.txt" | `version.txt` malformed | Ensure a line `dataset_version=vX.Y.Z` |
| Tag Release warns "docs/release_notes/vX.Y.Z.md not found" | Release note file absent | Non-blocking — GitHub generates the body. Commit the note (Step 4) to use a curated body |
| Tag Release fails at "README.md does not reference the vX.Y.Z archives" | The `auto/downloads-<version>` PR is not merged | Merge the downloads PR (Step 6), then re-run |
| Tag Release fails at "tag vX.Y.Z already exists at <sha>, not at HEAD" | A tag for that version exists at another commit | Bump `version.txt`, or delete the misplaced tag |
| CDN URL returns non-200 after upload | Custom domain not connected / not propagated | Check R2 custom domain; re-run the workflow |
| `connect: connection refused` on 15432 | Docker not running | `docker compose -f docker/docker-compose.yaml up -d` |
| Decree check "markup may have changed" | GSO page markup changed | Update `internal/decree_check` patterns + testdata |
| R2 upload fails with region error | `AWS_DEFAULT_REGION` not `auto` | The workflow sets it; check secrets/variables |

### Cleaning up a bad publish

```bash
# Fill these in from the repository variables/secrets (or the Cloudflare dashboard)
export R2_ACCOUNT_ID="<account id>"
export R2_BUCKET_NAME="vietnamese-provinces-database"
export AWS_ACCESS_KEY_ID="<R2 access key id>"
export AWS_SECRET_ACCESS_KEY="<R2 secret access key>"
export AWS_DEFAULT_REGION=auto
export AWS_ENDPOINT_URL_S3="https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"

aws s3 rm --recursive "s3://${R2_BUCKET_NAME}/<version>/"
```

---

## 7. Quick reference

```bash
# Generate locally
cd dataset-generation-scripts && go run main.go && ./copy-datasets-to-repo.sh

# Check for a new decree (offline sample)
go run ./cmd/decreecheck --html-file ./internal/decree_check/testdata/nghidinh_sample.html --today 2026-09-21

# Publish archives (run before tagging)
gh workflow run "Publish Dataset Archives" --ref master

# Tag a release (after the downloads PR merges)
gh workflow run "Tag Release" --ref master
```
