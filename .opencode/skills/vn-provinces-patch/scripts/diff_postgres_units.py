#!/usr/bin/env python3
"""Diff two PostgreSQL VN provinces dataset files and emit an SQL upgrade patch.

Compares only the province-unit data (administrative_regions, administrative_units,
provinces, wards) between an OLD and a NEW version of
`postgresql/postgres_ImportData_vn_units.sql`. Ignores the leading generated-timestamp
header comment, all `--`/`/* */` comments, and any GIS content. Row ordering is not
significant (rows are compared as keyed sets).

The OLD version can be read from a local file (`--old-file`) or directly from git
history via `git show <ref>:<path>` (`--old-ref`).

Usage:
  diff_postgres_units.py --new <file> --old-file <file> [--output <patch.sql>] [--detect-renames]
  diff_postgres_units.py --new <file> --old-ref <ref> [--path <repo-path>] [--output <patch.sql>] [--detect-renames]

Output:
  * A summary of added / changed / deleted records per table (stdout).
  * The generated patch SQL (written to `--output` if given, else stdout).

Exit code 0 = success, 2 = no data changes, 1 = error.
"""

import argparse
import os
import re
import subprocess
import sys

DEFAULT_PATH = "postgresql/postgres_ImportData_vn_units.sql"

TABLE_ORDER = [
    "administrative_regions",
    "administrative_units",
    "provinces",
    "wards",
]

PK_COLUMN = {
    "administrative_regions": "id",
    "administrative_units": "id",
    "provinces": "code",
    "wards": "code",
}

INSERT_RE = re.compile(r"^\s*INSERT INTO (\w+)\s*\(([^)]*)\)\s*VALUES\b", re.IGNORECASE)


# ---------------------------------------------------------------------------
# SQL parsing
# ---------------------------------------------------------------------------

def split_top_level(text, sep):
    """Split `text` on `sep` ignoring separators inside single-quoted strings."""
    parts = []
    start = 0
    in_str = False
    i = 0
    length = len(text)
    while i < length:
        c = text[i]
        if in_str:
            if c == "'":
                if i + 1 < length and text[i + 1] == "'":
                    i += 1
                else:
                    in_str = False
        elif c == "'":
            in_str = True
        elif c == sep:
            parts.append(text[start:i])
            start = i + 1
        i += 1
    parts.append(text[start:])
    return parts


def unquote(value):
    value = value.strip()
    if value.upper() == "NULL":
        return None
    if value.startswith("'") and value.endswith("'"):
        return value[1:-1].replace("''", "'")
    return value


def parse_value_tuples(text):
    """Parse `(...)` value tuples out of the text following the VALUES keyword."""
    tuples = []
    i = 0
    length = len(text)
    while i < length:
        c = text[i]
        if c in " ,;\n\r\t":
            i += 1
            continue
        if c == "(":
            j = i
            depth = 0
            in_str = False
            while j < length:
                ch = text[j]
                if in_str:
                    if ch == "'":
                        if j + 1 < length and text[j + 1] == "'":
                            j += 1
                        else:
                            in_str = False
                else:
                    if ch == "'":
                        in_str = True
                    elif ch == "(":
                        depth += 1
                    elif ch == ")":
                        depth -= 1
                        if depth == 0:
                            break
                j += 1
            inner = text[i + 1:j]
            values = [unquote(v) for v in split_top_level(inner, ",")]
            tuples.append(values)
            i = j + 1
        else:
            i += 1
    return tuples


def parse_import_sql(text):
    """Parse INSERT statements into `{table: {pk_value: {column: value}}}`."""
    tables = {}
    lines = text.splitlines()
    n = len(lines)
    i = 0
    while i < n:
        line = lines[i]
        m = INSERT_RE.match(line)
        if not m:
            i += 1
            continue
        table = m.group(1)
        columns = [c.strip() for c in m.group(2).split(",")]
        buf = line[line.find("VALUES") + len("VALUES"):]
        j = i + 1
        while ";" not in buf and j < n:
            buf += lines[j]
            j += 1
        table_rows = tables.setdefault(table, {})
        pk = PK_COLUMN.get(table, "id" if "id" in columns else columns[0])
        pk_index = columns.index(pk)
        for values in parse_value_tuples(buf):
            if len(values) != len(columns):
                continue
            key = values[pk_index]
            table_rows[key] = dict(zip(columns, values))
        i = j
    return tables


def table_order(tables):
    known = [t for t in TABLE_ORDER if t in tables]
    extra = sorted(t for t in tables if t not in TABLE_ORDER)
    return known + extra


def read_git_file(ref, path, repo=None):
    cmd = ["git", "show", f"{ref}:{path}"]
    if repo:
        cmd = ["git", "-C", repo, "show", f"{ref}:{path}"]
    proc = subprocess.run(cmd, capture_output=True, text=True)
    if proc.returncode != 0:
        sys.stderr.write(f"git show {ref}:{path} failed:\n{proc.stderr}\n")
        sys.exit(1)
    return proc.stdout


# ---------------------------------------------------------------------------
# SQL generation
# ---------------------------------------------------------------------------

def sql_value(value):
    if value is None:
        return "NULL"
    return "'" + value.replace("'", "''") + "'"


def gen_insert(table, columns, record):
    cols_sql = ",".join(columns)
    vals_sql = ",".join(sql_value(record[c]) for c in columns)
    return f"INSERT INTO {table}({cols_sql}) VALUES({vals_sql});"


def gen_update(table, pk, key, old_record, new_record):
    sets = []
    for column in new_record:
        if column == pk:
            continue
        if old_record.get(column) != new_record.get(column):
            sets.append(f"{column}={sql_value(new_record[column])}")
    if not sets:
        return None
    return f"UPDATE {table} SET {', '.join(sets)} WHERE {pk}={sql_value(key)};"


