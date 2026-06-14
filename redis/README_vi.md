# Bộ dữ liệu tỉnh thành Việt Nam cho Redis

Thư mục này chứa bản xuất dữ liệu đơn vị hành chính Việt Nam dành cho Redis. Dữ liệu lưu **34 tỉnh, thành phố** và **3.321 xã, phường, đặc khu** dưới dạng Redis hash, đồng thời có thêm các khóa tra cứu province-to-ward được dựng sẵn để đọc nhanh.

## Tệp dữ liệu

- `redis_vn_provinces_dataset.redis`: tệp lệnh Redis đã được commit sẵn, có thể nạp trực tiếp bằng `redis-cli`

Nếu bạn sinh lại dữ liệu từ `dataset-generation-scripts`, trình tạo dữ liệu có thể sinh ra tên tệp có timestamp như `redis_vn_provinces_dataset_<timestamp>.redis`. Cấu trúc key vẫn giữ nguyên.

## Nạp dữ liệu vào Redis

### Yêu cầu

- Có Redis server đang chạy
- Có `redis-cli` trên máy
- Có quyền truy cập tệp dữ liệu trong thư mục này

### Nạp vào Redis cục bộ

```bash
cd redis
redis-cli < redis_vn_provinces_dataset.redis
```

### Nạp vào Redis từ xa

```bash
cd redis
redis-cli -h 127.0.0.1 -p 6379 -a '<password>' -n 0 < redis_vn_provinces_dataset.redis
```

Tệp đang được commit là một script lệnh Redis thuần (`HSET`, `SADD`, ...), vì vậy cách nạp đúng cho định dạng hiện tại là dùng stdin redirection như trên.

### Kiểm tra sau khi nạp

```bash
redis-cli DBSIZE
redis-cli EXISTS province:01
redis-cli EXISTS ward:00004
redis-cli SCARD province:01:wards
```

Kết quả mong đợi:

- `DBSIZE` trả về giá trị lớn hơn `0`
- `EXISTS province:01` trả về `1`
- `EXISTS ward:00004` trả về `1`
- `SCARD province:01:wards` trả về số lượng xã, phường thuộc tỉnh/thành `01`

## Kiến trúc dữ liệu

Bản xuất Redis là biểu diễn key-value của bộ dữ liệu, không phải schema quan hệ như SQL. Mỗi thực thể được lưu bằng hash, còn quan hệ tỉnh -> xã/phường được dựng sẵn bằng set và hash tra cứu.

### Các nhóm key

| Mẫu key | Kiểu Redis | Mục đích |
|---------|------------|----------|
| `administrativeUnit:{id}` | Hash | Danh mục loại đơn vị hành chính như thành phố trực thuộc trung ương, tỉnh, phường, xã |
| `region:{id}` | Hash | Danh mục vùng địa lý của Việt Nam |
| `province:{code}` | Hash | Bản ghi tỉnh hoặc thành phố |
| `ward:{code}` | Hash | Bản ghi phường, xã hoặc đặc khu |
| `province:{province_code}:wards` | Set | Toàn bộ mã xã, phường thuộc một tỉnh |
| `province:{province_code}:wards:vn` | Hash | Ánh xạ mã xã/phường -> tên đầy đủ tiếng Việt trong một tỉnh |
| `province:{province_code}:wards:en` | Hash | Ánh xạ mã xã/phường -> tên đầy đủ tiếng Anh trong một tỉnh |

### Các trường được lưu

#### `administrativeUnit:{id}`

| Trường | Ý nghĩa |
|--------|---------|
| `id` | Mã loại đơn vị hành chính |
| `fullName` | Tên đầy đủ tiếng Việt |
| `fullNameEn` | Tên đầy đủ tiếng Anh |
| `shortName` | Tên ngắn tiếng Việt |
| `shortNameEn` | Tên ngắn tiếng Anh |
| `codeName` | Mã định danh theo tiếng Việt |

#### `region:{id}`

| Trường | Ý nghĩa |
|--------|---------|
| `name` | Tên vùng tiếng Việt |
| `nameEn` | Tên vùng tiếng Anh |
| `codeName` | Mã định danh theo tiếng Việt |

