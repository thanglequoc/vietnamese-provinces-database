# Design: `134_AddPostalCode_v2` — squashed code carryover

Date: 2026-08-18

## Objective

Create a mergeable branch `134_AddPostalCode_v2` from `master` containing exactly the
`dataset-generation-scripts/` changes from `134_AddPostalCode` (plus the security fix and
relevant docs), in a single commit, so GitHub can merge it back to `master`.

## Background

- `master` is at `cd580632`; branch `134_AddPostalCode` is 75 commits ahead.
- GitHub cannot merge `134_AddPostalCode` into `master` because the diff is too large:
  **3,490 files changed**, the overwhelming majority being regenerated dataset exports
  (`json/` alone: 3,363 files).
- 49 of the 75 commits touch `dataset-generation-scripts/`, yielding a net change of
  **44 files** in that folder (the postal code feature + dataset writer improvements).
  No deletions, no renames in the folder.
- The merge-base is `master` HEAD itself, so there are no real conflicts — the problem is
  purely diff size.

## Approach chosen (Option B — squash)

Carry over the exact file states from `origin/134_AddPostalCode` for the selected paths,
as a **single commit** on a new branch. No commit history is preserved; the final folder
state is identical to the source branch.

## Scope

### Included

| Path | Reason |
|------|--------|
| `dataset-generation-scripts/` (whole folder, 44 files) | The actual code change |
| `.env.agent.example`, `.gitignore`, `AGENTS.md` | Security fix (remove password) + Meta doc updates |
| `development/134_AddPostalCode/` | Feature docs/plans/seeds tooling |
| `development/chunk-gis-sql-output*` | Related design/plan for chunked GIS SQL output |
| `docs/gis/` | `gis_readme.md`, `gis_readme_vi.md` updates |

### Excluded (regenerated later by user)

`json/`, `elasticsearch/` (mappings + ndjson), `mongodb/`, `mysql/`, `oracle/`,
`postgresql/`, `sqlserver/`, `redis/`, and the 7 unrelated old-plan deletions under
`development/` (from an unrelated cleanup commit).

## Execution steps

1. **Pre-flight**: confirm working tree is clean. Untracked files (`docs/release_notes_*.md`,
   `.env.agent`) are left alone.
2. **Create branch**: `git checkout -b 134_AddPostalCode_v2` (from current `master`).
3. **Copy exact file states** from `origin/134_AddPostalCode`:
   ```
   git checkout origin/134_AddPostalCode -- \
     dataset-generation-scripts/ \
     .env.agent.example \
     .gitignore \
     AGENTS.md \
     development/134_AddPostalCode/ \
     'development/chunk-gis-sql-output*' \
     docs/gis/
   ```
4. **Verify parity**: `git diff origin/134_AddPostalCode -- dataset-generation-scripts/`
   must be empty (folder identical to source branch).
5. **Commit once**: e.g. `feat: add postal code support to dataset generation scripts (134_AddPostalCode_v2)`.
6. **Final checks**: `git status` shows only intended files; optionally run `go build ./...`
   in `dataset-generation-scripts/`.

## Risks

- `go build` may fail if code depends on `output/` artifacts or `.env` at compile time.
  We verify and report; we do not silently fix out-of-scope issues.

## Success criteria

- New branch `134_AddPostalCode_v2` exists off `master` with exactly the included paths
  at source-branch content and **one** commit.
- `git diff origin/134_AddPostalCode -- dataset-generation-scripts/` is empty.
- GitHub can open/merge the PR cleanly (small diff).
