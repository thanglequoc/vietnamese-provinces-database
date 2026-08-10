# Phần mở rộng Dữ liệu GIS — Vietnamese Provinces Database
**Dữ liệu GIS Dataset mới nhất: **v4.0.0** (June 20, 2026)**

|Nền tảng|Link tải|Kích thước|
|--------|-------------|---------|
|PostgreSQL/PostGIS|[Tải về file SQL GIS Dataset cho PostgreSQL][gis_dataset_postgresql_bucket_url]|~152.07 MB|
|MySQL|[Tải về file SQL GIS Dataset cho MySQL][gis_dataset_mysql_bucket_url]|~150.44 MB|
|Microsoft SQL Server|[Tải về file SQL GIS Dataset cho SQL Server][gis_dataset_sqlserver_bucket_url]|~152.13 MB|
|Elasticsearch| Truy cập dữ liệu GIS trong thư mục `elasticsearch/` của repository này | - |
|MongoDB| Truy cập dữ liệu GIS trong thư mục `mongodb/gis/` của repository này | - |

## Mục lục

1. [Giới thiệu](#gioi-thieu)
2. [Phạm vi dữ liệu](#pham-vi-du-lieu)
3. [Xuất GeoJSON](#xuat-geojson)
4. [Cài đặt](#cai-dat)
5. [Lược đồ cơ sở dữ liệu](#luoc-do-co-so-du-lieu)
6. [Mẹo tối ưu truy vấn](#meo-toi-uu-truy-van)
7. [Khả năng tương thích phiên bản](#kha-nang-tuong-thich-phien-ban)
8. [Câu hỏi thường gặp & Khắc phục sự cố](#cau-hoi-thuong-gap--khac-phuc-su-co)

---

## Giới thiệu

**Bộ dữ liệu GIS** là một phần mở rộng tùy chọn của dự án Vietnamese Provinces Database, cung cấp dữ liệu hình học ranh giới hành chính có độ chính xác cao cho các đơn vị hành chính của Việt Nam. Bộ dữ liệu này chứa thông tin ranh giới địa lý chi tiết của toàn bộ 34 tỉnh/thành phố và 3.321 xã/phường, hỗ trợ các truy vấn không gian (geospatial queries) và trực quan hóa bản đồ một cách hiệu quả.

Dữ liệu ranh giới GIS được xây dựng dựa trên dữ liệu từ [Bản đồ tra cứu các đơn vị hành chính Việt Nam](https://sapnhap.bando.com.vn), do Nhà xuất bản Tài nguyên, Môi trường và Bản đồ Việt Nam thuộc Bộ Nông nghiệp và Môi trường phát hành.

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

> **Lưu ý:** Đối với các nền tảng dựa trên document (Elasticsearch, MongoDB), dữ liệu GIS tương tự được phân phối dưới dạng document index/collection thay vì bảng SQL. Xem mục [Elasticsearch](#elasticsearch) và [MongoDB](#mongodb).

---

## Xuất GeoJSON

Dữ liệu GeoJSON hiện đã có sẵn bên cạnh các file SQL import. Bộ dữ liệu này phù hợp để dùng trực tiếp cho trình xem bản đồ, ứng dụng client-side và các quy trình GIS dựa trên file.

### Vị trí xuất dữ liệu

Các file được tạo tại:

```text
json/geojson/
```

Cấu trúc thư mục:

```text
geojson/
  README.md
  {province_code}_{province_code_name}/
    {province_code}_{province_code_name}.geojson
    wards/
      {ward_code}_{ward_code_name}.geojson
```

Quá trình xuất cũng tạo file nén:

```text
vn_provinces_wards_geojson_<datetime>.zip
```

### Cấu trúc GeoJSON

Mỗi file `.geojson` là một `FeatureCollection` với:

- `type`, `bbox` và `features` ở cấp cao nhất
- mỗi file chỉ chứa một `Feature`
- `id` của feature được gán theo mã đơn vị hành chính
- `bbox` của feature
- dữ liệu hình học theo chuẩn GeoJSON
- các thuộc tính mô tả đơn vị hành chính

### Xem kết quả

Mở [https://geojson.io](https://geojson.io) và tải bất kỳ file `.geojson` nào trong `json/geojson/` để kiểm tra hình học và thuộc tính một cách trực quan.

---

## Cài đặt

Bộ dữ liệu GIS yêu cầu các tập lệnh SQL khởi tạo (bootstrap scripts) để tạo bảng và các tập lệnh nhập dữ liệu (import scripts) để nạp dữ liệu ranh giới. Các tập lệnh hiện được cung cấp cho PostgreSQL, MySQL và SQL Server. Các định dạng dựa trên document (Elasticsearch và MongoDB) cũng được trình bày trong mục này.

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

Tải tập dữ liệu GIS đã chia nhỏ từ thư mục [postgresql/gis](../../postgresql/gis/).
Dữ liệu được chia thành các phần nhỏ hơn 40 MB, đặt tên
`postgresql_ImportData_gis_<timestamp>-part-NN.sql`. Nhập từng phần **theo thứ
tự**, theo danh sách trong tệp `.manifest` đi kèm. Ví dụ:

```bash
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_ImportData_gis_2026-06-20__12_32_01-part-01.sql
psql -U <username> -d <database_name> -f postgresql/gis/postgresql_ImportData_gis_2026-06-20__12_32_01-part-02.sql
# ... một lệnh psql cho mỗi phần, theo thứ tự trong manifest
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

Tải tập dữ liệu GIS đã chia nhỏ từ thư mục [mysql/gis](../../mysql/gis/).
Dữ liệu được chia thành các phần nhỏ hơn 40 MB, đặt tên
`mysql_ImportData_gis_<timestamp>-part-NN.sql`. Nhập từng phần **theo thứ tự**,
theo danh sách trong tệp `.manifest` đi kèm. Ví dụ:

```bash
mysql -u <username> -p <database_name> < mysql/gis/mysql_ImportData_gis_2026-06-20__12_32_01-part-01.sql
mysql -u <username> -p <database_name> < mysql/gis/mysql_ImportData_gis_2026-06-20__12_32_01-part-02.sql
# ... một lệnh mysql cho mỗi phần, theo thứ tự trong manifest
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

Tải tập dữ liệu GIS đã chia nhỏ từ thư mục [sqlserver/gis](../../sqlserver/gis/).
Dữ liệu được chia thành các phần nhỏ hơn 40 MB, đặt tên
`mssql_ImportData_gis_<timestamp>-part-NN.sql`. Nhập từng phần **theo thứ tự**,
theo danh sách trong tệp `.manifest` đi kèm. Ví dụ:

```cmd
sqlcmd -S <server_name> -d <database_name> -U <username> -P <password> -i sqlserver/gis/mssql_ImportData_gis_2026-06-20__12_32_02-part-01.sql
sqlcmd -S <server_name> -d <database_name> -U <username> -P <password> -i sqlserver/gis/mssql_ImportData_gis_2026-06-20__12_32_02-part-02.sql
# ... một lệnh sqlcmd cho mỗi phần, theo thứ tự trong manifest
```

#### Lưu ý riêng cho SQL Server

- Dữ liệu hình học được lưu bằng kiểu dữ liệu không gian gốc `geometry` của SQL Server.
- Các đối tượng hình học thường được tạo bằng các hàm như `geometry::STGeomFromText()` và `geometry::Point()`.
- Các phép toán không gian sử dụng cú pháp phương thức của đối tượng SQL Server (ví dụ: `geom.STContains()` và `geom.STDistance()`).
- Chỉ mục không gian được triển khai bằng tính năng **Spatial Index** tích hợp sẵn của SQL Server.

---

### Elasticsearch

Dữ liệu GIS được cung cấp dưới dạng document Elasticsearch trong thư mục `elasticsearch/` của repository này. Hai index được cung cấp:

| Index | Số document | Mô tả |
|-------|-------------|-------|
| `provinces` | 34 | Metadata cấp tỉnh/thành phố kèm phường nhúng và từ khóa tìm kiếm (không có hình học GIS) |
| `provinces-gis` | 34 | Cấu trúc giống `provinces` nhưng bổ sung đối tượng `GIS` ở cả cấp tỉnh và cấp phường (Center, BoundingBox, Geometry, Properties) |

Trong index `provinces-gis`, mỗi document tỉnh/thành phố và phường nhúng mang một đối tượng `GIS` gồm `Center` (`geo_point`), `BoundingBox`, `Geometry` (`geo_shape`) và `Properties` (bao gồm `GisServerId` và `AreaKm2`).

Để biết cách tạo index, nhập dữ liệu và các ví dụ truy vấn, xem [README Elasticsearch](../../elasticsearch/README.md).

---

### MongoDB

Dữ liệu GIS được cung cấp dưới dạng document MongoDB trong thư mục `mongodb/gis/` của repository này. Hai collection được cung cấp:

| Collection | Số document | Mô tả |
|------------|-------------|-------|
| `provinces-gis` | 34 | Document cấp tỉnh/thành phố có hình học GIS (Center, BoundingBox, Geometry, Properties) |
| `wards-gis` | 3.321 | Document cấp phường độc lập có hình học GIS kèm trường `ProvinceCode` để join giữa các collection |

Mỗi document mang một đối tượng `GIS` gồm `Center` (GeoJSON `Point`), `BoundingBox`, `Geometry` (GeoJSON `MultiPolygon`/`Polygon`) và `Properties` (bao gồm `GisServerId` và `AreaKm2`).

Để biết cách nhập dữ liệu, tạo index và các ví dụ truy vấn, xem [README MongoDB GIS](../../mongodb/gis/README.md).

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
| **Elasticsearch** | 7.x / 8.x | ✓ Hỗ trợ đầy đủ | Sử dụng `geo_shape` cho ranh giới và `geo_point` cho tâm |
| **MongoDB** | 4.0+ | ✓ Hỗ trợ đầy đủ | Sử dụng hình học GeoJSON với index `2dsphere` |
| **WGS 84 (SRID 4326)** | ISO 19115 | ✓ Tiêu chuẩn | Được hỗ trợ trên tất cả nền tảng |

### Yêu cầu tối thiểu

- **PostgreSQL:** 12 (phát hành năm 2019)
- **PostGIS:** 3.0 (phát hành năm 2019)
- **MySQL:** 8.0 (phát hành năm 2018)
- **SQL Server:** 2019 (phát hành năm 2019)
- **Elasticsearch:** 7.x (phát hành năm 2019)
- **MongoDB:** 4.0 (phát hành năm 2019)

### Phiên bản khuyến nghị để đạt hiệu năng tối ưu

- **PostgreSQL:** 15+ (ổn định, hỗ trợ xử lý song song tốt hơn)
- **PostGIS:** 3.3+ (cải thiện hiệu năng)
- **MySQL:** 8.0.30+ (các bản phát hành mới nhất của nhánh 8.0)
- **SQL Server:** 2022 (phiên bản mới nhất)
- **Elasticsearch:** 8.x (phiên bản chính mới nhất)
- **MongoDB:** 6.0+ (phiên bản chính mới nhất)

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
| geojson.io | Công cụ Web | Công cụ trực tuyến nhẹ để trực quan hóa và kiểm tra dữ liệu GeoJSON xuất từ bộ xuất của trình tạo dữ liệu. |

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
- [Các hàm không gian trong MySQL](https://dev.mysql.com/doc/refman/8.4/en/spatial-analysis-functions.html)
- [Dữ liệu không gian trong SQL Server](https://docs.microsoft.com/en-us/sql/relational-databases/spatial/spatial-data-sql-server)
- [Truy vấn không gian trong Elasticsearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/geo-queries.html)
- [Mapping geo_shape trong Elasticsearch](https://www.elastic.co/guide/en/elasticsearch/reference/current/geo-shape.html)
- [Truy vấn không gian trong MongoDB](https://www.mongodb.com/docs/manual/geospatial-queries/)
- [Index 2dsphere trong MongoDB](https://www.mongodb.com/docs/manual/core/2dsphere/)
- [Khái niệm và tiêu chuẩn GIS](https://en.wikipedia.org/wiki/Geographic_information_system)
- [Hệ quy chiếu WGS 84](https://epsg.io/4326)

---

## Đóng góp

Nếu bạn phát hiện lỗi trong bộ dữ liệu GIS hoặc có đề xuất cải tiến, vui lòng tạo một issue trong repository của dự án nha.

**Cập nhật gần nhất:** August 2, 2026

[gis_dataset_postgresql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/postgresql_ImportData_gis_2026-06-20__12_32_01.sql.manifest
[gis_dataset_mysql_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/mysql_ImportData_gis_2026-06-20__12_32_01.sql.manifest
[gis_dataset_sqlserver_bucket_url]: https://vn-provinces-ds.thanglequoc.xyz/v4.0.0/GISDataSet/mssql_ImportData_gis_2026-06-20__12_32_02.sql.manifest
