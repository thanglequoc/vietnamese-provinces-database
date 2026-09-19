Bản release v4.2.0 này mở rộng hỗ trợ NoSQL và search engine với bộ dữ liệu Elasticsearch, bổ sung collection GIS cho MongoDB, và tự động kiểm tra/sửa lỗi hình học GIS trong quy trình tạo dữ liệu.

## 🔍 Hỗ trợ Elasticsearch (Mới)

Vietnamese Provinces Database hiện đã cung cấp bộ dữ liệu Elasticsearch, cho phép tìm kiếm full-text, autocomplete, và truy vấn không gian GIS trực tiếp trên Elasticsearch.

### Các index

| Index | Số lượng tài liệu | Mô tả |
|-------|-------------------|-------|
| `provinces` | 34 | Tài liệu tỉnh với wards lồng nhau, search keywords, đơn vị hành chính (không có GIS) |
| `provinces-gis` | 34 | Cấu trúc giống `provinces` + dữ liệu GIS cho cả tỉnh và ward (bounding box + GeoJSON polygons) |

### Đặc điểm chính

- **Full-text search & autocomplete**: Trường `SearchKeywords` được tính toán sẵn (code, tên không dấu, tên tiếng Anh, codeName) cho tìm kiếm nhanh
- **Nested wards**: Mỗi tài liệu tỉnh chứa mảng ward lồng nhau — truy vấn nested để tìm ward trong tỉnh
- **GIS spatial queries**: Index `provinces-gis` hỗ trợ `geo_shape` (geometry) và `geo_point` (center) cho truy vấn không gian
- **NDJSON bulk format**: Tương thích trực tiếp với Elasticsearch Bulk API

📖 Hướng dẫn cài đặt, import, và truy vấn chi tiết xem tại: [`elasticsearch/README.md`](../elasticsearch/README.md)

---

## 🍃 MongoDB GIS Dataset (Mới)

Bổ sung hai collection GIS cho MongoDB, cho phép truy vấn không gian hiệu quả với 2dsphere indexes và GeoJSON geometry.

### Các collection

| Collection | Số lượng tài liệu | Mô tả |
|------------|-------------------|-------|
| `provinces-gis` | 34 | Tài liệu tỉnh với GIS geometry (bounding box + GeoJSON MultiPolygon) |
| `wards-gis` | 3.321 | Tài liệu ward độc lập với GIS geometry + `ProvinceCode` để join |

### Đặc điểm chính

- **Thiết kế tách collection**: Ward GIS được tách thành collection riêng thay vì nhúng vào tài liệu tỉnh — tránh giới hạn 16MB BSON (Hà Nội có 500+ ward) và cho phép truy vấn không gian cấp ward hiệu quả
- **2dsphere indexes**: Index trên `GIS.Geometry` và `GIS.Center` cho cả hai collection, hỗ trợ `$geoIntersects`, `$near`, `$geoWithin`
- **Cross-collection joins**: Trường `ProvinceCode` trên `wards-gis` cho phép `$lookup` join với `provinces-gis`
- **Chunked ward data**: Dữ liệu ward GIS được chia thành 8 phần kèm tệp manifest để import

📖 Hướng dẫn cài đặt, import, và truy vấn chi tiết xem tại: [`mongodb/README.md`](../mongodb/README.md)

---

## 🔧 Kiểm tra & sửa lỗi hình học GIS tự động (Cải thiện)

Dữ liệu GIS từ nguồn upstream đôi khi chứa hình học ward không hợp lệ (self-intersecting polygons). Bản release này tích hợp bước kiểm tra và sửa lỗi tự động vào quy trình tạo dữ liệu.

Sử dụng chuỗi PostGIS `ST_CollectionExtract(ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)), 3)` để sửa các hình học không hợp lệ. Bước này:

- Chạy tự động mỗi lần tạo lại bộ dữ liệu
- Ghi log kiểm toán với chi tiết các ward đã sửa
- Idempotent — chạy lại không tạo thay đổi nào nếu không có hình học lỗi

