# GIS Geometry Fix — Before/After Comparison Report

**Generated**: 2026-07-24
**Source**: `postgresql_ImportData_gis_2026-07-12__19_50_50.sql` (old) vs `postgresql_ImportData_gis_2026-07-20__23_14_35.sql` (new)
**Fix**: `ST_MakeValid` + `ST_CollectionExtract(..., 3)` on 78 self-intersecting ward geometries
**Tables**: `gis_wards` / `gis_provinces` (published schema, not `sapnhap_geojson_objects`)

## Summary Dashboard

| Metric | Value |
|--------|-------|
| Total wards compared | 78 |
| Total provinces compared | 34 |
| 🟢 OK (< 0.001%) | 78 |
| 🟡 WARN (0.001%-0.01%) | 0 |
| 🔴 ALARM (> 0.01%) | 0 |
| Max area change (wards) | 0.000000% (all wards identical) |
| Provinces with unexpected change | 0 |
| Old wards confirmed invalid | 78/78 ✅ |
| New wards confirmed valid | 78/78 ✅ |
| Remaining invalid geometries (new) | 0 ✅ |
| FK orphaned records | 0 ✅ |
| NULL geometries | 0 ✅ |

## Tier 1: Area Comparison — Wards (78 Fixed)

All 78 wards show **0.000000% area change** — the `ST_MakeValid` fix preserved geodesic area perfectly.

