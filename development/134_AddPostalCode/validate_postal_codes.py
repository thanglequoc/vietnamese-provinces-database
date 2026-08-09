#!/usr/bin/env python3
"""Cross-validate parsed postal codes against DB wards_tmp.

Checks:
1. Every DB ward (province, name, unit) has a matching parsed row.
2. Every parsed row maps to a DB ward.
3. STT continuity per province.
"""
import re
import unicodedata
from collections import defaultdict

POSTAL_MD = "postal_codes.md"
DB_TSV = "db_wards.tsv"

PROVINCE_MAP = {
    "TỈNH AN GIANG": ("An Giang", "91"),
    "TỈNH BẮC NINH": ("Bắc Ninh", "24"),
    "TỈNH CÀ MAU": ("Cà Mau", "96"),
    "TỈNH CAO BẰNG": ("Cao Bằng", "04"),
    "TP. CẦN THƠ": ("Cần Thơ", "92"),
    "TP. ĐÀ NẴNG": ("Đà Nẵng", "48"),
    "TỈNH ĐẮK LẮK": ("Đắk Lắk", "66"),
    "TỈNH ĐIỆN BIÊN": ("Điện Biên", "11"),
    "TỈNH ĐỒNG NAI": ("Đồng Nai", "75"),
    "TỈNH ĐỒNG THÁP": ("Đồng Tháp", "82"),
    "TỈNH GIA LAI": ("Gia Lai", "52"),
    "TP. HÀ NỘI": ("Hà Nội", "01"),
    "TỈNH HÀ TĨNH": ("Hà Tĩnh", "42"),
    "TP. HẢI PHÒNG": ("Hải Phòng", "31"),
    "TP. HỒ CHÍ MINH": ("Hồ Chí Minh", "79"),
    "TP. HUẾ": ("Huế", "46"),
    "TỈNH HƯNG YÊN": ("Hưng Yên", "33"),
    "TỈNH KHÁNH HÒA": ("Khánh Hoà", "56"),
    "TỈNH LAI CHÂU": ("Lai Châu", "12"),
    "TỈNH LẠNG SƠN": ("Lạng Sơn", "20"),
    "TỈNH LÀO CAI": ("Lào Cai", "15"),
    "TỈNH LÂM ĐỒNG": ("Lâm Đồng", "68"),
    "TỈNH NINH BÌNH": ("Ninh Bình", "37"),
    "TỈNH NGHỆ AN": ("Nghệ An", "40"),
    "TỈNH PHÚ THỌ": ("Phú Thọ", "25"),
    "TỈNH QUẢNG NINH": ("Quảng Ninh", "22"),
    "TỈNH QUẢNG NGÃI": ("Quảng Ngãi", "51"),
    "TỈNH QUẢNG TRỊ": ("Quảng Trị", "44"),
    "TỈNH SƠN LA": ("Sơn La", "14"),
    "TỈNH TÂY NINH": ("Tây Ninh", "80"),
    "TỈNH TUYÊN QUANG": ("Tuyên Quang", "08"),
    "TỈNH THÁI NGUYÊN": ("Thái Nguyên", "19"),
    "TỈNH THANH HÓA": ("Thanh Hoá", "38"),
    "TỈNH VĨNH LONG": ("Vĩnh Long", "86"),
}

UNIT_PREFIX = {"Phường": "P.", "Xã": "X.", "Đặc khu": "Đặc khu"}


def strip_tone(s):
    s = unicodedata.normalize("NFD", s)
    s = "".join(c for c in s if not unicodedata.combining(c)).lower()
    s = s.replace("\u2019", "'").replace("\u2018", "'")
    s = re.sub(r"\s+", "", s)
    s = re.sub(r"[''-]", "", s)
    return s


def norm_name(n):
    n = re.sub(r"^(X\.|P\.|TT\.|TX\.)\s*", "", n).strip()
    n = re.sub(r"^Đặc\s+khu\s+", "", n).strip()
    return n


