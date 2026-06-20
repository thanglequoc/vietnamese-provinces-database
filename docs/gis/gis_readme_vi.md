# Phần mở rộng Dữ liệu GIS — Vietnamese Provinces Database
**Dữ liệu GIS Dataset mới nhất: **v4.0.0** (June 20, 2026)**

|Nền tảng|Link tải|Kích thước|
|--------|-------------|---------|
|PostgreSQL/PostGIS|[Tải về file SQL GIS Dataset cho PostgreSQL][gis_dataset_postgresql_bucket_url]|~152.07 MB|
|MySQL|[Tải về file SQL GIS Dataset cho MySQL][gis_dataset_mysql_bucket_url]|~150.44 MB|
|Microsoft SQL Server|[Tải về file SQL GIS Dataset cho SQL Server][gis_dataset_sqlserver_bucket_url]|~152.13 MB|

## Mục lục

1. [Giới thiệu](#gioi-thieu)
2. [Phạm vi dữ liệu](#pham-vi-du-lieu)
3. [Cài đặt](#cai-dat)
4. [Lược đồ cơ sở dữ liệu](#luoc-do-co-so-du-lieu)
5. [Mẹo tối ưu truy vấn](#meo-toi-uu-truy-van)
6. [Khả năng tương thích phiên bản](#kha-nang-tuong-thich-phien-ban)
7. [Câu hỏi thường gặp & Khắc phục sự cố](#cau-hoi-thuong-gap--khac-phuc-su-co)

---

## Giới thiệu

**Bộ dữ liệu GIS** là một phần mở rộng tùy chọn của dự án Vietnamese Provinces Database, cung cấp dữ liệu hình học ranh giới hành chính có độ chính xác cao cho các đơn vị hành chính của Việt Nam. Bộ dữ liệu này chứa thông tin ranh giới địa lý chi tiết của toàn bộ 34 tỉnh/thành phố và 3.321 xã/phường, hỗ trợ các truy vấn không gian (geospatial queries) và trực quan hóa bản đồ một cách hiệu quả.

### Bộ dữ liệu GIS là gì?

Bộ dữ liệu GIS bao gồm:

- **Hình học ranh giới địa lý (Geographic Boundary Geometries)** được lưu trữ dưới dạng dữ liệu không gian (`Multipolygon`)
- **Hộp giới hạn (Bounding Box)** cho từng đơn vị hành chính (hữu ích cho việc lọc sơ bộ truy vấn)
- **Mã đơn vị hành chính được đồng bộ hóa** với Vietnamese Provinces Database chính

### Tại sao nên sử dụng?

Bộ dữ liệu GIS hỗ trợ:

* **Trực quan hóa bản đồ** – Hiển thị ranh giới tỉnh/thành phố và xã/phường trên các ứng dụng bản đồ web, di động và desktop.
* **Tra cứu ranh giới hành chính** – Xác định đơn vị hành chính tương ứng với một vị trí địa lý cụ thể.
* **Truy vấn Point-in-Polygon** – Xác định tỉnh/thành phố hoặc xã/phường chứa một tọa độ kinh độ/vĩ độ nhất định.
* **Phân tích không gian** – Thực hiện các phép toán địa lý như tính khoảng cách, giao cắt, diện tích và nhiều thao tác không gian khác trực tiếp trong cơ sở dữ liệu.

---

## Phạm vi dữ liệu

### Phạm vi địa lý

Bộ dữ liệu cung cấp đầy đủ ranh giới cho toàn bộ các đơn vị hành chính của Việt Nam:

| Cấp hành chính | Số lượng | Phạm vi bao phủ |
|---------------|----------|----------------|
| **Tỉnh/Thành phố** | 34 | 100% (bao gồm Hà Nội, TP. Hồ Chí Minh và Đà Nẵng) |
| **Xã/Phường** | 3.321 | 100% (toàn bộ xã, phường và đặc khu hành chính) |

### Hệ quy chiếu tọa độ

- **Tiêu chuẩn:** WGS 84 (World Geodetic System 1984)
- **SRID:** 4326
- **Định dạng:** Kinh độ, vĩ độ (theo tiêu chuẩn OGC)
- **Datum:** Chuẩn GPS toàn cầu

### Các loại hình học (Geometry Types)

Mỗi đơn vị hành chính được biểu diễn bằng hai đối tượng hình học:

1. **Bounding Box (`bbox`)**
- Loại: `Polygon`
- Mục đích: Lọc sơ bộ nhanh cho các truy vấn không gian
- Hiệu năng: Tính toán nhanh hơn so với việc sử dụng toàn bộ hình học ranh giới

2. **Ranh giới đầy đủ (`geom`)**
- Loại: `Multipolygon`
- Mục đích: Hiển thị ranh giới chính xác và thực hiện các phép toán Point-in-Polygon
- Phạm vi: Bao phủ toàn bộ khu vực hành chính, bao gồm cả các vùng không liền kề (nếu có)

### Các bảng dữ liệu chính

Bộ dữ liệu bao gồm hai bảng chính:
1. **`gis_provinces`** — 34 bản ghi cấp tỉnh/thành phố chứa dữ liệu ranh giới địa lý
2. **`gis_wards`** — 3.321 bản ghi cấp xã/phường chứa dữ liệu ranh giới địa lý

### Mối quan hệ với cơ sở dữ liệu chính

Các đơn vị hành chính trong bộ dữ liệu GIS được đồng bộ với Vietnamese Provinces Database thông qua mã đơn vị hành chính:

- **Tỉnh/Thành phố:** Trường `gis_provinces.province_code` liên kết với `provinces.code`
- **Xã/Phường:** Trường `gis_wards.ward_code` liên kết với `wards.code`

Điều này cho phép kết hợp (join) dữ liệu ranh giới GIS với thông tin đơn vị hành chính, tên gọi và quan hệ phân cấp một cách liền mạch.
[![image.png](https://i.postimg.cc/zBJQwyY1/image.png)](https://postimg.cc/nsPTpcLd)

---

## Cài đặt

Bộ dữ liệu GIS yêu cầu các tập lệnh SQL khởi tạo (bootstrap scripts) để tạo bảng và các tập lệnh nhập dữ liệu (import scripts) để nạp dữ liệu ranh giới. Các tập lệnh hiện được cung cấp cho PostgreSQL, MySQL và SQL Server.

### PostgreSQL + PostGIS

#### Yêu cầu

- **PostgreSQL:** Phiên bản 12 trở lên
- **PostGIS:** Phiên bản 3.0 trở lên (bắt buộc để hỗ trợ kiểu dữ liệu không gian)
- **Dung lượng ổ đĩa:** Khoảng 100 MB cho toàn bộ dữ liệu hình học
- **Cơ sở dữ liệu hiện có:** Cần cài đặt Vietnamese Provinces Database trước

#### Bước 1: Kích hoạt PostGIS Extension

Kết nối tới cơ sở dữ liệu PostgreSQL và kích hoạt PostGIS:

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
-- Kiểm tra phiên bản
SELECT postgis_version();
```

Kết quả mong đợi: Thông tin phiên bản PostGIS (ví dụ: `3.1.4`).

#### Bước 2: Tạo các bảng GIS (Bootstrap)

Thực thi tập lệnh bootstrap để tạo toàn bộ bảng, chỉ mục và ràng buộc:

```bash
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_CreateGISTables.sql
```

#### Bước 3: Nhập dữ liệu GIS

Giải nén gói dữ liệu GIS PostgreSQL tại thư mục [postgresql/gis](../../postgresql/gis/)

Hoặc tải trực tiếp tập dữ liệu GIS từ [vn-province bucket][gis_dataset_postgresql_bucket_url]

Thực thi tập lệnh nhập dữ liệu để nạp dữ liệu ranh giới:

```bash
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_ImportData_gis_2026-06-20__12_32_01.sql
```

### MySQL / MariaDB

#### Yêu cầu

- **MySQL:** Phiên bản 8.0 trở lên
- **MariaDB:** Phiên bản 10.4 trở lên (có hỗ trợ dữ liệu không gian)
- **Dung lượng ổ đĩa:** Khoảng 120 MB cho toàn bộ dữ liệu hình học
- **Cơ sở dữ liệu hiện có:** Cần cài đặt Vietnamese Provinces Database trước

#### Bước 1: Tạo các bảng GIS (Bootstrap)

Thực thi tập lệnh bootstrap để tạo toàn bộ bảng, chỉ mục và ràng buộc:

```bash
mysql -u <username> -p <database_name> < mysql/gis/mysql_CreateGISTables.sql
```

#### Bước 2: Nhập dữ liệu GIS

Giải nén gói dữ liệu GIS MySQL tại thư mục [mysql/gis](../../mysql/gis/)

Hoặc tải trực tiếp tập dữ liệu GIS từ [vn-province bucket][gis_dataset_mysql_bucket_url]

Thực thi tập lệnh nhập dữ liệu:

```bash
mysql -u <username> -p <database_name> < mysql/gis/mysql_ImportData_gis_2026-06-20__12_32_01.sql
```

#### Lưu ý riêng cho MySQL

- MySQL sử dụng **SPATIAL INDEX** (hoặc hiển thị dưới tên **SPATIAL KEY**) cho các cột hình học thay vì chỉ mục GiST như PostgreSQL.
- Dữ liệu hình học được lưu bằng các kiểu dữ liệu không gian gốc như `MULTIPOLYGON` với SRID 4326.
- Các hàm không gian tuân theo quy ước đặt tên `ST_*` của OGC (ví dụ: `ST_Contains`, `ST_Intersects`, và `ST_Distance`).
- Xem mục [Coordinate Handling Across Databases](#coordinate-handling-across-databases) để biết thêm chi tiết về thứ tự tọa độ và cách xử lý SRID.

---

### Microsoft SQL Server

#### Yêu cầu

- **SQL Server:** Phiên bản 2019 trở lên
- **Dung lượng ổ đĩa:** Khoảng 110 MB cho toàn bộ dữ liệu hình học
- **Cơ sở dữ liệu hiện có:** Cần cài đặt Vietnamese Provinces Database trước

#### Bước 1: Tạo các bảng GIS (Bootstrap)

Thực thi tập lệnh bootstrap để tạo toàn bộ bảng, chỉ mục và ràng buộc:

```cmd
sqlcmd -S <server_name> -d <database_name> -U <username> -P <password> -i sqlserver/gis/mssql_CreateGISTables.sql
```

#### Bước 2: Nhập dữ liệu GIS

Giải nén gói dữ liệu GIS SQL Server tại thư mục [sqlserver/gis](../../sqlserver/gis/)

Hoặc tải trực tiếp tập dữ liệu GIS từ [vn-province bucket][gis_dataset_sqlserver_bucket_url]

Thực thi tập lệnh nhập dữ liệu:

```cmd
sqlcmd -S <server_name> -d <database_name> -U <username> -P <password> -i sqlserver/gis/mssql_ImportData_gis_2026-06-20__12_32_02.sql
```

#### Lưu ý riêng cho SQL Server

- Dữ liệu hình học được lưu bằng kiểu dữ liệu không gian gốc `geometry` của SQL Server.
- Các đối tượng hình học thường được tạo bằng các hàm như `geometry::STGeomFromText()` và `geometry::Point()`.
- Các phép toán không gian sử dụng cú pháp phương thức của đối tượng SQL Server (ví dụ: `geom.STContains()` và `geom.STDistance()`).
- Chỉ mục không gian được triển khai bằng tính năng **Spatial Index** tích hợp sẵn của SQL Server.

---

## Lược đồ cơ sở dữ liệu

### gis_provinces

Lưu trữ dữ liệu ranh giới GIS của tất cả các đơn vị hành chính cấp tỉnh/thành phố.

| Cột | Kiểu dữ liệu | Cho phép NULL | Mô tả |
|------|------|------|------|
| `id` | `INTEGER` | Không | Khóa chính thay thế (surrogate primary key) của bản ghi GIS. |
| `province_code` | `VARCHAR(20)` | Không | Mã tỉnh/thành phố dùng để liên kết với bảng `provinces`. |
| `gis_server_id` | `VARCHAR(50)` | Có | Mã định danh gốc của dữ liệu ranh giới từ nguồn GIS. |
| `area_km2` | `NUMERIC(12,5)` | Có | Diện tích tỉnh/thành phố (km²). |
| `bbox` | `geometry(Polygon, 4326)` | Có | Hình học Bounding Box dùng cho lọc sơ bộ và tối ưu hóa chỉ mục không gian. |
| `geom` | `geometry(MultiPolygon, 4326)` | Có | Hình học ranh giới đầy đủ dùng cho trực quan hóa và phân tích không gian. |

**Khóa chính:** `id`  
**Khóa ngoại:** `province_code` → `provinces.code`

---

### gis_wards

Lưu trữ dữ liệu ranh giới GIS của tất cả các đơn vị hành chính cấp xã/phường.

| Cột | Kiểu dữ liệu | Cho phép NULL | Mô tả |
|------|------|------|------|
| `id` | `INTEGER` | Không | Khóa chính thay thế (surrogate primary key) của bản ghi GIS. |
| `ward_code` | `VARCHAR(20)` | Không | Mã xã/phường dùng để liên kết với bảng `wards`. |
| `gis_server_id` | `VARCHAR(50)` | Có | Mã định danh gốc của dữ liệu ranh giới từ nguồn GIS. |
| `area_km2` | `NUMERIC(12,5)` | Có | Diện tích đơn vị hành chính (km²). |
| `bbox` | `geometry(Polygon, 4326)` | Có | Hình học Bounding Box dùng cho lọc sơ bộ và tối ưu hóa chỉ mục không gian. |
| `geom` | `geometry(MultiPolygon, 4326)` | Có | Hình học ranh giới đầy đủ dùng cho trực quan hóa và phân tích không gian. |

**Khóa chính:** `id`  
**Khóa ngoại:** `ward_code` → `wards.code`

---

## Mẹo tối ưu truy vấn

#### Sử dụng Bounding Box để lọc sơ bộ trong các truy vấn không gian lớn

Bộ dữ liệu lưu trữ cả hình học ranh giới đầy đủ (`geom`) và Bounding Box đơn giản hóa (`bbox`).

Đối với các truy vấn trên phạm vi rộng, việc lọc trước bằng `bbox` có thể giúp giảm đáng kể số lượng hình học cần thực hiện các phép tính không gian tốn kém.

```sql
-- Ví dụ: Tìm xã/phường chứa Trường đua Phú Thọ
-- (tọa độ: 106.65735539347402, 10.767883941562623)
SELECT *
FROM gis_wards
WHERE bbox && ST_SetSRID(
        ST_Point(106.65735539347402, 10.767883941562623),
        4326
      )
AND ST_Contains(
        geom,
        ST_SetSRID(
            ST_Point(106.65735539347402, 10.767883941562623),
            4326
        )
    );
```

---

## Khả năng tương thích phiên bản

Bộ dữ liệu GIS đã được kiểm thử và hỗ trợ trên các nền tảng sau:

| Thành phần | Phiên bản | Trạng thái | Ghi chú |
|------------|------------|------------|------------|
| **PostgreSQL** | 12+ | ✓ Hỗ trợ đầy đủ | Khuyến nghị: 14+ để có hiệu năng tốt hơn |
| **PostGIS** | 3.0+ | ✓ Bắt buộc | Cần thiết cho các kiểu dữ liệu và hàm không gian |
| **MySQL** | 8.0+ | ✓ Hỗ trợ đầy đủ | Hỗ trợ cả MariaDB 10.4+ |
| **MariaDB** | 10.4+ | ✓ Hỗ trợ đầy đủ | Chức năng tương đương MySQL 8.0 |
| **SQL Server** | 2019+ | ✓ Hỗ trợ đầy đủ | Tích hợp sẵn kiểu dữ liệu không gian |
| **WGS 84 (SRID 4326)** | ISO 19115 | ✓ Tiêu chuẩn | Được hỗ trợ trên tất cả nền tảng |

### Yêu cầu tối thiểu

- **PostgreSQL:** 12 (phát hành năm 2019)
- **PostGIS:** 3.0 (phát hành năm 2019)
- **MySQL:** 8.0 (phát hành năm 2018)
- **SQL Server:** 2019 (phát hành năm 2019)

### Phiên bản khuyến nghị để đạt hiệu năng tối ưu

- **PostgreSQL:** 15+ (ổn định, hỗ trợ xử lý song song tốt hơn)
- **PostGIS:** 3.3+ (cải thiện hiệu năng)
- **MySQL:** 8.0.30+ (các bản phát hành mới nhất của nhánh 8.0)
- **SQL Server:** 2022 (phiên bản mới nhất)

---

## Câu hỏi thường gặp & Khắc phục sự cố

### Hỏi: Tôi có cần cài đặt bộ dữ liệu GIS không?

**Trả lời:** Chỉ khi bạn cần dữ liệu ranh giới địa lý (hiển thị bản đồ, truy vấn không gian, geocoding...). Vietnamese Provinces Database vẫn hoạt động độc lập cho các nhu cầu tra cứu đơn vị hành chính, quan hệ phân cấp và dữ liệu mô tả.

### Hỏi: Tôi gặp lỗi "PostGIS extension not found" thì phải làm sao?

**Trả lời:** PostGIS chưa được cài đặt. Đối với PostgreSQL:

```sql
CREATE EXTENSION postgis;
-- hoặc
sudo apt-get install postgresql-14-postgis  -- Ubuntu/Debian
brew install postgis                        -- macOS
```

### Hỏi: Làm thế nào để xuất dữ liệu ranh giới sang GeoJSON?

**PostgreSQL / PostGIS:**

```sql
SELECT
    id,
    province_code,
    ST_AsGeoJSON(geom) AS geojson
FROM gis_provinces;
```

Xuất kết quả truy vấn và chuyển đổi cột `geojson` sang định dạng GeoJSON tiêu chuẩn.

**Một bộ dữ liệu GeoJSON chuyên biệt sẽ được cung cấp trong các phiên bản tương lai.**

### Hỏi: Tôi có thể sử dụng dữ liệu GIS với các thư viện bản đồ như Mapbox hoặc Leaflet không?

**Trả lời:** Có. Bộ dữ liệu GIS tương thích với nhiều thư viện bản đồ phổ biến như Leaflet, Mapbox GL JS, OpenLayers và ArcGIS.

Cách tiếp cận được khuyến nghị là xuất dữ liệu hình học dưới dạng **GeoJSON** từ cơ sở dữ liệu và cung cấp thông qua API:

```javascript
fetch('/api/provinces/01/boundary')
    .then(response => response.json())
    .then(geojson => {
        L.geoJSON(geojson).addTo(map);
    });
```

Quy trình điển hình:

1. Truy vấn dữ liệu ranh giới từ bộ dữ liệu GIS.
2. Chuyển đổi dữ liệu hình học sang GeoJSON (ví dụ bằng `ST_AsGeoJSON()` trong PostGIS).
3. Trả về dữ liệu GeoJSON từ API backend.
4. Hiển thị GeoJSON bằng thư viện bản đồ mong muốn.

> GeoJSON là định dạng trao đổi dữ liệu được khuyến nghị cho các ứng dụng bản đồ web.

### Hỏi: Dữ liệu ranh giới được cập nhật với tần suất như thế nào?

**Trả lời:** Dữ liệu được cập nhật khi Chính phủ Việt Nam ban hành các nghị quyết hoặc quyết định điều chỉnh địa giới hành chính (thông thường từ 2–4 lần mỗi năm). Hãy theo dõi repository của dự án để nhận các bản cập nhật mới nhất.

### Hỏi: Có công cụ GUI nào được khuyến nghị để trực quan hóa dữ liệu GIS không?

**Trả lời:** Có. Các công cụ sau hoạt động rất tốt với bộ dữ liệu GIS:

| Công cụ | Loại | Khuyến nghị |
|----------|----------|----------|
| DBeaver | Công cụ quản lý cơ sở dữ liệu | Khuyến nghị cho đa số người dùng. Hỗ trợ PostgreSQL, MySQL và SQL Server, đồng thời có thể xem trực tiếp dữ liệu hình học trên nền bản đồ OpenStreetMap. |
| QGIS | Phần mềm GIS Desktop | Phù hợp cho phân tích GIS chuyên sâu, chỉnh sửa dữ liệu không gian và biên tập bản đồ. |
| geojson.io | Công cụ Web | Công cụ trực tuyến nhẹ để trực quan hóa và kiểm tra dữ liệu GeoJSON xuất từ cơ sở dữ liệu. |

**Ví dụ: Xem dữ liệu hình học trong DBeaver**

[![image.png](https://i.postimg.cc/dVmQJSFg/image.png)](https://postimg.cc/k2GPcsBy)

**Ví dụ: Xuất GeoJSON để sử dụng với geojson.io**

```sql
SELECT
    province_code,
    ST_AsGeoJSON(geom) AS geojson
FROM gis_provinces;
```

Sao chép kết quả GeoJSON và dán vào https://geojson.io/ để xem trực quan ranh giới hành chính trên bản đồ.

Đối với đa số người dùng của dự án, **DBeaver** là lựa chọn được khuyến nghị vì miễn phí, đa nền tảng và hỗ trợ PostgreSQL, MySQL cũng như SQL Server.

---

## Tài nguyên tham khảo

- [Tài liệu PostGIS](https://postgis.net/documentation/)
- [Kiểu dữ liệu không gian trong MySQL](https://dev.mysql.com/doc/refman/8.0/en/spatial-types.html)
- [Dữ liệu không gian trong SQL Server](https://docs.microsoft.com/en-us/sql/relational-databases/spatial/spatial-data-sql-server)
- [Khái niệm và tiêu chuẩn GIS](https://en.wikipedia.org/wiki/Geographic_information_system)
- [Hệ quy chiếu WGS 84](https://epsg.io/4326)

---

## Đóng góp

Nếu bạn phát hiện lỗi trong bộ dữ liệu GIS hoặc có đề xuất cải tiến, vui lòng tạo một issue trong repository của dự án nha.

**Cập nhật gần nhất:** June 20, 2026

[gis_dataset_postgresql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/postgresql_ImportData_gis_2026-06-20__12_32_01.sql
[gis_dataset_mysql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/mysql_ImportData_gis_2026-06-20__12_32_01.sql
[gis_dataset_sqlserver_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/mssql_ImportData_gis_2026-06-20__12_32_02.sql
