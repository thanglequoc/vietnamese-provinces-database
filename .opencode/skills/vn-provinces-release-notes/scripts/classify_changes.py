#!/usr/bin/env python3
"""Classify git changes between two refs for the VN provinces release-notes skill.

Runs read-only git commands and emits a categorized Markdown summary separating
**end-user-facing** changes (published dataset formats, READMEs, user docs, new artifacts)
from **internal-only** changes (dataset-generation-scripts, agent tooling, dev docs, etc.)
that must be excluded from the release note.

Usage:
  classify_changes.py --base <ref> [--head <ref>] [--no-commit-log]

Example:
  classify_changes.py --base v4.2.0
  classify_changes.py --base abc1234 --head HEAD

Exit code 0 = success, 1 = error.
"""

import argparse
import subprocess
import sys

USER_FACING_DIRS = [
    "postgresql/",
    "mysql/",
    "sqlserver/",
    "oracle/",
    "mongodb/",
    "redis/",
    "json/",
    "elasticsearch/",
]

INTERNAL_PATTERNS = [
    "dataset-generation-scripts/",
    ".opencode/",
    ".github/",
    ".claude/",
    "development/",
    "docs/superpowers/",
    "opencode.json",
    ".env.agent",
]

# Paths that are NOT end-user-facing even though they live at repo top level.
EXCLUDE_PATHS = INTERNAL_PATTERNS + [
    "docs/release_notes/",
    "docs/release_notes_v",
    "patch/",
    ".gitignore",
    "AGENTS.md",
]

# Files that carry the actual dataset content per format.
DATASET_CONTENT_SUFFIXES = (
    "_ImportData_vn_units.sql",
    "_ImportData_gis-part-",
    "_ImportData_gis.sql",
    "generated_data_vn_units",
    "mongo_data_vn_unit",
    "mongo_data_vn_province_gis",
    "mongo_data_vn_ward_gis",
    "vn_provinces_dataset",
    ".ndjson",
    "provinces.json",
    "wards.json",
)

CREATE_TABLES_MARKERS = ("CreateTables", "create_tables", "db_table_init")

DOC_PATH_MARKERS = ("README", "docs/gis/")

# Dense path prefixes whose many files should be aggregated into a single bullet,
# e.g. json/geojson/ (3,355 files) or <format>/gis/ part files.
DENSE_PREFIXES = (
    "json/geojson/",
    "postgresql/gis/",
    "mysql/gis/",
    "sqlserver/gis/",
    "oracle/gis/",
)


