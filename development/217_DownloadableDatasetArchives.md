# 217 — Downloadable Dataset Archives via Cloudflare R2

## Objectives

Make each published dataset downloadable without cloning the whole repository:

1. Add a manually-triggered GitHub Actions workflow that resolves the release
   version from `dataset-generation-scripts/version.txt`, compresses each
   top-level dataset folder into a ZIP archive, and uploads it to the
   maintainer's Cloudflare R2 bucket.
2. Publish the resulting CDN URLs in a download table in `README.md` /
   `README_vi.md`.
3. Keep the repository's existing generation/patch workflows untouched.

## Decisions

| Topic | Decision |
|-------|----------|
| Archive granularity | One ZIP per top-level format folder, as-is (8 archives: `json`, `postgresql`, `mysql`, `sqlserver`, `oracle`, `mongodb`, `redis`, `elasticsearch`). GIS content stays inside each engine folder. Postal-code data is embedded in the base datasets, so no separate archive. |
| README write-back | Open a PR against `master` (`peter-evans/create-pull-request`), matching `new-decree-data-patch.yml`. |
| Docs scope | Bilingual Download section in the root READMEs (marker block). |
| Upload tool | `aws-cli` v2 against the R2 S3 endpoint (`AWS_ENDPOINT_URL_S3`, `AWS_DEFAULT_REGION=auto`). |
| Credentials | **Option A** — endpoint derived from `R2_ACCOUNT_ID`. Secrets: `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`. Variables: `R2_ACCOUNT_ID`, `R2_BUCKET_NAME`, `R2_PUBLIC_BASE_URL`. |
| Version resolution | `dataset_version` from `dataset-generation-scripts/version.txt` (must match the git tag); optional manual `version` input overrides. |
| Manual only | `workflow_dispatch` only; no tag-push trigger (future enhancement). |
| Decree chaining | Not chained. Archive publishing runs as a separate manual step after the dataset PR is merged and the release is tagged. |

## Bucket conventions

- Key: `<version>/<format>/<archive>`
- `<format>`: lowercase dataset folder id — `postgresql`, `mysql`, `sqlserver`,
  `oracle`, `json`, `mongodb`, `redis`, `elasticsearch`
- Archive: `vn_provinces_<format>_dataset_<version>.zip`
- Example:
  `https://vn-provinces-ds.thanglequoc.xyz/v5.1.0/postgresql/vn_provinces_postgresql_dataset_v5.1.0.zip`
- ZIP built from repo root (extracts to a named folder), `zip -r -X -9`,
  excluding `.DS_Store` / `__MACOSX` / `Thumbs.db`.
- Bucket: `vietnamese-provinces-database` (verified reachable via the
  `vn-provinces-ds.thanglequoc.xyz` custom domain).

## Components

### 1. `.github/workflows/publish-dataset-archives.yml`

- `workflow_dispatch` inputs: `version` (optional), `dry_run` (bool).
- `permissions: contents: write, pull-requests: write`;
  `concurrency: publish-dataset-archives`.
- **Job `package`**
  1. Resolve version (input or latest release).
  2. Checkout the workflow ref into `repo/` (sparse: `.github`) for the scripts.
  3. Checkout the resolved tag into `data/` (sparse: the 8 dataset folders).
  4. Run `package-datasets.sh <version> <out> --repo-root data`.
  5. Upload each archive + `<version>/downloads.json` to R2 (`aws s3 cp`).
  6. Verify each CDN URL with `curl` (warn only) and write the run summary.
  7. Upload `downloads.json` as the `downloads-manifest` artifact.
- **Job `docs`** (`needs: package`, skipped when `dry_run`)
  1. Checkout `master` (sparse: `README.md`, `README_vi.md`).
  2. Download the `downloads-manifest` artifact.
  3. Run `update_download_tables.py`.
  4. `peter-evans/create-pull-request@v8` → branch `auto/downloads-<version>`.

### 2. `.github/scripts/package-datasets.sh`

`package-datasets.sh <version> <output-dir> [--repo-root DIR] [--only ID,ID] [--dry-run]`

Zips each dataset folder, computes size + SHA-256, and writes `downloads.json`:

```json
{
  "version": "v5.1.0",
  "generated_at": "2026-09-18T00:00:00Z",
  "base_url": "https://vn-provinces-ds.thanglequoc.xyz",
  "datasets": [
    { "id": "postgresql",
      "archive": "vn_provinces_postgresql_dataset_v5.1.0.zip",
      "size_bytes": 1, "size_human": "1.00 MB", "sha256": "...", "url": "..." }
  ]
}
```

### 3. `.github/scripts/update_download_tables.py`

Reads `downloads.json` and regenerates the dataset table between
`<!-- DOWNLOAD_TABLE:START/END -->` in `README.md` (EN) and `README_vi.md` (VI).

## Edge cases

- Missing `dataset_version` in `version.txt`, or the tag not existing → clear failure.
- Re-run same version → R2 PUT overwrites; PR no-ops if README unchanged.
- `.DS_Store` / `__MACOSX` excluded; `zip -X` for cross-platform extraction.
- CDN URL not 200 after upload → warn, do not fail the run.
- Largest archive is well under R2 single-PUT limits; `aws s3 cp` auto-multiparts.

## Verification

1. Local: `package-datasets.sh v5.1.0 /tmp/archives` then
   `update_download_tables.py` on a scratch copy.
2. After merge: run the workflow and confirm each URL returns 200 with matching
   `Content-Length`.
3. Confirm the docs PR updates both READMEs; re-run is idempotent.

## Out of scope

- Auto-publishing on tag push (manual trigger only for now).
- A dedicated combined `GISDataSet` archive (GIS ships inside each format).
- Migrating historical versions (`v3.2.0`, `v4.0.0`, `v4.1.0`) to ZIP archives.
