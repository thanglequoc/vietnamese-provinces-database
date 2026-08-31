# Release v5.0.0 — Ward code correction

## Nội dung bản cập nhật

- Sửa mã đơn vị hành chính của **Xã Ba Chẽ**, huyện Ba Chẽ, tỉnh Quảng Ninh:
  - Mã cũ: `06978`
  - Mã mới: `06970`

Không có thêm / xóa / thay đổi bản ghi nào khác trong dữ liệu đơn vị hành chính (administrative_regions, administrative_units, provinces, wards).

### Hướng dẫn cập nhật

Cập nhật bằng cách chạy trực tiếp tệp patch [v5.0.0_patch.sql](v5.0.0_patch.sql) trên cơ sở dữ liệu đã có dữ liệu v4.2.0 trở xuống.

---

_English_

## Release v5.0.0 — Ward code correction

- Correct the administrative code of **Xã Ba Chẽ**, Ba Chẽ district, Quảng Ninh province:
  - Old code: `06978`
  - New code: `06970`

No other administrative-unit rows (administrative_regions, administrative_units, provinces, wards) were added, changed, or removed.

### How to apply

Execute the patch [v5.0.0_patch.sql](v5.0.0_patch.sql) directly on a database previously upgraded from v4.2.0 (or earlier).
