# vn-provinces-release-notes skill

Generates a **bilingual (Vietnamese + English) end-user release note** for the Vietnamese
provinces database by diffing the current repository state (`HEAD`/working tree) against the
**most recent release tag** (or a user-provided tag/commit), focusing only on changes
beneficial to end users and **excluding** the internal `dataset-generation-scripts/` folder.
The note is stored under `docs/release_notes/<version>.md` and the
`docs/release_notes/README.md` index is updated.

This skill follows the open [Agent Skills](https://agent-skills.anthropic.dev/) spec
(`SKILL.md` with `name`/`description` frontmatter), so any platform that understands
`SKILL.md` can use it.

## Files

```
.opencode/skills/vn-provinces-release-notes/
├── SKILL.md                          # The skill definition + workflow (source of truth)
├── README.md                         # This file
└── scripts/
    ├── .gitignore
    └── classify_changes.py           # Classifies git changes into user-facing vs internal buckets
```

## Quick start

```bash
# From the repo root. OLD = most recent release tag, NEW = HEAD.
python3 .opencode/skills/vn-provinces-release-notes/scripts/classify_changes.py --base v4.2.0

# Compare against a specific baseline / head instead.
python3 .opencode/skills/vn-provinces-release-notes/scripts/classify_changes.py \
  --base abc1234 --head HEAD
```

The script prints a categorized Markdown summary: schema changes (new columns), dataset
content regenerated per format, new artifacts/formats, documentation changes, remaining
user-facing changes, the commit narrative, and the internal-only changes to exclude.

Run `python3 .../classify_changes.py --help` for all options.

The full end-to-end flow (baseline resolution, release-version prompt, bilingual note
writing, index update, verification) is documented in `SKILL.md`.

## Making the skill available on other platforms

The folder is platform-neutral. To use it outside opencode:

- **Claude Code** (repo-local, native): symlink into `.claude/skills/`:
  ```bash
  mkdir -p .claude/skills
  ln -s ../../.opencode/skills/vn-provinces-release-notes .claude/skills/vn-provinces-release-notes
  ```
- **opencode**: auto-discovers `.opencode/skills/` (no config needed). The repo-level
  `opencode.json` also lists it explicitly under `skills.paths` for clarity.
- **Cursor / Windsurf / others that support `skills.paths`**: point their skills path at
  `.opencode/skills` or symlink into the platform's skills directory.

Note: use a symlink or a build step — do not copy, to avoid drift between copies.
