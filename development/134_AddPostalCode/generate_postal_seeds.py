#!/usr/bin/env python3
"""Generate postal code seed JSON files from postal_codes.md and db_wards.tsv.

Outputs:
  ../../dataset-generation-scripts/resources/postal/province_postal_code_prefixes.json
  ../../dataset-generation-scripts/resources/postal/ward_postal_codes.json

The province summary table gives 2-digit prefix strings per province. The per-
province tables give 5-digit postal codes per ward. Province codes are resolved
by normalizing province names from db_wards.tsv.
"""
import json
import os
import re
import unicodedata

POSTAL_MD = "postal_codes.md"
DB_TSV = "db_wards.tsv"
OUT_DIR = "../../dataset-generation-scripts/resources/postal"

# Decree spellings that differ from the DB-convention ward name. Keyed by
# postal code (unique) so regeneration reproduces the corrected seed names.
WARD_NAME_OVERRIDES = {
    "66126": "Lang Biang - Đà Lạt",
    "66459": "B'Lao",
}


def strip_tone(s):
    s = unicodedata.normalize("NFD", s)
    s = "".join(c for c in s if not unicodedata.combining(c))
    s = s.replace("Đ", "D").replace("đ", "d")
    return s


def norm(s):
    s = s.strip().lower()
    s = s.replace("tp. ", "").replace("tỉnh ", "")
    s = re.sub(r"\s+", "", s)
    s = strip_tone(s)
    s = re.sub(r"['''-]", "", s)
    return s


def cells(line):
    inner = line.strip()
    if inner.startswith("|"):
        inner = inner[1:]
    if inner.endswith("|"):
        inner = inner[:-1]
    return [c.strip() for c in inner.split("|")]


def strip_unit_prefix(name):
    name = re.sub(r"^(X\.|P\.|TT\.|TX\.)\s*", "", name.strip()).strip()
    return re.sub(r"^Đặc\s+khu\s+", "", name).strip()


def load_province_code_map():
    code_by_norm = {}
    with open(DB_TSV, encoding="utf-8") as f:
        for line in f:
            if line.startswith(" code ") or line.startswith("-+-"):
                continue
            parts = [p.strip() for p in line.split("|")]
            if len(parts) < 6:
                continue
            try:
                int(parts[0])
            except ValueError:
                continue
            prov_name = parts[5].strip()
            prov_code = parts[3].strip()
            key = norm(prov_name)
            if key not in code_by_norm:
                code_by_norm[key] = prov_code
    return code_by_norm


def parse_province_prefixes(lines, code_by_norm):
    """Parse the 'Mã bưu chính cấp tỉnh' summary table."""
    result = []
    in_summary = False
    for line in lines:
        if line.strip().startswith("## "):
            in_summary = False
            continue
        c = cells(line)
        if any("Tên tỉnh" in x for x in c):
            in_summary = True
            continue
        if not in_summary:
            continue
        if len(c) < 3:
            continue
        try:
            int(c[0])
        except ValueError:
            continue
        name = c[1]
        prefix = ", ".join(p.strip() for p in c[2].split(","))
        if not name:
            continue
        code = code_by_norm.get(norm(name))
        if code is None:
            raise SystemExit(f"UNMATCHED province in summary: {name}")
        result.append({"code": code, "postal_code_prefix": prefix})
    return result


def parse_ward_postal_codes(lines, code_by_norm):
    result = []
    cur_prov_code = None
    for line in lines:
        stripped = line.strip()
        m = re.match(r"^##\s+(TỈNH|TP\.)\s+(.+)$", stripped)
        if m:
            prov_name = f"{m.group(1)} {m.group(2)}"
            cur_prov_code = code_by_norm.get(norm(prov_name))
            if cur_prov_code is None:
                raise SystemExit(f"UNMATCHED province header: {prov_name}")
            continue
        c = cells(line)
        if len(c) < 3 or cur_prov_code is None:
            continue
        try:
            int(c[0])
        except ValueError:
            continue
        name_raw = c[1]
        postal = c[2]
        if not re.fullmatch(r"\d{5}", postal):
            continue
        name = strip_unit_prefix(name_raw)
        name = WARD_NAME_OVERRIDES.get(postal, name)
        if not name:
            continue
        result.append({"province_code": cur_prov_code, "name": name, "postal_code": postal})
    return result


def main():
    lines = open(POSTAL_MD, encoding="utf-8").read().splitlines()
    code_by_norm = load_province_code_map()

    prefixes = parse_province_prefixes(lines, code_by_norm)
    wards = parse_ward_postal_codes(lines, code_by_norm)

    os.makedirs(OUT_DIR, exist_ok=True)
    with open(os.path.join(OUT_DIR, "province_postal_code_prefixes.json"), "w", encoding="utf-8") as f:
        json.dump(prefixes, f, ensure_ascii=False, indent=2)
    with open(os.path.join(OUT_DIR, "ward_postal_codes.json"), "w", encoding="utf-8") as f:
        json.dump(wards, f, ensure_ascii=False, indent=2)

    print(f"province prefixes: {len(prefixes)} (expected 34)")
    print(f"ward postal codes: {len(wards)} (expected 3321)")


if __name__ == "__main__":
    main()
