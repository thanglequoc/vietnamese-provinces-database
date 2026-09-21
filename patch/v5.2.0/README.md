# Release v5.2.0 — Bắc Ninh trở thành thành phố & thành lập 12 phường

## Nội dung bản cập nhật

### 1. Bắc Ninh trở thành thành phố

Theo Nghị quyết **39/2026/QH16**, tỉnh Bắc Ninh được chuyển thành **Thành phố Bắc Ninh** trực thuộc trung ương. Patch cập nhật bản ghi `provinces.code = '24'`:

| Cột | Trước | Sau |
|-----|-------|-----|
| `full_name` | `Tỉnh Bắc Ninh` | `Thành phố Bắc Ninh` |
| `full_name_en` | `Bac Ninh Province` | `Bac Ninh City` |
| `administrative_unit_id` | `2` (Tỉnh) | `1` (Thành phố trực thuộc trung ương) |

### 2. Thành lập 12 phường thuộc tỉnh Bắc Ninh

Theo Nghị quyết **388/NQ-UBTVQH16**, 12 xã thuộc tỉnh Bắc Ninh được thành lập thành **phường**. Patch cập nhật `full_name`, `full_name_en` và `administrative_unit_id` (`4` Xã → `3` Phường); các cột `code`, `name`, `name_en`, `code_name`, `province_code`, `postal_code` không đổi.

| Mã | Tên | Trước | Sau |
|----|-----|-------|-----|
| `07294` | Bố Hạ | `Xã Bố Hạ` | `Phường Bố Hạ` |
| `07375` | Lạng Giang | `Xã Lạng Giang` | `Phường Lạng Giang` |
| `07399` | Kép | `Xã Kép` | `Phường Kép` |
| `07444` | Lục Nam | `Xã Lục Nam` | `Phường Lục Nam` |
| `07840` | Hiệp Hoà | `Xã Hiệp Hoà` | `Phường Hiệp Hoà` |
| `09193` | Yên Phong | `Xã Yên Phong` | `Phường Yên Phong` |
| `09292` | Phù Lãng | `Xã Phù Lãng` | `Phường Phù Lãng` |
| `09313` | Chi Lăng | `Xã Chi Lăng` | `Phường Chi Lăng` |
| `09319` | Tiên Du | `Xã Tiên Du` | `Phường Tiên Du` |
| `09454` | Gia Bình | `Xã Gia Bình` | `Phường Gia Bình` |
| `09475` | Nhân Thắng | `Xã Nhân Thắng` | `Phường Nhân Thắng` |
| `09496` | Lương Tài | `Xã Lương Tài` | `Phường Lương Tài` |

### 3. Bảng metadata bộ dữ liệu

Cập nhật bảng một dòng `vn_provinces_metadata` mô tả phiên bản bộ dữ liệu đang cài đặt:

| dataset_version | latest_decree | generated_at |
|-----------------|---------------|--------------|
| `v5.2.0` | `388/NQ-UBTVQH16` | `2026-09-20 13:13:17` (UTC) |

Không có thêm / xóa bản ghi nào khác trong dữ liệu đơn vị hành chính (administrative_regions, administrative_units, provinces, wards).

### Hướng dẫn cập nhật

Chạy trực tiếp tệp patch [v5.2.0_patch.sql](v5.2.0_patch.sql) trên cơ sở dữ liệu đã có dữ liệu v5.1.0 (hoặc cũ hơn). Patch dùng `UPDATE` theo khóa chính và ghi đè bảng metadata nên có thể chạy lại an toàn.

> **Lưu ý:** patch này chỉ dành cho **PostgreSQL**. Dữ liệu GIS không nằm trong phạm vi patch.

---

_English_

## Release v5.2.0 — Bắc Ninh becomes a city & 12 wards established

### 1. Bắc Ninh becomes a city

Under Resolution **39/2026/QH16**, Bắc Ninh province was promoted to **Bắc Ninh City**, a centrally-governed municipality. The patch updates the `provinces.code = '24'` row:

| Column | Before | After |
|--------|--------|-------|
| `full_name` | `Tỉnh Bắc Ninh` | `Thành phố Bắc Ninh` |
| `full_name_en` | `Bac Ninh Province` | `Bac Ninh City` |
| `administrative_unit_id` | `2` (Province) | `1` (Municipality) |

### 2. Establishment of 12 wards in Bắc Ninh province

Under Resolution **388/NQ-UBTVQH16**, 12 communes of Bắc Ninh province were established as **wards (phường)**. The patch updates `full_name`, `full_name_en` and `administrative_unit_id` (`4` Commune → `3` Ward); the `code`, `name`, `name_en`, `code_name`, `province_code` and `postal_code` columns are unchanged.

| Code | Name | Before | After |
|------|------|--------|-------|
| `07294` | Bố Hạ | `Xã Bố Hạ` | `Phường Bố Hạ` |
| `07375` | Lạng Giang | `Xã Lạng Giang` | `Phường Lạng Giang` |
| `07399` | Kép | `Xã Kép` | `Phường Kép` |
| `07444` | Lục Nam | `Xã Lục Nam` | `Phường Lục Nam` |
| `07840` | Hiệp Hoà | `Xã Hiệp Hoà` | `Phường Hiệp Hoà` |
| `09193` | Yên Phong | `Xã Yên Phong` | `Phường Yên Phong` |
| `09292` | Phù Lãng | `Xã Phù Lãng` | `Phường Phù Lãng` |
| `09313` | Chi Lăng | `Xã Chi Lăng` | `Phường Chi Lăng` |
| `09319` | Tiên Du | `Xã Tiên Du` | `Phường Tiên Du` |
| `09454` | Gia Bình | `Xã Gia Bình` | `Phường Gia Bình` |
| `09475` | Nhân Thắng | `Xã Nhân Thắng` | `Phường Nhân Thắng` |
| `09496` | Lương Tài | `Xã Lương Tài` | `Phường Lương Tài` |

### 3. Dataset metadata table

Updates the single-row `vn_provinces_metadata` table describing the installed dataset release:

| dataset_version | latest_decree | generated_at |
|-----------------|---------------|--------------|
| `v5.2.0` | `388/NQ-UBTVQH16` | `2026-09-20 13:13:17` (UTC) |

No other administrative-unit rows (administrative_regions, administrative_units, provinces, wards) were added or removed.

### How to apply

Execute the patch [v5.2.0_patch.sql](v5.2.0_patch.sql) directly on a database that already contains v5.1.0 (or earlier) data. The patch uses primary-key `UPDATE`s and overwrites the metadata table, so it is safe to re-run.

> **Note:** this patch targets **PostgreSQL** only. GIS data is not covered.
