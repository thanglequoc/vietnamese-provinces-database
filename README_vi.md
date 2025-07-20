![Repository Banner](https://i.imgur.com/QZQhcql.png)
![Made in Vietnam](https://raw.githubusercontent.com/webuild-community/badge/master/svg/made.svg)

[Read in English version](README.md)

# Dữ liệu Tỉnh thành, Quận huyện Việt Nam

Đây là tập lệnh cơ sở dữ liệu SQL của toàn bộ đơn vị hành chính Việt Nam, bao gồm **34 tỉnh thành** và các Quận huyện, phường xã liên quan.  
Dữ liệu được cập nhật theo nghị định gần nhất: [19/2025/QĐ-TTg][source government decree]  

Nếu bạn thấy dự án này hữu ích, hãy để lại một ⭐ để ủng hộ nhé — điều đó sẽ tiếp thêm động lực để chúng tôi tiếp tục cải tiến và mang đến những công cụ giá trị cho cộng đồng. Ngoài ra, việc "star" repo cũng giúp bạn dễ dàng theo dõi các bản cập nhật trong tương lai.

## Tổng quan

Tác giả của dự án không làm việc, hay đại diện cho **Tổng cục Thống kê Việt Nam**, lẫn chính phủ nước Việt Nam.
Dữ liệu của Tỉnh thành, Quận huyện và Phường xã được tổng kết và hệ thống dựa trên dữ liệu tỉnh thành được cung cấp bởi [API từ trang web Đơn vị hành chính của Tổng cục Thống kê Việt Nam][source goverment API].  

**Lưu ý**: Do API SOAP của Tổng cục Thống kê (GSO) chưa được cập nhật theo thay đổi mới nhất liên quan đến việc tách thành 34 tỉnh, nên dữ liệu mới nhất hoàn toàn dựa trên văn bản chính thức Nghị định [19/2025/QĐ-TTg][decree 19/2025/QĐ-TTg].

Ngoài ra, cơ sở dữ liệu này còn có thêm những thông tin bổ sung, xin xem chi tiết trong phần **Các thay đổi thêm** ngay bên dưới.  

### Các phiên bản của bộ dữ liệu và Nghị định của Chính phủ

Chính phủ Việt Nam có thể ban hành những nghị định để thay đổi cấu trúc của các đơn vị hành chính Việt Nam theo thời gian. Bạn có thể theo dõi danh mục nghị định thay đổi đơn vị được ban hành tại [đây][decree issued page].

Bảng dưới thông kê các nghị định đã được ban hành, cùng thời gian có hiệu lực, cùng phiên bản của bộ dữ liệu tỉnh thành Việt Nam từ phiên bản đầu tiên.

|Nghị định|Ngày ban hành|Ngày có hiệu lực|Phiên bản|
|-------------|-----------|-------------|---------------|
|[19/2025/QĐ-TTg][decree 19/2025/QĐ-TTg]|30/06/2025|01/07/2025|v3.0.0|
|Sửa chính tả, cutoff dữ liệu trước  19/2025/QĐ-TTg|15/01/2025|01/03/2025|v2.4.1
|[1365/NQ-UBTVQH15][decree 1365/NQ-UBTVQH15]|15/01/2025|01/03/2025|v2.4.0
|[1314/NQ-UBTVQH15][decree 1314/NQ-UBTVQH15]|30/11/2024|01/01/2025|v2.3.0
|[1203/NQ-UBTVQH15][decree 1203/NQ-UBTVQH15]|28/09/2024|01/11/2024|v2.2.0
|[1106/NQ-UBTVQH15][decree 1106/NQ-UBTVQH15]|23/07/2024|01/09/2024|v2.1.0
|[1013/NQ-UBTVQH15][decree 1013/NQ-UBTVQH15]|19/03/2024|01/05/2024|v2.0.1
|[939/NQ-UBTVQH15][decree 939/NQ-UBTVQH15]|13/12/2023|01/02/2024 |v2.0.0
|Từ [721/NQ-UBTVQH15][decree 721/NQ-UBTVQH15] đến<br>[730/NQ-UBTVQH15][decree 730/NQ-UBTVQH15]|13/02/2023|10/04/2023 |v1.0.4.1
|[569/NQ-UBTVQH15][decree 569/NQ-UBTVQH15],<br>[570/NQ-UBTVQH15][decree 570/NQ-UBTVQH15]|11/08/2022|01/10/2022 |v1.0.3.1
|[510/NQ-UBTVQH15][decree 510/NQ-UBTVQH15]|12/05/2022|01/07/2022|v1.0.2
|[469/NQ-UBTVQH15][decree 469/NQ-UBTVQH15]|15/02/2022|10/04/2022|v1.0.1
|[387/NQ-UBTVQH15][decree 387/NQ-UBTVQH15]|22/09/2021|01/11/2021|v1.0.0


### Các thay đổi thêm

- Thêm bảng quan hệ `administrative_regions`
- Thêm bảng quan hệ `administrative_units`
- Đặt dữ liệu tên đơn vị hành chính cho các giá trị tỉnh thành, phường xã  
- Tạo các tên riêng bằng tiếng Anh cho các giá trị tỉnh thành, phường xã  
- Tạo mã từ tên các tỉnh thành, phường xã  

## Hướng dẫn cài đặt

### Postgresql

Bạn có thể nạp dữ liệu vào cơ sở dữ liệu hiện có, hoặc tạo một cơ sở dữ liệu mới:

```sql
CREATE DATABASE vietnamese_administrative_units;
```

Chạy tệp `CreateTable_vn_units.sql` trong [thư mục postgresql](postgresql) trước để khởi tạo các bảng và quan hệ cần thiết.  
Sau đó chạy tiếp tệp `ImportData_vn_units.sql` để nạp dữ liệu vào các bảng đã tạo.

### MySQL - MariaDB

Bạn có thể nạp dữ liệu vào cơ sở dữ liệu hiện có, hoặc tạo một cơ sở dữ liệu mới:

```sql
CREATE DATABASE vietnamese_administrative_units;
```

Chạy tệp `CreateTable_vn_units.sql` trong [thư mục mysql](mysql) trước để khởi tạo các bảng và quan hệ cần thiết.  
Sau đó chạy tiếp tệp `ImportData_vn_units.sql` để nạp dữ liệu vào các bảng đã tạo.

### Microsoft SQL Server

Bạn có thể nạp dữ liệu vào cơ sở dữ liệu hiện có, hoặc tạo một cơ sở dữ liệu mới:

```sql
CREATE DATABASE vietnamese_administrative_units;
```

Chạy tệp `CreateTable_vn_units.sql` trong [thư mục sqlserver](sqlserver) trước để khởi tạo các bảng và quan hệ cần thiết.  
Sau đó chạy tiếp tệp `ImportData_vn_units.sql` để nạp dữ liệu vào các bảng đã tạo.

### Oracle

Bạn có thể nạp dữ liệu vào cơ sở dữ liệu hiện có, hoặc tạo một cơ sở dữ liệu mới:

```sql
CREATE DATABASE vietnamese_administrative_units;
```

Chạy tệp `CreateTable_vn_units.sql` trong [thư mục oracle](oracle) trước để khởi tạo các bảng và quan hệ cần thiết.  
Sau đó chạy tiếp tệp `ImportData_vn_units.sql` để nạp dữ liệu vào các bảng đã tạo.

## Lược đồ quan hệ

![VN_administrative_units db](https://i.imgur.com/XEIgaXV.png)

### Bảng quan hệ `administrative_regions`

![VN Geographical Regions](https://i.imgur.com/CiyxQi0.png)  
Bảng quan hệ `administrative_regions` chứa danh sách **8** khu vực địa lý của Việt Nam, với định danh `id` tăng dần theo vị trí khu vực theo chiều từ Bắc vào Nam.

#### Cấu trúc bảng dữ liệu

|Cột|Kiểu dữ liệu|Ý nghĩa|Ràng buộc|
|------|-----------|---------|------------|
|`id`|integer|Mã định danh của khu vực|Khoá chính|
|`name`|varchar(255)|Tên khu vực bằng tiếng Việt||
|`name_en`|varchar(255)|Tên khu vực bằng tiếng Anh||
|`code_name`|varchar(255)|Tên mã khu vực bằng tiếng Việt, tạo theo định dạng chữ thường xếp gạch||
|`code_name_en`|varchar(255)|Tên mã khu vực bằng tiếng Anh, tạo theo định dạng chữ thường xếp gạch||

#### Dữ liệu mẫu

|id|name|name_en|code_name|code_name_en|
|--|----|-------|---------|------------|
|1|Đông Bắc Bộ|Northeast|dong_bac_bo|northest|
|2|Tây Bắc Bộ|Northwest|tay_bac_bo|northwest|
|3|Đồng bằng sông Hồng|Red River Delta|dong_bang_song_hong|red_river_delta|
|4|Bắc Trung Bộ|North Central Coast|bac_trung_bo|north_central_coast|
|5|Duyên hải Nam Trung Bộ|South Central Coast|duyen_hai_nam_trung_bo|south_central_coast|
|6|Tây Nguyên|Central Highlands|tay_nguyen|central_highlands|
|7|Đông Nam Bộ|Southeast|dong_nam_bo|southeast|
|8|Đồng bằng sông Cửu Long|Mekong River Delta|dong_bang_song_cuu_long|southwest|

### Bảng quan hệ `administrative_units`

![VN Units](https://i.imgur.com/U0Warh3.png)  

Bảng quan hệ `administrative_units` chứa danh sách các đơn vị hành chính với định danh `id` được xếp theo 2 phân cấp đơn vị hành chính.  

#### Cấu trúc bảng dữ liệu

|Cột|Kiểu dữ liệu|Ý nghĩa|Ràng buộc|
|------|-----------|---------|------------|
|`id`|integer|Mã định danh của đơn vị hành chính|Khoá chính|
|`full_name`|varchar(255)|Tên tiếng Việt đầy đủ của đơn vị hành chính||
|`full_name_en`|varchar(255)|Tên tiếng Anh đầy đủ của đơn vị hành chính||
|`short_name`|varchar(255)|Tên tiếng Việt thông dụng của đơn vị hành chính||
|`short_name_en`|varchar(255)|Tên tiếng Anh thông dụng của đơn vị hành chính||
|`code_name`|varchar(255)|Tên mã đơn vị dạng tiếng Việt dựa trên cột `full_name`, tạo theo định dạng chữ thường xếp gạch||
|`code_name_en`|varchar(255)|Tên mã đơn vị dạng tiếng Anh dựa trên cột `full_name_en`, tạo theo định dạng chữ thường xếp gạch||

#### Dữ liệu mẫu

|id|full_name|full_name_en|short_name|short_name_en|code_name|code_name_en|
|--|---------|------------|----------|-------------|---------|------------|
|1|Thành phố trực thuộc trung ương|Municipality|Thành phố|City|thanh_pho_truc_thuoc_trung_uong|municipality|
|2|Tỉnh|Province|Tỉnh|Province|tinh|province|
|3|Phường|Ward|Phường|Ward|phuong|ward|
|4|Xã|Commune|Xã|Commune|xa|commune|
|5|Đặc khu tại hải đảo|Special administrative region|Đặc khu|Special administrative region|dac_khu|special_administrative_region|

### Bảng quan hệ `provinces`
![Provincial level](https://i.imgur.com/cLTRHkf.png)  
Bảng quan hệ `provinces` chứa danh sách đơn vị hành chính **cấp 1 - Tỉnh thành**, bao gồm **34** thành phố trực thuộc trung ương và tỉnh.  

#### Cấu trúc bảng dữ liệu

|Cột|Kiểu dữ liệu|Ý nghĩa|Ràng buộc|
|------|-----------|---------|------------|
|`code`|varchar(20)|Mã đơn vị chính thức, quy ước bởi chính phủ|Khoá chính|
|`name`|varchar(255)|Tên tiếng Việt||
|`name_en`|varchar(255)|Tên tiếng Anh||
|`full_name`|varchar(255)|Tên tiếng Việt đầy đủ kèm tên đơn vị hành chính||
|`full_name_en`|varchar(255)|Tên tiếng Anh đầy đủ kèm tên đơn vị hành chính||
|`code_name`|varchar(255)|Tên mã dựa trên cột `name`, tạo theo định dạng chữ thường xếp gạch||
|`administrative_unit_id`|integer|Mã đơn vị hành chính của đối tượng|Khoá ngoại, liên kết đến bảng `administrative_units.id` |


#### Dữ liệu mẫu

|code|name|name_en|full_name|full_name_en|code_name|administrative_unit_id|
|----|----|-------|---------|------------|---------|----------------------|
|01|Hà Nội|Ha Noi|Thành phố Hà Nội|Ha Noi City|ha_noi|1|
|56|Khánh Hòa|Khanh Hoa|Tỉnh Khánh Hòa|Khanh Hoa Province|khanh_hoa|2|
|79|Hồ Chí Minh|Ho Chi Minh|Thành phố Hồ Chí Minh|Ho Chi Minh City|ho_chi_minh|1|
|96|Cà Mau|Ca Mau|Tỉnh Cà Mau|Ca Mau Province|ca_mau|2|
|..|...........|...........|.....................|................|...........|..|..|

### Bảng quan hệ `wards`
[![Commune level](https://i.postimg.cc/5NfSpCG4/ward-structure.avif)](https://postimg.cc/mh69gtBJ)
Bảng quan hệ `wards` chứa danh sách **đơn vị hành chính cấp 2**, bao gồm **3321** xã phường, đặc khu.

#### Cấu trúc bảng dữ liệu

|Cột|Kiểu dữ liệu|Ý nghĩa|Ràng buộc|
|------|-----------|---------|------------|
|`code`|varchar(20)|Mã đơn vị chính thức, quy ước bởi chính phủ|Khoá chính|
|`name`|varchar(255)|Tên tiếng Việt||
|`name_en`|varchar(255)|Tên tiếng Anh||
|`full_name`|varchar(255)|Tên tiếng Việt đầy đủ kèm tên đơn vị hành chính||
|`full_name_en`|varchar(255)|Tên tiếng Anh đầy đủ kèm tên đơn vị hành chính||
|`code_name`|varchar(255)|Tên mã dựa trên cột `name`, tạo theo định dạng chữ thường xếp gạch||
|`province_code`|integer|Mã tỉnh thành (`province`) mà đối tượng phường xã này thuộc về|Khoá ngoại, liên kết đến bảng `provinces.code`|
|`administrative_unit_id`|integer|Mã đơn vị hành chính của đối tượng|Foreign Key, references to `administrative_units.id` |

#### Dữ liệu mẫu

|code|name|name_en|full_name|full_name_en|code_name|province_code|administrative_unit_id|
|----|----|-------|---------|------------|---------|-------------|----------------------|
|25920|Tân Hiệp|Tan Hiep|Phường Tân Hiệp|Tan Hiep Ward|tan_hiep|79|3|
|25942|Dĩ An|Di An|Phường Dĩ An|Di An Ward|di_an|79|3|
|25945|Tân Đông Hiệp|Tan Dong Hiep|Phường Tân Đông Hiệp|Tan Dong Hiep Ward|tan_dong_hiep|79|3|
|25951|Đông Hòa|Dong Hoa|Phường Đông Hòa|Dong Hoa Ward|dong_hoa|79|3|
|25966|Lái Thiêu|Lai Thieu|Phường Lái Thiêu|Lai Thieu Ward|lai_thieu|79|3|
|25969|Thuận Giao|Thuan Giao|Phường Thuận Giao|Thuan Giao Ward|thuan_giao|79|3|
|25975|An Phú|An Phu|Phường An Phú|An Phu Ward|an_phu|79|3|

## Câu truy vấn SQL mẫu

Bạn có thể dễ dàng viết các câu truy vấn để lấy, lọc dữ liệu tương ứng bằng cách tạo các kết (`JOIN`) giữa các bảng dựa trên giá trị khoá chính, khoá ngoại.  
Phía sau là một vài câu truy vấn mẫu để tham khảo:  

### Tìm toàn bộ xã phường thuộc quận huyện

Tìm toàn bộ xã phường thuộc **tỉnh Khánh Hoà**  

```sql
SELECT w.code, w."name" , w.full_name , w.full_name_en ,au.full_name as administrative_unit_name
FROM wards w 
INNER JOIN administrative_units au 
ON w.administrative_unit_id = au.id
WHERE w.province_code = '56' -- Khanh Hoa province code
ORDER BY w.code;
```

|code|name|full_name|full_name_en|administrative_unit_name|
|----|----|---------|------------|------------------------|
|22333|Bắc Nha Trang|Phường Bắc Nha Trang|Bac Nha Trang Ward|Phường|
|22366|Nha Trang|Phường Nha Trang|Nha Trang Ward|Phường|
|22390|Tây Nha Trang|Phường Tây Nha Trang|Tay Nha Trang Ward|Phường|
|22402|Nam Nha Trang|Phường Nam Nha Trang|Nam Nha Trang Ward|Phường|
|22411|Bắc Cam Ranh|Phường Bắc Cam Ranh|Bac Cam Ranh Ward|Phường|
|22420|Cam Ranh|Phường Cam Ranh|Cam Ranh Ward|Phường|
|22423|Ba Ngòi|Phường Ba Ngòi|Ba Ngoi Ward|Phường|
|22432|Cam Linh|Phường Cam Linh|Cam Linh Ward|Phường|

## Dữ liệu định dạng Non-SQL

Ngoài định dạng SQL, dataset tỉnh thành Việt Nam này cũng cung cấp thêm dưới các định dạng sau:

- Dạng **JSON**, bao gồm bản đầy đủ, tối giản, và tốn giản chỉ có tiếng Việt
- **MongoDB**
- **Redis**

## Câu hỏi thường gặp

### Dự án này xây dựng dữ liệu từ đâu?

Dữ liệu về các tỉnh và phường được tạo dựa trên [API dữ liệu tỉnh do Tổng cục Thống kê Việt Nam cung cấp trên website chính thức][source goverment API].
Lưu ý: Do API SOAP của Tổng cục Thống kê chưa được cập nhật theo thay đổi mới nhất liên quan đến việc chia tách 34 tỉnh. Vì vậy, dữ liệu mới nhất hoàn toàn dựa trên văn bản [Nghị định 19/2025/QĐ-TTg][decree 19/2025/QĐ-TTg].

### Các khoá định danh được định nghĩa dựa trên đâu?

|Bảng quan hệ|Khoá chính|
|-----|-----------|
|`administrative_regions`|Khoá chính: `id`. Tăng dần `1` đến `8` theo vị trí vùng địa lý hướng từ Bắc vào Nam
|`administrative_units`|Khoá chính: `id`. Tăng dần từ `1` đến `5` theo phân cấp bậc đơn vị hành chính
|`provinces`|Khoá chính: `code`. Được quy ước theo đối tượng **Tỉnh thành** do chính phủ ban hành
|`wards`|Khoá chính: `code`. Được quy ước theo đối tượng **Phường xã** do chính phủ ban hành

### Tại sao mối quan hệ giữa tỉnh và vùng hành chính (region) bị loại bỏ từ phiên bản v3.0.0?

Sau khi số lượng tỉnh được sáp nhập còn 37, các tỉnh mới — chẳng hạn như Phú Thọ, được hình thành từ Vĩnh Phúc, Phú Thọ và Hòa Bình — hiện nay trải dài trên nhiều vùng hành chính trước đây. Do các tỉnh cũ thuộc các vùng khác nhau, nên việc xác định vùng hành chính cho tỉnh mới không còn chính xác nữa.

### Tôi tìm thấy một vài lỗi trong tệp dữ liệu SQL này?

Nếu bạn có bất kỳ một đề xuất nào có thể cải tiến dự án, xin vui lòng [tạo một Issue](https://github.com/ThangLeQuoc/VietnameseProvincesDatabase/issues) và cung cấp thông tin cụ thể.  
Hoặc tốt hơn nữa, bạn có thể đóng góp xây dựng dự án này bằng các [tạo Pull Request](https://github.com/ThangLeQuoc/VietnameseProvincesDatabase/pulls)  
Tất cả các đóng góp đến dự án đều được trân trọng ghi nhận.

##### Nguồn tham khảo
Bản đồ Việt Nam dùng làm banner từ [vietcentertourist](https://vietcentertourist.com/assets/images/vietnam.png)

[source danhmuchanhchinh gov]: https://danhmuchanhchinh.gso.gov.vn/
[source government decree]: https://danhmuchanhchinh.gso.gov.vn/NghiDinh.aspx
[source goverment API]: https://danhmuchanhchinh.gso.gov.vn/DMDVHC.asmx
[source danhmuchanhchinh gov]: https://danhmuchanhchinh.gso.gov.vn/
[source government decree]: https://danhmuchanhchinh.gso.gov.vn/NghiDinh.aspx
[decree issued page]: https://danhmuchanhchinh.gso.gov.vn/NghiDinh.aspx
[decree 387/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-387-NQ-UBTVQH15-thanh-lap-Toa-an-nhan-dan-thanh-pho-Tu-Son-thuoc-tinh-Bac-Ninh-490766.aspx
[decree 469/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-469-NQ-UBTVQH15-2022-thanh-lap-phuong-thuoc-thi-xa-Pho-Yen-Thai-Nguyen-504359.aspx
[decree 510/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-510-NQ-UBTVQH15-2022-thanh-lap-thi-tran-Phuong-Son-huyen-Luc-Nam-Bac-Giang-516371.aspx
[decree 569/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-569-NQ-UBTVQH15-2022-thanh-lap-thi-tran-Binh-Phu-thuoc-huyen-Cai-Lay-Tien-Giang-525909.aspx
[decree 570/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-570-NQ-UBTVQH15-2022-thanh-lap-thi-xa-Chon-Thanh-Binh-Phuoc-525910.aspx
[decree 721/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-721-NQ-UBTVQH15-2023-thanh-lap-thi-xa-Tinh-Bien-va-phuong-thuoc-thi-xa-An-Giang-556498.aspx
[decree 730/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-730-NQ-UBTVQH15-2023-thanh-lap-thi-tran-Kim-Long-thi-tran-Tam-Hong-Vinh-Phuc-556504.aspx
[decree 939/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-939-NQ-UBTVQH15-2023-nhap-xa-Thieu-Phu-vao-thi-tran-Thieu-Hoa-Thanh-Hoa-592292.aspx
[decree 1013/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-1013-NQ-UBTVQH15-2024-thanh-lap-cac-phuong-thuoc-thi-xa-Go-Cong-Tien-Giang-606022.aspx
[decree 1106/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-1106-NQ-UBTVQH15-2024-sap-xep-don-vi-hanh-chinh-cap-xa-Tuyen-Quang-619244.aspx  
[decree 1203/NQ-UBTVQH15]: https://thuvienphapluat.vn/banan/tin-tuc/nghi-quyet-ve-sap-xep-don-vi-hanh-chinh-tai-63-tinh-thanh-pho-giai-doan-20232025-11897  
[decree 1314/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-1314-NQ-UBTVQH15-2024-sap-xep-don-vi-hanh-chinh-cap-huyen-cap-xa-thanh-pho-Hue-634158.aspx
[decree 1365/NQ-UBTVQH15]: https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Nghi-quyet-1365-NQ-UBTVQH15-2025-thanh-lap-cac-phuong-thuoc-thi-xa-Phu-My-Vung-Tau-640985.aspx
[decree 19/2025/QĐ-TTg]: https://www.nso.gov.vn/default/2025/07/quyet-dinh-ban-hanh-bang-danh-muc-va-ma-so-cac-don-vi-hanh-chinh-viet-nam/