| Rank | ward_code | Name | Prov | Old Area (km²) | New Area (km²) | Diff (km²) | Diff % | Status |
|------|-----------|------|------|----------------|----------------|------------|--------|--------|
| 1 | 03358 | Núa Ngam | 11 | 266.854418 | 266.854418 | 0.000000 | 0.000000 | 🟢 OK |
| 2 | 03356 | Sam Mứn | 11 | 156.174317 | 156.174317 | 0.000000 | 0.000000 | 🟢 OK |
| 3 | 03352 | Thanh An | 11 | 57.395060 | 57.395060 | 0.000000 | 0.000000 | 🟢 OK |
| 4 | 03460 | Hua Bum | 12 | 355.865254 | 355.865254 | 0.000000 | 0.000000 | 🟢 OK |
| 5 | 03760 | Mường Bám | 14 | 76.198779 | 76.198779 | 0.000000 | 0.000000 | 🟢 OK |
| 6 | 03583 | Khổng Lào | 12 | 188.089552 | 188.089552 | 0.000000 | 0.000000 | 🟢 OK |
| 7 | 03434 | Nậm Hàng | 12 | 336.014425 | 336.014425 | 0.000000 | 0.000000 | 🟢 OK |
| 8 | 19066 | Bắc Gianh | 44 | 31.138823 | 31.138823 | 0.000000 | 0.000000 | 🟢 OK |
| 9 | 05542 | Lam Vỹ | 19 | 72.520198 | 72.520198 | 0.000000 | 0.000000 | 🟢 OK |
| 10 | 25510 | Trà Vong | 80 | 92.021761 | 92.021761 | 0.000000 | 0.000000 | 🟢 OK |
| 11 | 06565 | Khuất Xá | 20 | 124.944876 | 124.944876 | 0.000000 | 0.000000 | 🟢 OK |
| 12 | 06541 | Mẫu Sơn | 20 | 137.198853 | 137.198853 | 0.000000 | 0.000000 | 🟢 OK |
| 13 | 30028 | An Long | 82 | 72.357783 | 72.357783 | 0.000000 | 0.000000 | 🟢 OK |
| 14 | 06577 | Thống Nhất | 20 | 190.595866 | 190.595866 | 0.000000 | 0.000000 | 🟢 OK |
| 15 | 31673 | Trần Đề | 92 | 89.385097 | 89.385097 | 0.000000 | 0.000000 | 🟢 OK |
| 16 | 06607 | Xuân Dương | 20 | 206.615999 | 206.615999 | 0.000000 | 0.000000 | 🟢 OK |
| 17 | 00958 | Thượng Sơn | 08 | 142.614826 | 142.614826 | 0.000000 | 0.000000 | 🟢 OK |
| 18 | 16186 | Hoá Quỳ | 38 | 116.529129 | 116.529129 | 0.000000 | 0.000000 | 🟢 OK |
| 19 | 19351 | Nam Đông Hà | 44 | 34.799397 | 34.799397 | 0.000000 | 0.000000 | 🟢 OK |
| 20 | 20656 | Nông Sơn | 48 | 112.802903 | 112.802903 | 0.000000 | 0.000000 | 🟢 OK |
| 21 | 20965 | Núi Thành | 48 | 124.511742 | 124.511742 | 0.000000 | 0.000000 | 🟢 OK |
| 22 | 20669 | Quế Phước | 48 | 359.777018 | 359.777018 | 0.000000 | 0.000000 | 🟢 OK |
| 23 | 22741 | Bảo An | 56 | 18.465190 | 18.465190 | 0.000000 | 0.000000 | 🟢 OK |
| 24 | 22888 | Phước Dinh | 56 | 154.526045 | 154.526045 | 0.000000 | 0.000000 | 🟢 OK |
| 25 | 22624 | Tây Khánh Vĩnh | 56 | 294.979747 | 294.979747 | 0.000000 | 0.000000 | 🟢 OK |
| 26 | 23586 | Hội Phú | 52 | 34.817295 | 34.817295 | 0.000000 | 0.000000 | 🟢 OK |
| 27 | 21835 | Bình Phú | 52 | 200.659520 | 200.659520 | 0.000000 | 0.000000 | 🟢 OK |
| 28 | 21997 | Canh Liên | 52 | 325.940930 | 325.940930 | 0.000000 | 0.000000 | 🟢 OK |
| 29 | 23764 | Ia Grai | 52 | 238.901368 | 238.901368 | 0.000000 | 0.000000 | 🟢 OK |
| 30 | 23767 | Ia Hrung | 52 | 169.553335 | 169.553335 | 0.000000 | 0.000000 | 🟢 OK |
| 31 | 23728 | Ia Khươl | 52 | 352.241208 | 352.241208 | 0.000000 | 0.000000 | 🟢 OK |
| 32 | 21892 | Xuân An | 52 | 67.528267 | 67.528267 | 0.000000 | 0.000000 | 🟢 OK |
| 33 | 24502 | Ea Phê | 66 | 85.119583 | 85.119583 | 0.000000 | 0.000000 | 🟢 OK |
| 34 | 24846 | Lang Biang - Đà Lạt | 68 | 322.899590 | 322.899590 | 0.000000 | 0.000000 | 🟢 OK |
| 35 | 25459 | Tân Ninh | 80 | 20.980087 | 20.980087 | 0.000000 | 0.000000 | 🟢 OK |
| 36 | 25567 | Ninh Thạnh | 80 | 52.507470 | 52.507470 | 0.000000 | 0.000000 | 🟢 OK |
| 37 | 25585 | Châu Thành | 80 | 94.047092 | 94.047092 | 0.000000 | 0.000000 | 🟢 OK |
| 38 | 25588 | Hảo Đước | 80 | 93.593654 | 93.593654 | 0.000000 | 0.000000 | 🟢 OK |
| 39 | 28087 | Nhựt Tảo | 80 | 34.713237 | 34.713237 | 0.000000 | 0.000000 | 🟢 OK |
| 40 | 28075 | Tân Trụ | 80 | 30.974324 | 30.974324 | 0.000000 | 0.000000 | 🟢 OK |
| 41 | 25498 | Thạnh Bình | 80 | 175.383741 | 175.383741 | 0.000000 | 0.000000 | 🟢 OK |
| 42 | 26461 | Xuân Định | 75 | 51.850651 | 51.850651 | 0.000000 | 0.000000 | 🟢 OK |
| 43 | 25843 | Tây Nam | 79 | 115.373483 | 115.373483 | 0.000000 | 0.000000 | 🟢 OK |
| 44 | 25777 | Dầu Tiếng | 79 | 186.242974 | 186.242974 | 0.000000 | 0.000000 | 🟢 OK |
| 45 | 25807 | Thanh An | 79 | 139.991874 | 139.991874 | 0.000000 | 0.000000 | 🟢 OK |
| 46 | 31261 | Cờ Đỏ | 92 | 44.659844 | 44.659844 | 0.000000 | 0.000000 | 🟢 OK |
| 47 | 31249 | Thạnh Phú | 92 | 99.376927 | 99.376927 | 0.000000 | 0.000000 | 🟢 OK |
| 48 | 32071 | Trí Phải | 96 | 166.583348 | 166.583348 | 0.000000 | 0.000000 | 🟢 OK |
| 49 | 00832 | Bạch Đích | 08 | 95.128077 | 95.128077 | 0.000000 | 0.000000 | 🟢 OK |
| 50 | 01096 | Pà Vầy Sủ | 08 | 83.351060 | 83.351060 | 0.000000 | 0.000000 | 🟢 OK |
| 51 | 00820 | Yên Minh | 08 | 154.682480 | 154.682480 | 0.000000 | 0.000000 | 🟢 OK |
| 52 | 04402 | Phong Dụ Hạ | 15 | 137.083208 | 137.083208 | 0.000000 | 0.000000 | 🟢 OK |
| 53 | 02842 | Tả Củ Tỷ | 15 | 67.160701 | 67.160701 | 0.000000 | 0.000000 | 🟢 OK |
| 54 | 19333 | Đông Hà | 44 | 38.060291 | 38.060291 | 0.000000 | 0.000000 | 🟢 OK |
| 55 | 21040 | Bình Sơn | 51 | 99.744277 | 99.744277 | 0.000000 | 0.000000 | 🟢 OK |
| 56 | 23908 | Ia Tôr | 52 | 102.495933 | 102.495933 | 0.000000 | 0.000000 | 🟢 OK |
| 57 | 30154 | Tân Long | 82 | 95.496497 | 95.496497 | 0.000000 | 0.000000 | 🟢 OK |
| 58 | 15661 | Tân Thành | 38 | 89.701978 | 89.701978 | 0.000000 | 0.000000 | 🟢 OK |
| 59 | 03472 | Mường Mô | 12 | 395.097727 | 395.097727 | 0.000000 | 0.000000 | 🟢 OK |
| 60 | 16177 | Xuân Bình | 38 | 181.696388 | 181.696388 | 0.000000 | 0.000000 | 🟢 OK |
| 61 | 03549 | Phong Thổ | 12 | 266.671969 | 266.671969 | 0.000000 | 0.000000 | 🟢 OK |
| 62 | 01075 | Nậm Dịch | 08 | 97.871128 | 97.871128 | 0.000000 | 0.000000 | 🟢 OK |
| 63 | 03394 | Sin Suối Hồ | 12 | 255.949825 | 255.949825 | 0.000000 | 0.000000 | 🟢 OK |
| 64 | 23332 | Đăk Rơ Wa | 51 | 157.780885 | 157.780885 | 0.000000 | 0.000000 | 🟢 OK |
| 65 | 20257 | Hoà Cường | 48 | 15.546017 | 15.546017 | 0.000000 | 0.000000 | 🟢 OK |
| 66 | 20242 | Hải Châu | 48 | 6.725307 | 6.725307 | 0.000000 | 0.000000 | 🟢 OK |
| 67 | 22759 | Phan Rang | 56 | 9.193517 | 9.193517 | 0.000000 | 0.000000 | 🟢 OK |
| 68 | 22870 | Ninh Phước | 56 | 64.790967 | 64.790967 | 0.000000 | 0.000000 | 🟢 OK |
| 69 | 22504 | Đại Lãnh | 56 | 169.077974 | 169.077974 | 0.000000 | 0.000000 | 🟢 OK |
| 70 | 23602 | An Phú | 52 | 32.240949 | 32.240949 | 0.000000 | 0.000000 | 🟢 OK |
| 71 | 21943 | An Nhơn Nam | 52 | 60.267807 | 60.267807 | 0.000000 | 0.000000 | 🟢 OK |
| 72 | 21925 | An Nhơn Bắc | 52 | 32.081849 | 32.081849 | 0.000000 | 0.000000 | 🟢 OK |
| 73 | 23611 | Gào | 52 | 182.469001 | 182.469001 | 0.000000 | 0.000000 | 🟢 OK |
| 74 | 21985 | Tuy Phước Tây | 52 | 68.786862 | 68.786862 | 0.000000 | 0.000000 | 🟢 OK |
| 75 | 02788 | Bản Lầu | 15 | 125.421961 | 125.421961 | 0.000000 | 0.000000 | 🟢 OK |
| 76 | 24529 | Vụ Bổn | 66 | 109.305101 | 109.305101 | 0.000000 | 0.000000 | 🟢 OK |
| 77 | 12452 | Trần Hưng Đạo | 33 | 9.263134 | 9.263134 | 0.000000 | 0.000000 | 🟢 OK |
| 78 | 11983 | Sơn Nam | 33 | 21.551648 | 21.551648 | 0.000000 | 0.000000 | 🟢 OK |

