# 134_AddPostalCode_v2 Branch Carryover Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create branch `134_AddPostalCode_v2` from `master` carrying the exact `dataset-generation-scripts/` content (plus security fix + docs) from `origin/134_AddPostalCode` as a single commit, so GitHub can merge it cleanly.

**Architecture:** `git checkout origin/134_AddPostalCode -- <paths>` copies exact file states from the source branch into the new branch's working tree and index, then a single commit freezes them. No history preserved. Generated data folders are excluded by simply not listing them in the path set.

**Tech Stack:** git 2.50.1; Go 1.24.0 (for optional build verification only).

## Global Constraints

- Working tree must be clean before creating the branch (untracked files may remain).
- New branch name must be exactly `134_AddPostalCode_v2`.
- Only these paths may be staged in the single commit:
  `dataset-generation-scripts/`, `.env.agent.example`, `.gitignore`, `AGENTS.md`,
  `development/134_AddPostalCode/`, `development/chunk-gis-sql-output*`, `docs/gis/`.
- Must NOT stage any generated data: `json/`, `elasticsearch/`, `mongodb/`, `mysql/`,
  `oracle/`, `postgresql/`, `sqlserver/`, `redis/`, or `development/` old-plan deletions.
- Parity check must pass: `git diff origin/134_AddPostalCode -- dataset-generation-scripts/` is empty.
- The design spec is at `docs/superpowers/specs/2026-08-18-134-add-postal-code-v2-branch-design.md`.

---

### Task 1: Create the branch and copy source-branch content

**Files:**
- Modify (via git checkout, no manual editing): `dataset-generation-scripts/`, `.env.agent.example`, `.gitignore`, `AGENTS.md`, `development/134_AddPostalCode/`, `development/chunk-gis-sql-output*`, `docs/gis/`

**Interfaces:**
- Consumes: `origin/134_AddPostalCode` remote-tracking branch (exists locally, verified).
- Produces: working tree + index with the included paths at source-branch content, on the new branch.

- [ ] **Step 1: Verify clean working tree and untracked files**

Run: `git status --short`
Expected: only `??` untracked entries (`.env.agent`, `docs/release_notes_v4.1.0.md`, `docs/release_notes_v4.2.0.md`). No `M`/`A`/`D` staged or unstaged modifications.

- [ ] **Step 2: Create the new branch from current `master`**

Run: `git checkout -b 134_AddPostalCode_v2`
Expected: `Switched to a new branch '134_AddPostalCode_v2'`

- [ ] **Step 3: Copy exact file states from source branch**

Run:
```bash
git checkout origin/134_AddPostalCode -- \
  dataset-generation-scripts/ \
  .env.agent.example \
  .gitignore \
  AGENTS.md \
  development/134_AddPostalCode/ \
  'development/chunk-gis-sql-output*' \
  docs/gis/
```
Expected: no output (files staged silently). Confirm with `git status --short` showing the included paths as `M`/`A` staged entries.

- [ ] **Step 4: Verify parity of the dataset-generation-scripts folder**

Run: `git diff origin/134_AddPostalCode -- dataset-generation-scripts/`
Expected: empty output (folder byte-identical to source branch).

- [ ] **Step 5: Verify no generated data was staged**

Run: `git diff --cached --name-only`
Expected: all staged paths start with one of the allowed prefixes. No `json/`, `elasticsearch/`, `mongodb/`, `mysql/`, `oracle/`, `postgresql/`, `sqlserver/`, `redis/`, and no `development/` files outside `134_AddPostalCode/` or `chunk-gis-sql-output*`.

---

### Task 2: Commit and verify final state

**Files:**
- Modify: (none — commit only)

**Interfaces:**
- Consumes: staged changes from Task 1.
- Produces: single commit on `134_AddPostalCode_v2`; clean working tree.

- [ ] **Step 1: Commit the staged changes**

Run: `git commit -m "feat: add postal code support to dataset generation scripts (134_AddPostalCode_v2)"`
Expected: commit succeeds; message printed with the new SHA and file summary.

- [ ] **Step 2: Confirm single-commit, clean status**

Run: `git status --short`
Expected: only the pre-existing untracked files listed (`??`). No staged or unstaged modifications.
Run: `git log --oneline master..HEAD`
Expected: exactly one commit line (the commit from Step 1).

- [ ] **Step 3: Optional build verification**

Run: `go build ./...` from `dataset-generation-scripts/`
Expected: compiles, or a clear error explaining an out-of-scope missing dependency. If it fails, report the error to the user — do NOT fix out-of-scope issues silently.

- [ ] **Step 4: Report branch readiness**

Report: branch name, commit SHA, `git diff origin/134_AddPostalCode -- dataset-generation-scripts/` = empty, and (if run) the `go build` result.