def run_git(args, repo="."):
    """Run a read-only git command, return stdout as str."""
    proc = subprocess.run(
        ["git", "-C", repo] + args,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        raise RuntimeError("git %s failed: %s" % (" ".join(args), proc.stderr.strip()))
    return proc.stdout


def is_internal(path):
    return any(path.startswith(p) for p in INTERNAL_PATTERNS)


def is_excluded(path):
    if is_internal(path):
        return True
    return any(path.startswith(p) for p in EXCLUDE_PATHS)


def parse_diff_status(text):
    """Parse `git diff --name-status` output into a list of (status, path) tuples."""
    changes = []
    for line in text.splitlines():
        if not line.strip():
            continue
        parts = line.split("\t")
        status = parts[0][0]  # A / M / D / R / C
        if status == "R":
            # rename: old\tnew -> use the new path
            changes.append((status, parts[2]))
        else:
            changes.append((status, parts[-1]))
    return changes


def added_columns_for(path, base, head):
    """Return column names newly added in a CREATE TABLE file between base and head."""
    diff = run_git(["diff", base, head, "--", path])
    columns = []
    for line in diff.splitlines():
        if not line.startswith("+") or line.startswith("+++"):
            continue
        raw = line[1:]
        if not (raw.startswith("\t") or raw.startswith(" ")):
            continue
        token = raw.split()[0].strip().rstrip(",")
        if not token or token.startswith("(") or any(c in token for c in "();"):
            continue
        if token.upper() in {"CONSTRAINT", "PRIMARY", "UNIQUE", "FOREIGN", "CHECK",
                             "CREATE", "ALTER", "TABLE", "INSERT", "SET", "REFERENCES",
                             "INDEX"}:
            continue
        columns.append(token)
    return columns


def format_of(path):
    """Best-effort top-level format directory for a dataset path."""
    for d in USER_FACING_DIRS:
        if path.startswith(d):
            return d.rstrip("/")
    return None


def aggregate_added(paths):
    """Group added file paths into concise release-note bullets.

    Dense subdirectories (geojson, gis parts) collapse into a count; everything else is
    listed as a single path. Returns a list of bullet strings.
    """
    dense = {}
    singles = []
    for path in paths:
        matched = next((p for p in DENSE_PREFIXES if path.startswith(p)), None)
        if matched:
            dense[matched] = dense.get(matched, 0) + 1
        else:
            singles.append(path)
    bullets = []
    for prefix in sorted(dense):
        bullets.append("`%s` — %d new file(s)" % (prefix, dense[prefix]))
    for path in sorted(singles):
        bullets.append("`%s`" % path)
    return bullets


def aggregate_modified(paths):
    """Group modified file paths, collapsing dense subdirectories into a count."""
    dense = {}
    singles = []
    for path in paths:
        matched = next((p for p in DENSE_PREFIXES if path.startswith(p)), None)
        if matched:
            dense[matched] = dense.get(matched, 0) + 1
        else:
            singles.append(path)
    bullets = []
    for prefix in sorted(dense):
        bullets.append("`%s` — %d modified file(s)" % (prefix, dense[prefix]))
    for path in sorted(singles):
        bullets.append("`%s`" % path)
    return bullets


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base", required=True, help="OLD ref (tag or commit)")
    parser.add_argument("--head", default="HEAD", help="NEW ref (default: HEAD)")
    parser.add_argument("--no-commit-log", action="store_true",
                        help="Skip the commit-subject review section")
    args = parser.parse_args()

    changes = parse_diff_status(
        run_git(["diff", "-M", "--name-status", args.base, args.head]))

    # ---- Split into buckets -------------------------------------------------
    internal = []
    new_paths = []
    schema_files = []
    dataset_files = []
    doc_files = []
    other_user = []

    for status, path in changes:
        if is_internal(path):
            internal.append((status, path))
            continue
        if is_excluded(path):
            continue
        if any(m in path for m in CREATE_TABLES_MARKERS):
            schema_files.append(path)
        elif any(path.endswith(s) for s in DATASET_CONTENT_SUFFIXES):
            dataset_files.append(path)
        elif any(m in path for m in DOC_PATH_MARKERS):
            doc_files.append(path)
        elif status == "A":
            new_paths.append(path)
        else:
            other_user.append(path)

    # ---- Schema additions ----------------------------------------------------
    schema_notes = {}
    for path in schema_files:
        cols = added_columns_for(path, args.base, args.head)
        if cols:
            fmt = format_of(path) or path
            schema_notes.setdefault(fmt, [])
            schema_notes[fmt].extend(cols)

    # ---- Dataset regeneration per format ------------------------------------
    per_format = {}
    for path in dataset_files:
        fmt = format_of(path)
        if fmt:
            per_format[fmt] = per_format.get(fmt, 0) + 1

    # ---- Commit narrative ----------------------------------------------------
    commit_lines = []
    if not args.no_commit_log:
        try:
            commit_lines = run_git(["log", "%s..%s" % (args.base, args.head),
                                    "--oneline"]).splitlines()
        except RuntimeError:
            commit_lines = []

    # ---- Emit summary ---------------------------------------------------------
    out = []
    out.append("# Change classification: %s..%s" % (args.base, args.head))
    out.append("")

    total_user = (len(new_paths) + len(schema_files) + len(dataset_files)
                  + len(doc_files) + len(other_user))
    out.append("**User-facing changes:** %d path(s) | **Internal-only:** %d path(s)"
               % (total_user, len(internal)))
    out.append("")

    if schema_notes:
        out.append("## Schema changes (new columns)")
        out.append("")
        for fmt in sorted(schema_notes):
            cols = sorted(set(schema_notes[fmt]))
            out.append("- **%s**: %s" % (fmt, ", ".join("`%s`" % c for c in cols)))
        out.append("")

    if per_format:
        out.append("## Dataset content regenerated (files changed per format)")
        out.append("")
        for fmt in sorted(per_format):
            out.append("- **%s**: %d file(s)" % (fmt, per_format[fmt]))
        out.append("")

    if new_paths:
        out.append("## New artifacts / formats")
        out.append("")
        out.extend("- %s" % b for b in aggregate_added(new_paths))
        out.append("")

    if doc_files:
        out.append("## Documentation changes")
        out.append("")
        for p in sorted(set(doc_files)):
            out.append("- `%s`" % p)
        out.append("")

    if other_user:
        out.append("## Other user-facing changes")
        out.append("")
        out.extend("- %s" % b for b in aggregate_modified(other_user))
        out.append("")

    if commit_lines:
        out.append("## Commit narrative (filter out internal-only subjects)")
        out.append("")
        for line in commit_lines:
            out.append("- %s" % line)
        out.append("")

    out.append("## Internal-only changes (EXCLUDE from the release note)")
    out.append("")
    if internal:
        by_dir = {}
        for _, path in internal:
            top = next(p for p in INTERNAL_PATTERNS if path.startswith(p))
            by_dir[top] = by_dir.get(top, 0) + 1
        for top in sorted(by_dir):
            out.append("- `%s`: %d path(s)" % (top, by_dir[top]))
    else:
        out.append("- none")
    out.append("")

    print("\n".join(out))


if __name__ == "__main__":
    try:
        main()
    except RuntimeError as exc:
        print("error: %s" % exc, file=sys.stderr)
        sys.exit(1)
