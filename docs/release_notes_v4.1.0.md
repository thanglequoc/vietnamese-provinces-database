Bản release v4.1.0 này mang đến bộ dữ liệu GeoJSON hoàn chỉnh, vá dữ liệu GIS cho các lãnh thổ đảo, và chuẩn hóa tên tiếng Anh cho các đơn vị hành chính có tên gọi quốc tế.

## 🗺️ Xuất dữ liệu GeoJSON (Mới)

Vietnamese Provinces Database hiện đã cung cấp bộ dữ liệu GeoJSON hoàn chỉnh, cho phép sử dụng trực tiếp trong các ứng dụng bản đồ, client-side, và quy trình GIS dựa trên tệp.

### Phạm vi dữ liệu

| Cấp hành chính | Số lượng tệp |
|---------------|--------------|
| Tỉnh / Thành phố | 34 |
| Xã / Phường | 3.321 |
| **Tổng cộng** | **3.355** |

### Cấu trúc thư mục

Mỗi tỉnh/thành phố được xuất thành một thư mục riêng biệt:

```
json/geojson/
  {province_code}_{province_code_name}/
    {province_code}_{province_code_name}.geojson
    wards/
      {ward_code}_{ward_code_name}.geojson
```

### Định dạng GeoJSON

Mỗi tệp `.geojson` là một `FeatureCollection` hợp lệ bao gồm:

- `bbox` ở cấp FeatureCollection
- Đúng một `Feature` cho mỗi tệp
- `id` của Feature được đặt thành mã đơn vị hành chính
- `bbox` ở cấp Feature
- Dữ liệu hình học GeoJSON
- Thuộc tính (properties) với camelCase: `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `gisServerId`, `areaKm2`

### Tải xuống

Bộ dữ liệu GeoJSON cũng được đóng gói thành tệp zip để tải xuống hàng loạt:

```
vn_provinces_wards_geojson_<datetime>.zip
```

### Xem trước

Mở [https://geojson.io](https://geojson.io) và tải bất kỳ tệp `.geojson` nào từ thư mục `json/geojson/` để xem hình học và thuộc tính.

---

## 🏝️ Vá dữ liệu GIS cho đảo Hoàng Sa & Trường Sa (Sửa lỗi)

Dữ liệu GIS cấp tỉnh từ nguồn upstream `sapnhap.bando.com.vn` bị thiếu các lãnh thổ đảo. Bản release này vá lỗi bằng cách hợp nhất hình học cấp xã của các đảo vào hình học cấp tỉnh mẹ.

### Các tỉnh được vá

| Tỉnh | Mã | Đảo được hợp nhất | Mã xã | Trạng thái trước | Trạng thái sau |
|------|-----|-------------------|--------|-------------------|----------------|
| Đà Nẵng | 48 | Hoàng Sa | 20333 | `ST_Contains = false` | `ST_Contains = true` |
| Khánh Hòa | 56 | Trường Sa | 22736 | `ST_Contains = false` | `ST_Contains = true` |

### Cách tiếp cận

Sử dụng PostGIS `ST_Union` để hợp nhất hình học đảo vào hình học tỉnh mẹ, và `ST_Envelope` để tính lại bounding box. Việc vá được tích hợp vào quy trình tạo dữ liệu, đảm bảo sửa lỗi được áp dụng tự động mỗi lần tạo lại bộ dữ liệu.

### Kết quả

Hình học cấp tỉnh hiện đã bao gồm đầy đủ các đơn vị hành chính cấp dưới, bao gồm quần đảo Hoàng Sa và Trường Sa.

---

## 🇬🇧 Vá tên tiếng Anh

Bổ sung tên tiếng Anh chuẩn cho các đơn vị hành chính có tên gọi quốc tế được công nhận, thay thế cho tên tiếng Anh tự động sinh ra (bằng cách bỏ dấu thanh).

### Các đơn vị được vá

| Mã | Tên tiếng Việt | Tên cũ (En) | Tên mới (En) |
|-----|---------------|-------------|--------------|
| 01 | Hà Nội | Ha Noi | **Hanoi** |
| 31 | Hải Phòng | Hai Phong | **Haiphong** |
| 20333 | Hoàng Sa | Hoang Sa | **Paracel** |
| 22736 | Trường Sa | Truong Sa | **Spratly** |

### Phạm vi áp dụng

Các bản vá tên được áp dụng đồng bộ trên **tất cả** các định dạng cơ sở dữ liệu:

- PostgreSQL / MySQL
- Microsoft SQL Server
- Oracle
- JSON (đầy đủ & rút gọn)
- MongoDB
- Redis

---

## 📚 Cải thiện tài liệu & dọn dẹp

- Viết lại README thành hướng dẫn cài đặt hoàn chỉnh với điều kiện tiên quyết, các bước kiểm tra, và xử lý sự cố
- Cập nhật tài liệu GIS với URL nguồn dữ liệu mới
- Thêm tài liệu ví dụ truy vấn GIS
- Dọn dẹp mã nguồn chết (loại bỏ công cụ `compare-gis` và `manual_seed_dumper` không còn sử dụng)
- Thêm `.env` vào `.gitignore` để tránh vô tình commit thông tin xác thực

---

**English version**:

This v4.1.0 release brings a complete GeoJSON dataset, patches GIS data for island territories, and standardizes English names for administrative units with internationally recognized names.

## 🗺️ GeoJSON Export (New)

Vietnamese Provinces Database now provides a complete GeoJSON dataset, enabling direct use in map applications, client-side apps, and file-based GIS workflows.

### Coverage

| Administrative Level | File Count |
|----------------------|------------|
| Provinces / Municipalities | 34 |
| Wards / Communes | 3,321 |
| **Total** | **3,355** |

### Folder Structure

Each province/municipality is exported to its own directory:

```
json/geojson/
  {province_code}_{province_code_name}/
    {province_code}_{province_code_name}.geojson
    wards/
      {ward_code}_{ward_code_name}.geojson