## Tier 1: Province Guard Check (34 Provinces)

All 34 provinces show **0.000000% area change** — no cascading effects from ward fixes.

| Code | Name | Old Area (km²) | New Area (km²) | Diff % | Status |
|------|------|----------------|----------------|--------|--------|
| 01 | Hà Nội | 3354.9130 | 3354.9130 | 0.000000 | 🟢 OK |
| 04 | Cao Bằng | 6698.4142 | 6698.4142 | 0.000000 | 🟢 OK |
| 08 | Tuyên Quang | 13800.2522 | 13800.2522 | 0.000000 | 🟢 OK |
| 11 | Điện Biên | 9545.3337 | 9545.3337 | 0.000000 | 🟢 OK |
| 12 | Lai Châu | 9071.9246 | 9071.9246 | 0.000000 | 🟢 OK |
| 14 | Sơn La | 14118.7882 | 14118.7882 | 0.000000 | 🟢 OK |
| 15 | Lào Cai | 13245.6944 | 13245.6944 | 0.000000 | 🟢 OK |
| 19 | Thái Nguyên | 8383.2843 | 8383.2843 | 0.000000 | 🟢 OK |
| 20 | Lạng Sơn | 8308.9034 | 8308.9034 | 0.000000 | 🟢 OK |
| 22 | Quảng Ninh | 5641.5397 | 5641.5397 | 0.000000 | 🟢 OK |
| 24 | Bắc Ninh | 4709.1013 | 4709.1013 | 0.000000 | 🟢 OK |
| 25 | Phú Thọ | 9350.8952 | 9350.8952 | 0.000000 | 🟢 OK |
| 31 | Hải Phòng | 3087.0437 | 3087.0437 | 0.000000 | 🟢 OK |
| 33 | Hưng Yên | 2548.8210 | 2548.8210 | 0.000000 | 🟢 OK |
| 37 | Ninh Bình | 3830.1364 | 3830.1364 | 0.000000 | 🟢 OK |
| 38 | Thanh Hoá | 11091.0255 | 11091.0255 | 0.000000 | 🟢 OK |
| 40 | Nghệ An | 16485.8920 | 16485.8920 | 0.000000 | 🟢 OK |
| 42 | Hà Tĩnh | 5962.5957 | 5962.5957 | 0.000000 | 🟢 OK |
| 44 | Quảng Trị | 12675.2444 | 12675.2444 | 0.000000 | 🟢 OK |
| 46 | Huế | 4933.4250 | 4933.4250 | 0.000000 | 🟢 OK |
| 48 | Đà Nẵng | 12186.0386 | 12186.0386 | 0.000000 | 🟢 OK |
| 51 | Quảng Ngãi | 14806.8428 | 14806.8428 | 0.000000 | 🟢 OK |
| 52 | Gia Lai | 21581.0625 | 21581.0625 | 0.000000 | 🟢 OK |
| 56 | Khánh Hoà | 12292.0548 | 12292.0548 | 0.000000 | 🟢 OK |
| 66 | Đắk Lắk | 18086.3321 | 18086.3321 | 0.000000 | 🟢 OK |
| 68 | Lâm Đồng | 24246.2938 | 24246.2938 | 0.000000 | 🟢 OK |
| 75 | Đồng Nai | 12730.8879 | 12730.8879 | 0.000000 | 🟢 OK |
| 79 | Hồ Chí Minh | 6735.7928 | 6735.7928 | 0.000000 | 🟢 OK |
| 80 | Tây Ninh | 8523.7761 | 8523.7761 | 0.000000 | 🟢 OK |
| 82 | Đồng Tháp | 5814.5330 | 5814.5330 | 0.000000 | 🟢 OK |
| 86 | Vĩnh Long | 6190.8484 | 6190.8484 | 0.000000 | 🟢 OK |
| 91 | An Giang | 9837.3665 | 9837.3665 | 0.000000 | 🟢 OK |
| 92 | Cần Thơ | 6333.5639 | 6333.5639 | 0.000000 | 🟢 OK |
| 96 | Cà Mau | 7614.3119 | 7614.3119 | 0.000000 | 🟢 OK |

