#!/usr/bin/env python3
"""Regenerate the dataset download tables from a downloads.json manifest.

Reads the manifest produced by ``package-datasets.sh`` and rewrites the dataset
table between the DOWNLOAD_TABLE markers in README.md / README_vi.md.

Usage:
    update_download_tables.py --downloads <downloads.json> [--repo-root DIR]
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path

DATASETS = {
    "postgresql": {"en": "PostgreSQL / PostGIS", "vi": "PostgreSQL / PostGIS", "fmt_en": "SQL", "fmt_vi": "SQL"},
    "mysql": {"en": "MySQL / MariaDB", "vi": "MySQL / MariaDB", "fmt_en": "SQL", "fmt_vi": "SQL"},
    "sqlserver": {"en": "Microsoft SQL Server", "vi": "Microsoft SQL Server", "fmt_en": "SQL", "fmt_vi": "SQL"},
    "oracle": {"en": "Oracle", "vi": "Oracle", "fmt_en": "SQL", "fmt_vi": "SQL"},
    "json": {"en": "JSON", "vi": "JSON", "fmt_en": "JSON", "fmt_vi": "JSON"},
    "mongodb": {"en": "MongoDB", "vi": "MongoDB", "fmt_en": "NoSQL", "fmt_vi": "NoSQL"},
    "redis": {"en": "Redis", "vi": "Redis", "fmt_en": "NoSQL", "fmt_vi": "NoSQL"},
    "elasticsearch": {"en": "Elasticsearch", "vi": "Elasticsearch", "fmt_en": "NoSQL", "fmt_vi": "NoSQL"},
}

ROOT_ORDER = ["postgresql", "mysql", "sqlserver", "oracle", "json", "mongodb", "redis", "elasticsearch"]

ROOT_START = "<!-- DOWNLOAD_TABLE:START -->"
ROOT_END = "<!-- DOWNLOAD_TABLE:END -->"


def replace_block(text: str, start: str, end: str, body: str, path: Path) -> str:
    pattern = re.compile(re.escape(start) + r".*?" + re.escape(end), re.DOTALL)
    if not pattern.search(text):
        raise SystemExit(f"error: markers {start} / {end} not found in {path}")
    return pattern.sub(lambda _m: f"{start}\n{body}\n{end}", text, count=1)


def root_table(lang: str, version: str, entries: dict) -> str:
    if lang == "vi":
        header = "| Bộ dữ liệu | Định dạng | Phiên bản | Kích thước | Tải xuống |"
        sep = "|------------|-----------|-----------|------------|-----------|"
    else:
        header = "| Dataset | Format | Version | Size | Download |"
        sep = "|---------|--------|---------|------|----------|"

    rows = [header, sep]
    for dataset_id in ROOT_ORDER:
        entry = entries.get(dataset_id)
        if not entry:
            continue
        meta = DATASETS[dataset_id]
        name = meta[lang]
        fmt = meta[f"fmt_{lang}"]
        rows.append(
            f"| {name} | {fmt} | {version} | {entry['size_human']} | "
            f"[{entry['archive']}]({entry['url']}) |"
        )
    return "\n".join(rows)


def update_root_readme(path: Path, lang: str, version: str, entries: dict) -> bool:
    text = path.read_text(encoding="utf-8")
    updated = replace_block(text, ROOT_START, ROOT_END, root_table(lang, version, entries), path)
    if updated != text:
        path.write_text(updated, encoding="utf-8")
        return True
    return False


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--downloads", required=True, type=Path, help="Path to downloads.json")
    parser.add_argument("--repo-root", default=".", type=Path, help="Repository root (default: .)")
    args = parser.parse_args()

    manifest = json.loads(args.downloads.read_text(encoding="utf-8"))
    version = manifest["version"]
    entries = {entry["id"]: entry for entry in manifest.get("datasets", [])}

    root = args.repo_root
    targets = [
        (root / "README.md", lambda p: update_root_readme(p, "en", version, entries)),
        (root / "README_vi.md", lambda p: update_root_readme(p, "vi", version, entries)),
    ]

    changed = 0
    for path, updater in targets:
        if not path.exists():
            print(f"skip (missing): {path}", file=sys.stderr)
            continue
        if updater(path):
            changed += 1
            print(f"updated: {path}")
        else:
            print(f"unchanged: {path}")

    print(f"Done. {changed} file(s) updated for {version}.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