Tất cả hình học GIS được đảm bảo hợp lệ trước khi xuất ra các định dạng PostgreSQL, MySQL, MSSQL, GeoJSON, Elasticsearch, và MongoDB.

---

## 📚 Cập nhật tài liệu

- Cập nhật tài liệu GIS với thông tin định dạng mới
- Dọn dẹp README (loại bỏ phần Q&A lỗi thời)
- Thêm hướng dẫn truy vấn Elasticsearch và MongoDB

---

**English version**:

This v4.2.0 release expands NoSQL and search engine support with an Elasticsearch dataset, adds GIS collections for MongoDB, and integrates automatic GIS geometry validation/fix into the data generation pipeline.

## 🔍 Elasticsearch Support (New)

Vietnamese Provinces Database now provides an Elasticsearch dataset, enabling full-text search, autocomplete, and GIS spatial queries directly on Elasticsearch.

### Indices

| Index | Documents | Description |
|-------|-----------|-------------|
| `provinces` | 34 | Province documents with embedded wards, search keywords, administrative unit data (no GIS) |
| `provinces-gis` | 34 | Same structure plus GIS geometry for both provinces and wards (bounding boxes + GeoJSON polygons) |

### Key Features

- **Full-text search & autocomplete**: Pre-computed `SearchKeywords` field (code, tone-stripped name, English name, codeName) for fast lookups
- **Nested wards**: Each province document contains a nested ward array — use nested queries to find wards within a province
- **GIS spatial queries**: The `provinces-gis` index supports `geo_shape` (geometry) and `geo_point` (center) for spatial queries
- **NDJSON bulk format**: Directly compatible with the Elasticsearch Bulk API

📖 For detailed setup, import, and query instructions, see: [`elasticsearch/README.md`](../elasticsearch/README.md)

---

## 🍃 MongoDB GIS Dataset (New)

Adds two GIS collections for MongoDB, enabling efficient spatial queries with 2dsphere indexes and GeoJSON geometry.

### Collections

| Collection | Documents | Description |
|------------|-----------|-------------|
| `provinces-gis` | 34 | Province documents with GIS geometry (bounding boxes + GeoJSON MultiPolygon) |
| `wards-gis` | 3,321 | Standalone ward documents with GIS geometry + `ProvinceCode` for joins |

### Key Features

- **Separate collection design**: Ward GIS is split into a dedicated collection rather than embedded in province documents — avoids the 16MB BSON limit (Hà Nội has 500+ wards) and enables efficient ward-level spatial queries
- **2dsphere indexes**: Indexes on `GIS.Geometry` and `GIS.Center` for both collections, supporting `$geoIntersects`, `$near`, `$geoWithin`
- **Cross-collection joins**: The `ProvinceCode` field on `wards-gis` enables `$lookup` joins with `provinces-gis`
- **Chunked ward data**: Ward GIS data is split into 8 parts with a manifest file for import

📖 For detailed setup, import, and query instructions, see: [`mongodb/README.md`](../mongodb/README.md)

---

## 🔧 Automatic GIS Geometry Validation & Fix (Improvement)

GIS data from the upstream source occasionally contains invalid ward geometries (self-intersecting polygons). This release integrates an automatic validation and fix step into the data generation pipeline.

Uses the PostGIS chain `ST_CollectionExtract(ST_MakeValid(ST_GeomFromText(geom_wkt, 4326)), 3)` to repair invalid geometries. This step:

- Runs automatically every time the dataset is regenerated
- Writes an audit log with details of every fixed ward
- Is idempotent — re-running produces zero changes if no invalid geometries exist

All GIS geometries are guaranteed valid before being exported to PostgreSQL, MySQL, MSSQL, GeoJSON, Elasticsearch, and MongoDB formats.

---

## 📚 Documentation Updates

- Updated GIS documentation with new format information
- Cleaned up README (removed outdated Q&A sections)
- Added Elasticsearch and MongoDB query guides