def gen_delete(table, pk, key):
    return f"DELETE FROM {table} WHERE {pk}={sql_value(key)};"


# ---------------------------------------------------------------------------
# Diffing
# ---------------------------------------------------------------------------

def diff_tables(old_tables, new_tables, detect_renames):
    """Return per-table diffs and the generated patch SQL."""
    patch_lines = []
    summary = {}
    renames = {}

    for table in table_order(set(old_tables) | set(new_tables)):
        old_rows = old_tables.get(table, {})
        new_rows = new_tables.get(table, {})
        sample = next(iter(new_rows.values()), None) or next(iter(old_rows.values()), None)
        pk = PK_COLUMN.get(table, "id" if sample and "id" in sample else "code")

        added = [k for k in new_rows if k not in old_rows]
        deleted = [k for k in old_rows if k not in new_rows]
        changed = [k for k in new_rows if k in old_rows and old_rows[k] != new_rows[k]]

        if detect_renames and pk == "code" and added and deleted:
            remaining_added = list(added)
            remaining_deleted = list(deleted)
            for d in list(remaining_deleted):
                for a in list(remaining_added):
                    d_copy = {k: v for k, v in old_rows[d].items() if k != pk}
                    a_copy = {k: v for k, v in new_rows[a].items() if k != pk}
                    if d_copy == a_copy:
                        renames.setdefault(table, []).append((d, a))
                        remaining_deleted.remove(d)
                        remaining_added.remove(a)
                        break
            added = remaining_added
            deleted = remaining_deleted

        added.sort()
        deleted.sort()
        changed.sort()

        summary[table] = {"added": len(added), "changed": len(changed), "deleted": len(deleted)}

        if not (added or changed):
            continue

        patch_lines.append("")
        patch_lines.append(f"-- {table}: {len(added)} added, {len(changed)} changed")
        if renames.get(table):
            patch_lines.append(f"--   detected renames: {', '.join(f'{d}->{a}' for d, a in renames[table])}")

        new_columns = list(next(iter(new_rows.values())).keys()) if new_rows else []
        for key in added:
            patch_lines.append(gen_insert(table, new_columns, new_rows[key]))
        for key in changed:
            statement = gen_update(table, pk, key, old_rows[key], new_rows[key])
            if statement:
                patch_lines.append(statement)

    # DELETEs in reverse FK order (wards before provinces, etc.)
    order = table_order(set(old_tables) | set(new_tables))
    for table in reversed(order):
        old_rows = old_tables.get(table, {})
        new_rows = new_tables.get(table, {})
        pk = PK_COLUMN.get(table, "code")
        deleted = [k for k in old_rows if k not in new_rows]
        if detect_renames and pk == "code":
            rename_old = {d for d, _ in renames.get(table, [])}
            deleted = [k for k in deleted if k not in rename_old]
        deleted.sort()
        if not deleted:
            continue
        patch_lines.append("")
        patch_lines.append(f"-- {table}: {len(deleted)} deleted")
        for key in deleted:
            patch_lines.append(gen_delete(table, pk, key))

    return summary, "\n".join(patch_lines).strip(), renames


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def build_summary_text(summary):
    lines = ["Summary of province-unit data changes:"]
    for table in table_order(summary):
        info = summary[table]
        lines.append(f"  {table:<26} added={info['added']:<4} changed={info['changed']:<4} deleted={info['deleted']}")
    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--new", required=True, help="NEW dataset file (working tree)")
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--old-file", help="OLD dataset file")
    source.add_argument("--old-ref", help="git ref (commit/tag) for the OLD dataset")
    parser.add_argument("--path", default=DEFAULT_PATH, help="repo-relative path used with --old-ref")
    parser.add_argument("--repo", help="path to the git repository (defaults to cwd)")
    parser.add_argument("--output", help="write patch SQL to this file")
    parser.add_argument("--detect-renames", action="store_true",
                        help="treat deleted+added rows identical except code as a code rename")
    args = parser.parse_args()

    if args.old_file:
        with open(args.old_file, "r", encoding="utf-8") as fh:
            old_text = fh.read()
    else:
        old_text = read_git_file(args.old_ref, args.path, repo=args.repo)

    with open(args.new, "r", encoding="utf-8") as fh:
        new_text = fh.read()

    old_tables = parse_import_sql(old_text)
    new_tables = parse_import_sql(new_text)

    if not old_tables and not new_tables:
        sys.stderr.write("No INSERT statements parsed — check the input files.\n")
        sys.exit(1)

    summary, patch_sql, renames = diff_tables(old_tables, new_tables, detect_renames=args.detect_renames)

    print(build_summary_text(summary))
    if renames:
        for table, pairs in renames.items():
            print(f"Detected renames in {table}: " + ", ".join(f"{d} -> {a}" for d, a in pairs))

    if not patch_sql:
        print("No data changes between OLD and NEW. No patch generated.")
        sys.exit(2)

    header = (
        "-- =====================================================================\n"
        "-- VN provinces data patch (auto-generated by the vn-provinces-patch skill)\n"
        "-- Data only: administrative_regions, administrative_units, provinces, wards\n"
        "-- ====================================================================="
    )
    patch_sql = header + "\n" + patch_sql + "\n"

    if args.output:
        with open(args.output, "w", encoding="utf-8") as fh:
            fh.write(patch_sql + "\n")
        print(f"Patch written to: {args.output}")
    else:
        print("\n" + patch_sql)

    sys.exit(0)


if __name__ == "__main__":
    main()