## Tier 2: Topology Changes

**Expected**: All 78 wards show `ST_Equals = NULL` (old geometries were invalid, cannot compare).
**Expected**: All 78 old records show `old_was_invalid = true`.
**Expected**: All 78 new records show `new_is_invalid = false`.

### Validity Summary

| Check | Result |
|-------|--------|
| Old wards invalid (78 set) | 78/78 ✅ |
| New wards valid (78 set) | 78/78 ✅ |
| New provinces valid (all 34) | 34/34 ✅ |
| `ST_Equals` comparable | 0/78 (all old were invalid — NULL) |

### Point Count Changes (Top 10 by absolute delta)

| ward_code | Name | Old Points | New Points | Delta | Sub-geom Delta |
|-----------|------|------------|------------|-------|----------------|
| 24502 | Ea Phê | 13772 | 12573 | -1199 | 0 |
| 16177 | Xuân Bình | 6165 | 5171 | -994 | +1 |
| 24529 | Vụ Bổn | 11474 | 10803 | -671 | 0 |
| 06565 | Khuất Xá | 3765 | 3411 | -354 | 0 |
| 21985 | Tuy Phước Tây | 2214 | 1956 | -258 | 0 |
| 06607 | Xuân Dương | 3766 | 3516 | -250 | 0 |
| 16186 | Hoá Quỳ | 2029 | 1825 | -204 | -1 |
| 25498 | Thạnh Bình | 5034 | 4831 | -203 | 0 |
| 25510 | Trà Vong | 5359 | 5156 | -203 | 0 |
| 03352 | Thanh An | 2798 | 2657 | -141 | 0 |