def parse_postal(path):
    rows = []
    cur_prov = None
    with open(path, encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            m = re.match(r"^##\s+(TỈNH|TP\.)\s+(.+)$", line)
            if m:
                cur_prov = f"{m.group(1)} {m.group(2)}"
                continue
            m = re.match(r"^\|\s*(\d+)\s*\|\s*(.+?)\s*\|\s*(\d+)\s*\|$", line)
            if m and cur_prov:
                rows.append((cur_prov, int(m.group(1)), m.group(2), m.group(3)))
    return rows


def load_db(path):
    wards = []
    with open(path, encoding="utf-8") as f:
        for line in f:
            if line.startswith("-+-") or line.startswith(" code "):
                continue
            parts = [p.strip() for p in line.split("|")]
            if len(parts) < 6:
                continue
            try:
                int(parts[0].strip())
            except ValueError:
                continue
            name = parts[1].strip()
            prov_code = parts[3].strip()
            unit = parts[4].strip()
            prov_name = parts[5].strip()
            wards.append((prov_code, prov_name, name, unit))
    return wards


def main():
    postal = parse_postal(POSTAL_MD)
    db = load_db(DB_TSV)
    print(f"parsed postal rows: {len(postal)}")
    print(f"db wards: {len(db)}")

    db_by_prov_name = defaultdict(list)
    for prov_code, prov_name, name, unit in db:
        db_by_prov_name[strip_tone(prov_name)].append((prov_code, prov_name, name, unit))

    unmatched = []
    multi = []
    errors = []
    used_db = set()
    stt_problems = []

    for prov_pdf, stt, name, code in postal:
        if prov_pdf not in PROVINCE_MAP:
            errors.append(f"UNKNOWN province in postal: {prov_pdf}")
            continue
        db_prov_name, db_prov_code = PROVINCE_MAP[prov_pdf]
        candidates = db_by_prov_name.get(strip_tone(db_prov_name), [])
        if not candidates:
            errors.append(f"NO db candidates for {prov_pdf}")
            continue

        if name == "#VALUE!":
            unmatched.append((prov_pdf, stt, name, code))
            continue

        bare = norm_name(name)
        exact = [c for c in candidates if c[2].lower() == bare.lower()]
        if len(exact) == 1:
            used_db.add((exact[0][0], exact[0][1], exact[0][2], exact[0][3]))
            continue
        norm = strip_tone(bare)
        matched = [c for c in candidates if strip_tone(c[2]) == norm]
        if len(matched) == 1:
            used_db.add((matched[0][0], matched[0][1], matched[0][2], matched[0][3]))
        elif len(matched) == 0:
            unmatched.append((prov_pdf, stt, name, code))
        else:
            multi.append((prov_pdf, stt, name, code, [c[2] for c in matched]))

    stt_by_prov = defaultdict(list)
    for prov_pdf, stt, name, code in postal:
        stt_by_prov[prov_pdf].append(stt)
    for prov, stts in stt_by_prov.items():
        for expected, got in zip(range(1, len(stts) + 1), stts):
            if got != expected:
                stt_problems.append((prov, expected, got))

    db_unmatched = []
    for prov_code, prov_name, name, unit in db:
        if (prov_code, prov_name, name, unit) not in used_db:
            db_unmatched.append((prov_code, prov_name, name, unit))

    print(f"\nunmatched postal rows: {len(unmatched)}")
    for u in unmatched:
        print("  ", u)
    print(f"\nmulti-match rows: {len(multi)}")
    for m in multi:
        print("  ", m)
    print(f"\nSTT sequence problems: {len(stt_problems)}")
    for p in stt_problems[:20]:
        print("  ", p)
    print(f"\nDB wards not matched by postal data: {len(db_unmatched)}")
    for d in db_unmatched:
        print("  ", d)
    if errors:
        print("\nerrors:")
        for e in errors:
            print("  ", e)


if __name__ == "__main__":
    main()