```

### GeoJSON Format

Each `.geojson` file is a valid `FeatureCollection` containing:

- Top-level `bbox`
- Exactly one `Feature` per file
- Feature `id` set to the administrative unit code
- Feature-level `bbox`
- GeoJSON geometry data
- Properties in camelCase: `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `gisServerId`, `areaKm2`

### Download

The GeoJSON dataset is also packaged as a zip archive for bulk download:

```
vn_provinces_wards_geojson_<datetime>.zip
```

### Preview

Open [https://geojson.io](https://geojson.io) and load any `.geojson` file from the `json/geojson/` directory to inspect the geometry and properties.

---

## 🏝️ Patch GIS for Paracel & Spratly Islands (Fix)

The province-level GIS data from the upstream source `sapnhap.bando.com.vn` was missing island territories. This release patches the defect by merging ward-level island geometries into their parent province geometries.

### Patched Provinces

| Province | Code | Merged Island | Ward Code | Before | After |
|----------|------|---------------|-----------|--------|-------|
| Da Nang | 48 | Hoàng Sa (Paracel) | 20333 | `ST_Contains = false` | `ST_Contains = true` |
| Khanh Hoa | 56 | Trường Sa (Spratly) | 22736 | `ST_Contains = false` | `ST_Contains = true` |

### Approach

Uses PostGIS `ST_Union` to merge island geometry into the parent province geometry, and `ST_Envelope` to recalculate the bounding box. The patch is integrated into the generation flow, ensuring the fix is applied automatically every time the dataset is regenerated.

### Result

Province geometries now spatially contain all their administrative subdivisions, including the Paracel and Spratly Islands.

---

## 🇬🇧 English Name Patches

Added standard English names for administrative units with internationally recognized names, replacing the auto-generated English names (produced by stripping Vietnamese tone marks).

### Patched Units

| Code | Vietnamese Name | Old (En) | New (En) |
|------|-----------------|----------|----------|
| 01 | Hà Nội | Ha Noi | **Hanoi** |
| 31 | Hải Phòng | Hai Phong | **Haiphong** |
| 20333 | Hoàng Sa | Hoang Sa | **Paracel** |
| 22736 | Trường Sa | Truong Sa | **Spratly** |

### Scope

The name patches are applied consistently across **all** database formats:

- PostgreSQL / MySQL
- Microsoft SQL Server
- Oracle
- JSON (full & simplified)
- MongoDB
- Redis

---

## 📚 Documentation & Cleanup

- Rewrote README as a complete setup guide with prerequisites, verification steps, and troubleshooting
- Updated GIS documentation with new data source URL
- Added GIS example query documentation
- Cleaned up dead code (removed unused `compare-gis` tool and `manual_seed_dumper`)
- Added `.env` to `.gitignore` to prevent accidental credential commits