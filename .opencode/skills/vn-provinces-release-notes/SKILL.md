---
name: vn-provinces-release-notes
description: Generate a bilingual (Vietnamese + English) end-user release note by diffing the current repository state (HEAD/working tree) against the most recent release tag (or a user-provided git tag/commit), focusing only on changes beneficial to end users and excluding the internal dataset-generation-scripts folder, then store it under docs/release_notes/<version>.md and update the docs/release_notes/README.md index. Use when asked to "generate/draft/write release notes", "create a release note for the next version", "what changed since <tag>", "summarize changes for users", or when preparing a new release.
---

# VN Provinces Release Notes

Generate an **end-user-focused, bilingual (Vietnamese + English)** release note by comparing
the current state of the repo against the **most recent release tag** (or a user-specified
tag/commit). The note documents only changes that benefit consumers of the published
datasets. Internal generator work under `dataset-generation-scripts/` is **excluded**.

## Scope

- **Covered (end-user beneficial)**: published dataset formats — `postgresql/`, `mysql/`,
  `sqlserver/`, `oracle/`, `mongodb/`, `redis/`, `json/`, `elasticsearch/`; root
  `README.md` / `README_vi.md`; `docs/gis/` user documentation; new user-facing artifacts
  (e.g. `*.zip`, `*.ndjson`, mapping files).
- **Excluded (internal only)**: `dataset-generation-scripts/`, `.opencode/`, `.github/`,
  `.claude/`, `development/`, `docs/superpowers/`, `opencode.json`, `.env.agent*`,
  `.gitignore`, `AGENTS.md`, `patch/`, `output/`, and the release notes themselves
  (`docs/release_notes*/`).
- **Language**: bilingual — Vietnamese sections first, then an `**English version**:`
  block, matching `docs/release_notes_v4.1.0.md` / `docs/release_notes_v4.2.0.md`.
- **Output location**: `docs/release_notes/<release_version>.md` (new naming convention;
  legacy notes are flat files `docs/release_notes_vX.Y.Z.md`).

## Inputs & baseline resolution

| Role | Source |
|------|--------|
| **NEW** (current) | `HEAD` / working tree |
| **OLD** (baseline) | resolved in priority order below |

OLD baseline resolution:

1. **User-specified ref** — if the user summons the skill with a git tag or commit hash
   (e.g. `v4.2.0`, `abc1234`), treat that ref as the OLD version.
2. **Default** — no ref given: OLD = most recent release tag:
   `git tag --sort=-version:refname | head -1` (e.g. `v4.2.0`).

**Important**: if the resolved OLD baseline equals the current state (no user-facing diff),
tell the user there are no end-user changes to report and stop.

## Workflow

### Step 1 — Verify inputs & resolve the baseline

- Confirm the repo is a git worktree and the baseline ref exists.
- Resolve OLD per the rules above. Record it (needed for the classifier and verification).
- Run `git log <base>..HEAD --oneline | head -30` to skim the change narrative early.

### Step 2 — Run the classifier

From the repo root:

```bash
python3 .opencode/skills/vn-provinces-release-notes/scripts/classify_changes.py --base <BASE_REF>
```

Use `--head <ref>` to compare against something other than `HEAD`. The script emits a
categorized Markdown summary:

- **Schema changes (new columns)** — columns added to the `*_CreateTables*` files
- **Dataset content regenerated** — per-format counts of `*_ImportData*` / content files
- **New artifacts / formats** — newly added files, dense dirs aggregated
- **Documentation changes** — README / `docs/gis/` updates
- **Other user-facing changes** — remaining dataset file modifications (geojson, gis, zip, …)
- **Commit narrative** — subjects to review for the story, filtering internal-only commits
- **Internal-only changes** — a reminder of what must be excluded

### Step 3 — Turn the classification into a release story

Use the classifier output + `git log <base>..HEAD --oneline` to identify concrete
**user-facing** themes. Good source signals:

- **New formats / collections** — new top-level dirs or artifacts (e.g. Elasticsearch,
  MongoDB GIS, new JSON variants, re-partitioned GIS import files).
- **New data / fields** — schema additions (e.g. postal codes) and content-file
  regenerations across all formats.
- **Data corrections** — commit subjects mentioning fixes (e.g. ward code corrections,
  name patches, geometry validation). Confirm the exact affected units with targeted diffs:
  `git diff <base>..HEAD -- postgresql/postgres_ImportData_vn_units.sql | head -100`.
- **Documentation** — README rewrites, new query guides.

For each theme, confirm the specifics against the diff before writing. Do NOT invent or
infer details that are not visible in the changes.

### Step 4 — Ask for the release version

Ask the user: **"Which release version should this release note target?"** (e.g. `v4.3.0`).

Target path: `docs/release_notes/<release_version>.md`.

### Step 5 — Write the bilingual release note

Follow the style of `docs/release_notes_v4.1.0.md` / `docs/release_notes_v4.2.0.md`:

**Structure:**
1. Opening summary paragraph (1–2 sentences, Vietnamese) describing the release highlights.
2. Sections, ordered by importance, each with an emoji header and a category tag:
   - `## ✨ <Title> (Mới / New)` — new datasets, formats, collections
   - `## 🍃 <Title> (New)` / `## 🔍 <Title>` — format-specific additions
   - `## 🏝️ <Title> (Sửa lỗi / Fix)` — data corrections
   - `## 🔧 <Title> (Cải thiện / Improvement)` — quality/consistency improvements
   - `## 📚 Cập nhật tài liệu / Documentation & Cleanup`
3. Within a section: intro sentence → optional data table → **bold key-feature bullets**
   → link to the relevant format README (`[`name`/README.md](../name/README.md)`).
4. `---` separators between sections.
5. After all Vietnamese sections, `**English version**:` followed by the same content in
   English (mirror the exact structure).

**Rules:**
- Only end-user beneficial changes. Exclude everything under `dataset-generation-scripts/`
  and internal tooling.
- Include data tables when listing formats, collections, indices, or patched units
  (matching the existing examples).
- Keep numbers consistent with the generated data (e.g. 34 provinces, 3,321 wards).
- The release note is Vietnamese-first; the English block is a full mirror.

### Step 6 — Update the release notes index

Append the new version to `docs/release_notes/README.md`: version, date, one-line summary,
and a link to the note file.

### Step 7 — Verify

- Every claimed item must exist in the diff (`git diff <base>..HEAD --stat` scoped to the
  relevant path). Re-check counts and names.
- Confirm the file is written to `docs/release_notes/<release_version>.md` and the index is
  updated.

## Reporting

Summarize for the user:
- OLD baseline ref, NEW state, and target release version.
- The user-facing themes covered (and anything explicitly excluded).
- The output file path `docs/release_notes/<version>.md` and index update.

## Cross-platform portability

This skill follows the open **Agent Skills** spec (`SKILL.md` with `name`/`description`
frontmatter). It lives in `.opencode/skills/` for opencode but the identical folder can be
exposed to any agent platform — see the skill `README.md` for symlinking / `skills.paths`
wiring per platform.
