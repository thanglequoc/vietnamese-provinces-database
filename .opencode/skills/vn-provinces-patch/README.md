# vn-provinces-patch skill

Generates a **data-only** SQL upgrade patch for the Vietnamese provinces database by diffing the
current `postgresql/postgres_ImportData_vn_units.sql` against an earlier git commit/tag (default
`HEAD`), excluding the generated-timestamp header comment and all GIS data. The patch is verified
against the local PostGIS docker container, then stored under `patch/<release_version>/` together
with a `README.md`.

This skill follows the open [Agent Skills](https://agent-skills.anthropic.dev/) spec
(`SKILL.md` with `name`/`description` frontmatter), so any platform that understands `SKILL.md` can
use it.

## Files

```
.opencode/skills/vn-provinces-patch/
├── SKILL.md                          # The skill definition + workflow (source of truth)
├── README.md                         # This file
└── scripts/
    └── diff_postgres_units.py        # Diff engine: parses both SQL files, emits patch + summary
```

## Quick start

```bash
# From the repo root. OLD = HEAD, NEW = working tree.
python3 .opencode/skills/vn-provinces-patch/scripts/diff_postgres_units.py \
  --new postgresql/postgres_ImportData_vn_units.sql \
  --old-ref HEAD \
  --path postgresql/postgres_ImportData_vn_units.sql \
  --detect-renames \
  --output /tmp/vn_units_patch_draft.sql

# Diff against a specific baseline instead.
python3 .opencode/skills/vn-provinces-patch/scripts/diff_postgres_units.py \
  --new postgresql/postgres_ImportData_vn_units.sql \
  --old-ref v4.2.0 \
  --path postgresql/postgres_ImportData_vn_units.sql \
  --detect-renames \
  --output /tmp/vn_units_patch_draft.sql

# Diff two local files.
python3 .opencode/skills/vn-provinces-patch/scripts/diff_postgres_units.py \
  --new postgresql/postgres_ImportData_vn_units.sql \
  --old-file /path/to/previous/postgres_ImportData_vn_units.sql \
  --output /tmp/vn_units_patch_draft.sql
```

The script prints a per-table summary (added / changed / deleted) and writes the draft patch.
Exit code `2` = no data changes; `0` = patch generated; `1` = error.

Run `python3 .../diff_postgres_units.py --help` for all options.

The full end-to-end flow (baseline resolution, patch finalization, release-version folder,
**mandatory PostGIS verification**) is documented in `SKILL.md`.

## Making the skill available on other platforms

The folder is platform-neutral. To use it outside opencode:

- **Claude Code** (repo-local, native): symlink into `.claude/skills/`:
  ```bash
  mkdir -p .claude/skills
  ln -s ../../.opencode/skills/vn-provinces-patch .claude/skills/vn-provinces-patch
  ```
- **opencode**: auto-discovers `.opencode/skills/` (no config needed). The repo-level
  `opencode.json` also lists it explicitly under `skills.paths` for clarity.
- **Cursor / Windsurf / others that support `skills.paths`**: point their skills path at
  `.opencode/skills` or symlink into the platform's skills directory.

Note: use a symlink or a build step — do not copy, to avoid drift between copies.
