# Release v5.1.0 — Quảng Ninh city promotion & dataset metadata table

## Nội dung bản cập nhật

### 1. Quảng Ninh trở thành thành phố

Theo Nghị quyết **36/2026/QH16** (ban hành 24/08/2026, hiệu lực từ 01/09/2026), tỉnh Quảng Ninh được chuyển thành **Thành phố Quảng Ninh** trực thuộc trung ương. Patch cập nhật bản ghi `provinces.code = '22'`:

| Cột | Trước | Sau |
|-----|-------|-----|
| `full_name` | `Tỉnh Quảng Ninh` | `Thành phố Quảng Ninh` |
| `full_name_en` | `Quang Ninh Province` | `Quang Ninh City` |
| `administrative_unit_id` | `2` (Tỉnh) | `1` (Thành phố trực thuộc trung ương) |

### 2. Bảng metadata bộ dữ liệu

Bổ sung bảng một dòng `vn_provinces_metadata` mô tả phiên bản bộ dữ liệu đang cài đặt, và khởi tạo giá trị cho bản release này:

| dataset_version | latest_decree | generated_at |
|-----------------|---------------|--------------|
| `v5.1.0` | `36/2026/QH16` | `2026-09-12 07:42:08` (UTC) |

Không có thêm / xóa / thay đổi bản ghi nào khác trong dữ liệu đơn vị hành chính (administrative_regions, administrative_units, provinces, wards).

### Hướng dẫn cập nhật

Chạy trực tiếp tệp patch [v5.1.0_patch.sql](v5.1.0_patch.sql) trên cơ sở dữ liệu đã có dữ liệu v5.0.0 (hoặc cũ hơn). Patch dùng `CREATE TABLE IF NOT EXISTS` và ghi đè bảng metadata nên có thể chạy lại an toàn.

> **Lưu ý:** patch này chỉ dành cho **PostgreSQL**. Dữ liệu GIS không nằm trong phạm vi patch.

---

_English_

## Release v5.1.0 — Quảng Ninh city promotion & dataset metadata table

### 1. Quảng Ninh becomes a city

Under Resolution **36/2026/QH16** (issued 24/08/2026, effective 01/09/2026), Quảng Ninh province was promoted to **Quảng Ninh City**, a centrally-governed municipality. The patch updates the `provinces.code = '22'` row:

| Column | Before | After |
|--------|--------|-------|
| `full_name` | `Tỉnh Quảng Ninh` | `Thành phố Quảng Ninh` |
| `full_name_en` | `Quang Ninh Province` | `Quang Ninh City` |
| `administrative_unit_id` | `2` (Province) | `1` (Municipality) |

### 2. Dataset metadata table

Adds the single-row `vn_provinces_metadata` table describing the installed dataset release, and seeds it for this release:

| dataset_version | latest_decree | generated_at |
|-----------------|---------------|--------------|
| `v5.1.0` | `36/2026/QH16` | `2026-09-12 07:42:08` (UTC) |

No other administrative-unit rows (administrative_regions, administrative_units, provinces, wards) were added, changed, or removed.

### How to apply

Execute the patch [v5.1.0_patch.sql](v5.1.0_patch.sql) directly on a database that already contains v5.0.0 (or earlier) data. The patch uses `CREATE TABLE IF NOT EXISTS` and overwrites the metadata table, so it is safe to re-run.

> **Note:** this patch targets **PostgreSQL** only. GIS data is not covered.
