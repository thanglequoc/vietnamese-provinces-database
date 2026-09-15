# 180 — Automated Workflow to Detect New Government Decree

## Objectives

Automate the "new decree → regenerate dataset → open PR" loop:

1. On a schedule, read the GSO administrative-unit decree table at
   `https://danhmuchanhchinh.nso.gov.vn/NghiDinh.aspx`.
2. Determine the newest decree whose **effective date is on or before today**
   (Asia/Ho_Chi_Minh).
3. Compare it with the dataset's recorded `latest_decree` in
   `dataset-generation-scripts/version.txt`.
4. If it differs, regenerate the full dataset (`go run main.go`), copy the
   outputs into the published folders (`copy-datasets-to-repo.sh`), and open (or
   update) a Pull Request against `master`.

## Decisions

| Topic | Decision |
|-------|----------|
| Decree selection | Scan rows top-down; pick the first row whose effective date ≤ today (ICT). The GSO list is newest-*published*-first, and future-effective decrees can sit at the top, so the gate skips them. |
| Version bump | Auto minor-bump `dataset_version` (middle digit, `v5.1.0 → v5.2.0`) and set `latest_decree` to the detected decree. |
| CI generation scope | Full `main.go` (GIS included) + `copy-datasets-to-repo.sh`. |
| PR tooling | `peter-evans/create-pull-request` with a PAT secret so `test-go.yml` runs on the auto-PR. |
| Schedule | Mon/Wed/Fri 08:00 ICT and Sat 22:00 ICT → `0 1 * * 1,3,5` + `0 15 * * 6` (UTC). |

## Components

### 1. `internal/decree_check` (Go package)

- **Fetch** — HTTP GET of the GSO page with a browser-like User-Agent and a
  30s timeout.
- **Parse** — regex extraction of the DevExpress grid rows
  (`...DXDataRowN`) and their four `dxgv` cells:
  *Nghị Định Số*, *Ngày Ban Hành* (issue date), *Ngày Hiệu Lực* (effective
  date, `dd/MM/yyyy`), *Nội Dung*. Cells are HTML-unescaped and whitespace
  collapsed (some decree numbers have trailing spaces).
- **Select** — first row (document order) with `effectiveDate ≤ today`.
- **Compare** — normalized equality against `version.txt` `latest_decree`.
  A secondary ordering guard prevents "downgrading" when the detected decree
  appears lower in the table than the current one (i.e. older).
- **Version bump** — parse `vX.Y.Z`, increment the minor (middle) component and
  reset the patch component to `0`, then rewrite `version.txt`
  preserving header comments and key order.

### 2. `cmd/decreecheck` (CLI)

- `--check` (default): prints GitHub Actions `key=value` outputs.
- `--apply --decree <n>`: bumps `version.txt`.
- Testability flags: `--html-file`, `--today`, `--url`, `--version-file`.

### 3. `.github/workflows/new-decree-data-patch.yml`

- **Job A `detect`** — no DB, no secrets. Runs `decreecheck`, exposes
  `should_generate`, `decree`, `decree_slug`, `effective_date`,
  `current_decree`, `current_version`. Fails loudly on fetch/parse errors.
- **Job B `generate`** — `needs: detect`, runs only when
  `should_generate == 'true'`. PostGIS service on port 15432, writes `.env`,
  applies the version bump, runs `main.go`, runs
  `copy-datasets-to-repo.sh`, uploads generation logs on failure, then opens
  the PR.
- `workflow_dispatch` inputs `force` and `dry_run` support manual testing.

## Data flow

```
schedule / workflow_dispatch
        │
        ▼
Job A: decreecheck ── should_generate? ──no──► stop
        │ yes
        ▼
Job B: version.txt bump ─► main.go ─► copy-datasets-to-repo.sh ─► create-pull-request
```

## Edge cases

- **Future-effective top decree** — skipped by the effective-date gate.
- **Current decree already effective and equal** — `should_generate=false`.
- **Trailing whitespace / HTML entities in cells** — normalized before compare.
- **GSO markup change** — parser returns an error; Job A fails loudly.
- **Existing open PR for the same decree** — `create-pull-request` updates the
  same branch/PR instead of creating a duplicate.
- **No diff after regeneration** — `create-pull-request` no-ops.
- **External network failure in Job B** (DVHCVN SOAP / sapnhap GIS) — no PR;
  `output/generation-log.txt` is uploaded as an artifact.

## Assumptions

- "Run the copy over shell script" = `dataset-generation-scripts/copy-datasets-to-repo.sh`
  (no remote database is updated by this workflow).
- The GSO table's newest already-effective decree is the correct release marker.
- `DECREE_AUTOMATION_PAT` repository secret exists with `repo`+`workflow`
  (classic) or Contents RW + Pull requests RW (fine-grained).

## Out of scope

Release-note generation, README version-table updates, patch-folder
generation, and the pre-existing dead reference to
`resources/gis/sapnhap_bando_tables.sql` in `test-go.yml`.