#### `province:{code}`

| Trường | Ý nghĩa |
|--------|---------|
| `code` | Mã tỉnh chính thức |
| `name` | Tên ngắn tiếng Việt |
| `nameEn` | Tên ngắn tiếng Anh |
| `fullName` | Tên đầy đủ tiếng Việt |
| `fullNameEn` | Tên đầy đủ tiếng Anh |
| `codeName` | Mã định danh theo tiếng Việt |
| `administrativeUnitId` | Liên kết tới `administrativeUnit:{id}` |

#### `ward:{code}`

| Trường | Ý nghĩa |
|--------|---------|
| `code` | Mã xã/phường chính thức |
| `name` | Tên ngắn tiếng Việt |
| `nameEn` | Tên ngắn tiếng Anh |
| `fullName` | Tên đầy đủ tiếng Việt |
| `fullNameEn` | Tên đầy đủ tiếng Anh |
| `codeName` | Mã định danh theo tiếng Việt |
| `administrativeUnitId` | Liên kết tới `administrativeUnit:{id}` |
| `districtCode` | Hiện đang lưu mã tỉnh cha trong bản xuất hiện tại |

## Sơ đồ quan hệ

```mermaid
flowchart TD
    AU[administrativeUnit:{id}\nHash]
    RG[region:{id}\nHash]
    P[province:{code}\nHash]
    W[ward:{code}\nHash]
    PS[province:{code}:wards\nSet chứa mã ward]
    PVN[province:{code}:wards:vn\nHash mã -> tên tiếng Việt]
    PEN[province:{code}:wards:en\nHash mã -> tên tiếng Anh]

    P -->|administrativeUnitId| AU
    W -->|administrativeUnitId| AU
    P --> PS
    PS --> W
    P --> PVN
    P --> PEN
    PVN --> W
    PEN --> W
```

Các key `region:{id}` hiện được đưa vào như dữ liệu tham chiếu độc lập trong bản xuất Redis.

## Một số thao tác Redis cơ bản

### Lấy thông tin một tỉnh

```bash
redis-cli HGETALL province:01
```

Các trường trả về thường gồm:

- `code`
- `name`
- `nameEn`
- `fullName`
- `fullNameEn`
- `codeName`
- `administrativeUnitId`

### Lấy thông tin một xã/phường

```bash
redis-cli HGETALL ward:00004
```

### Lấy toàn bộ mã xã/phường của một tỉnh

```bash
redis-cli SMEMBERS province:01:wards
```

### Lấy toàn bộ tên xã/phường của một tỉnh

Tên tiếng Việt:

```bash
redis-cli HGETALL province:01:wards:vn
```

Tên tiếng Anh:

```bash
redis-cli HGETALL province:01:wards:en
```

### Lấy bản ghi đầy đủ của các xã/phường trong một tỉnh

Redis không hỗ trợ join kiểu SQL, nên cách làm phổ biến là:

1. Đọc danh sách mã xã/phường từ set của tỉnh.
2. Lấy từng hash `ward:{code}` tương ứng.

Ví dụ bằng shell:

```bash
redis-cli --raw SMEMBERS province:01:wards | while read -r ward_code; do
  echo "ward:$ward_code"
  redis-cli HGETALL "ward:$ward_code"
done
```

### Xem metadata loại đơn vị hành chính

```bash
redis-cli HGETALL administrativeUnit:1
```

## Lưu ý về mô hình dữ liệu

- Bản xuất Redis được tối ưu cho truy cập theo key, không phải truy vấn quan hệ.
- Quan hệ tỉnh -> xã/phường đã được dựng sẵn nên thao tác liệt kê theo tỉnh khá nhanh.
- Hiện chưa có reverse index riêng từ tên xã/phường sang mã xã/phường.
- Các key `region:{id}` được đưa vào như dữ liệu tham chiếu, nhưng `province:{code}` hiện chưa lưu trực tiếp thông tin vùng.
- Một số tên trường phản ánh đúng logic của generator hiện tại hơn là một schema lý tưởng. Cụ thể, `districtCode` trong `ward:{code}` hiện đang chứa mã tỉnh cha.