**Note**: All point count deltas are negative (vertices removed). This is expected — `ST_MakeValid` removes self-intersection artifacts (duplicate/crossing vertices). The area remains identical because the removed vertices were on the boundary of the polygon, not changing the enclosed area.

### Sub-Geometry Count Changes

| ward_code | Name | Delta | Note |
|-----------|------|-------|------|
| 16177 | Xuân Bình | +1 | Polygon split added |
| 03356 | Sam Mứn | +1 | Polygon split added |
| 05542 | Lam Vỹ | +1 | Polygon split added |
| 15661 | Tân Thành | +1 | Polygon split added |
| 03394 | Sin Suối Hồ | +1 | Polygon split added |
| 00820 | Yên Minh | +1 | Polygon split added |
| 21925 | An Nhơn Bắc | +1 | Polygon split added |
| 31261 | Cờ Đỏ | +1 | Polygon split added |
| 24846 | Lang Biang - Đà Lạt | +1 | Polygon split added |
| 25588 | Hảo Đước | +1 | Polygon split added |
| 04402 | Phong Dụ Hạ | +1 | Polygon split added |
| 12452 | Trần Hưng Đạo | +1 | Polygon split added |
| 19066 | Bắc Gianh | +1 | Polygon split added |
| 00958 | Thượng Sơn | +1 | Polygon split added |
| 03460 | Hua Bum | +1 | Polygon split added |
| 21943 | An Nhơn Nam | +1 | Polygon split added |
| 25843 | Tây Nam | +1 | Polygon split added |
| 16186 | Hoá Quỳ | -1 | Polygon merged |
| 25777 | Dầu Tiếng | -1 | Polygon merged |

### Fix Failures (still invalid after fix)

None — all 78 wards are now valid. ✅

## Tier 3: Data Integrity

| Check | Result |
|-------|--------|
| Total wards (new) | 3,321 |
| Total wards (old) | 3,321 |
| Total provinces (new) | 34 |
| Total provinces (old) | 34 |
| Remaining invalid wards (new) | 0 ✅ |
| Remaining invalid provinces (new) | 0 ✅ |
| Old invalid wards (78 set) | 78 ✅ |
| Ward FK orphans (new) | 0 ✅ |
| Province FK orphans (new) | 0 ✅ |
| NULL geom wards (new) | 0 ✅ |
| NULL geom provinces (new) | 0 ✅ |

## Final Verdict

✅ **SAFE TO USE** — No data corruption detected. All 78 ward fixes are area-preserving (0.000000% change). The `ST_MakeValid` + `ST_CollectionExtract(..., 3)` fix:

1. **Preserved area perfectly** — all 78 wards and 34 provinces show identical geodesic area before and after
2. **Fixed all invalid geometries** — 78/78 wards now valid, 0 remaining invalid
3. **No cascading effects** — all 34 province geometries unchanged
4. **No data loss** — record counts identical (3,321 wards, 34 provinces)
5. **No FK orphans** — all foreign key references intact
6. **Vertex changes expected** — self-intersection artifacts removed (negative point deltas), some polygons split/merged (sub-geometry deltas ±1)

The new GIS dataset (`postgresql_ImportData_gis_2026-07-20__23_14_35.sql`) is safe for production use.

---

## Implementation Notes

### Spec/Plan Deviations

The spec and plan assumed the GIS data used `sapnhap_geojson_objects` table with `geom_wkt` text column. The actual published data uses:
- `gis_provinces` table with `province_code`, `geom` (PostGIS geometry)
- `gis_wards` table with `ward_code`, `geom` (PostGIS geometry)

### Query Adaptations

- Used `geom` directly instead of `ST_GeomFromText(geom_wkt, 4326)`
- Cast `ST_Area()` results to `numeric` for `ROUND()` compatibility
- Used subquery wrapper for `ORDER BY` with column aliases
- Used `CASE WHEN ST_IsValid()` guard for `ST_Equals()` (throws on invalid geometries)