#!/usr/bin/env python3
"""Parse the markitdown output of the DOCX (Quyết định 2334/QĐ-BKHCN) and
reconstruct (province, stt, name, postal_code) rows.

The DOCX renders as proper markdown tables, but rows appear in two shapes:
  4-cell: ['', '1', 'X. An Phú', '90456']
  7-cell: ['', '1', '', 'X. Quảng Lâm', '', '21416', '']
Province headers:
  4-cell: ['1', '', '**TỈNH AN GIANG**', '']  OR
  7-cell: ['4', '', '', '**TỈNH CAO BẰNG**', '', '', '']
"""
import re
import unicodedata
from collections import OrderedDict

RAW = "postal_code_raw_docx.md"
OUT = "postal_codes.md"
ERR = "#VALUE!"

PROV_PAT = re.compile(r"^\*{0,2}(TỈNH|TP\.)\s+(.+?)\*{0,2}$")

# The source has a broken cell (#VALUE!) at Hải Phòng STT 34, code 05127.
# Cross-referenced against wards_tmp, this is the only unmatched Hải Phòng
# ward, so it is Nghi Dương (code 11713).
VALUE_FIX = {"TP. HẢI PHÒNG": {34: "X. Nghi Dương"}}


def cells(line):
    inner = line.strip()
    if inner.startswith("|"):
        inner = inner[1:]
    if inner.endswith("|"):
        inner = inner[:-1]
    return [c.strip() for c in inner.split("|")]


def split_tables(lines):
    """Return list of tables (each table = list of raw '|...|' lines)."""
    tables = []
    cur = []
    for l in lines:
        if l.strip().startswith("|") and l.strip() != "| --- |":
            cur.append(l)
        else:
            if cur:
                tables.append(cur)
                cur = []
    if cur:
        tables.append(cur)
    return tables


def parse_province_summary(lines):
    """Extract the province-level summary table (STT, province, zip, page)."""
    summary = []
    for table in split_tables(lines):
        header_seen = False
        for raw in table:
            c = cells(raw)
            if any("Tên tỉnh" in x or "Mã bưu chính" in x for x in c):
                header_seen = True
                continue
            if not header_seen:
                continue
            # row: [stt, province, zip, page]
            if len(c) < 4:
                continue
            try:
                stt = int(c[0])
            except ValueError:
                continue
            prov = c[1].strip()
            zipc = c[2].strip()
            page = c[3].strip()
            if not prov:
                continue
            summary.append((stt, prov, zipc, page))
    return summary


def parse_tables(lines):
    """Walk the doc tables, extracting province headers and data rows."""
    rows = []          # (province, stt, name, code)
    current_prov = None
    prov_stt_seen = set()
    errors = []

    for table in split_tables(lines):
        for raw in table:
            c = cells(raw)
            # skip pure-header rows
            if any("Số thứ tự" in x or "Đối tượng gán mã" in x or
                   "Mã bưu chính" in x for x in c):
                continue
            if any(x in ("(1)", "(2)", "(3)", "(4)") for x in c):
                continue

            # province header row: a cell matching "N" then a bold TỈNH/TP.
            nonempty = [x for x in c if x]
            prov_name = None
            for x in c:
                m = PROV_PAT.match(x)
                if m:
                    prov_name = f"{m.group(1)} {m.group(2)}"
                    break
            if prov_name:
                current_prov = prov_name
                continue

            # data row: has a 5-digit code somewhere (the source has a typo
            # "152213" for X. Tam Dương Bắc — corrected to 15221 below)
            code = None
            for x in c:
                if re.fullmatch(r"\d{5}", x) or x == "152213":
                    code = x
                    break
            if code is None:
                continue
            if code == "152213":
                code = "15221"

            # stt: the second cell (index 1) is the ward STT in both shapes
            try:
                stt = int(c[1])
            except (IndexError, ValueError):
                stt = None

            # name: find a cell matching X./P./Đặc khu (or #VALUE!)
            name = None
            for x in c:
                if x == ERR or re.match(r"^(X\.|P\.|TT\.|TX\.|Đặc khu)", x):
                    name = x
                    break
            if name == ERR and current_prov and stt is not None:
                fix = VALUE_FIX.get(current_prov, {}).get(stt)
                if fix:
                    name = fix
                else:
                    name = f"[[{ERR} — tên đơn vị lỗi trong nguồn]]"
            if name is None:
                name = "[[missing name]]"

            rows.append((current_prov, stt, name, code))

    return rows, errors


def main():
    lines = open(RAW, encoding="utf-8").read().splitlines()
    rows, errors = parse_tables(lines)
    print(f"parsed {len(rows)} rows, {len(errors)} errors")
    for e in errors:
        print("  ERROR:", e)

    by_prov = OrderedDict()
    for prov, stt, name, code in rows:
        by_prov.setdefault(prov, []).append((stt, name, code))

    with open(OUT, "w", encoding="utf-8") as f:
        f.write("# Mã bưu chính quốc gia — Phường, Xã và Đơn vị hành chính tương đương\n\n")
        f.write("> Nguyên bản: Quyết định số 2334/QĐ-BKHCN ngày 24 tháng 8 năm 2025 của Bộ Khoa học và Công nghệ\n\n")
        f.write("## Ghi chú nguồn dữ liệu\n\n")
        f.write("- **Hải Phòng, STT 34**: ô tên đơn vị trong nguồn bị lỗi (`#VALUE!`). Đã khớp chéo với `wards_tmp` → **X. Nghi Dương** (mã 05127).\n")
        f.write("- **Phú Thọ, STT 83**: `X. Tam Dương Bắc` ghi mã 6 chữ số `152213` trong nguồn (lỗi typo) → sửa thành **15221**.\n\n")

        summary = parse_province_summary(lines)
        f.write("## Mã bưu chính cấp tỉnh\n\n")
        f.write("| STT | Tên tỉnh, thành phố | Mã bưu chính | Trang |\n")
        f.write("|-----|---------------------|--------------|-------|\n")
        for stt, prov, zipc, page in summary:
            f.write(f"| {stt} | {prov} | {zipc} | {page} |\n")
        f.write("\n")

        for prov, prov_rows in by_prov.items():
            f.write(f"## {prov}\n\n")
            f.write("| STT | Đối tượng gán mã | Mã bưu chính |\n")
            f.write("|-----|------------------|--------------|\n")
            for stt, name, code in prov_rows:
                stt_s = "" if stt is None else str(stt)
                f.write(f"| {stt_s} | {name} | {code} |\n")
            f.write("\n")
    print(f"wrote {OUT}")


if __name__ == "__main__":
    main()
