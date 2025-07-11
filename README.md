![Repository Banner](https://i.imgur.com/nWlKT5I.png)
![Made in Vietnam](https://raw.githubusercontent.com/webuild-community/badge/master/svg/made.svg)

[Đọc phiên bản tiếng Việt](README_vi.md)

## ❗️ Dataset is temporary outdated, updating is In Progress
This dataset is currently awaiting for update from the upstream data from the [**General Statistics Office of Vietnam**][source goverment API].  
At the moment it does not include the breaking change of re-organizing the country administrative unit.  
Since the GSO SOAP API is still outdated, the author is actively follow the latest decree [19/2025/QĐ-TTg](https://thuvienphapluat.vn/van-ban/Bo-may-hanh-chinh/Quyet-dinh-19-2025-QD-TTg-Bang-danh-muc-va-ma-so-cac-don-vi-hanh-chinh-Viet-Nam-663707.aspx) to work on the data patch

----

# Vietnamese Provinces Database

A complete SQL database of Vietnamese administrative units, includes all **63 Vietnamese provinces** and associated districts, wards sub-divisions.  
Data is updated as of the most recent effective decree: [1365/NQ-UBTVQH15][source government decree]  
Don't forget to leave a :star: if you find this repository helpful, and to keep track of the latest release of this dataset in the future. It's would help to cheer us up so we can deliver valuable product to support our community.

## Overview

The author(s) of this repository is not associated with the **General Statistics Office of Vietnam**, nor the Vietnamese government.  
The data of provinces, districts and wards are created base on the [API province data provided by the General Statistics Office of Vietnam website][source goverment API].  
This dataset also include additional information apart from the original provinces, districts and wards data from the original data. Please see section [Additional change make by this repository](#additional-change-make-by-this-repository)

### Dataset releases and Government issued decrees

The Vietnamese Government may issues decree time to time to change the administrative unit structure. You can track the latest issued decrees [here][decree issued page].  

This dataset will be gradually updated to keep up with the latest **effective** decree. To check the status of your dataset and how to keep the dataset up-to-date, see section [How to update the existing dataset?](#how-to-update-the-existing-dataset).

The following tables contains the list of issued decree, its effected date from, tracked from the earliest version of this dataset.

|Issued Decree|Issued on |Effect from|Release Version|
|-------------|-----------|-------------|---------------|
|[1365/NQ-UBTVQH15][decree 1365/NQ-UBTVQH15]|15/01/2025|01/03/2025|v2.4.0
|[1318/NQ-UBTVQH15][decree 1314/NQ-UBTVQH15]|30/11/2024|01/01/2025|v2.3.0
|[1203/NQ-UBTVQH15][decree 1203/NQ-UBTVQH15]|28/09/2024|01/11/2024|v2.2.0
|[1106/NQ-UBTVQH15][decree 1106/NQ-UBTVQH15]|23/07/2024|01/09/2024|v2.1.0
|[1013/NQ-UBTVQH15][decree 1013/NQ-UBTVQH15]|19/03/2024|01/05/2024|v2.0.1
|[939/NQ-UBTVQH15][decree 939/NQ-UBTVQH15]|13/12/2023|01/02/2024 |v2.0.0
|From [721/NQ-UBTVQH15][decree 721/NQ-UBTVQH15] to<br>[730/NQ-UBTVQH15][decree 730/NQ-UBTVQH15]|13/02/2023|10/04/2023 |v1.0.4.1
|[569/NQ-UBTVQH15][decree 569/NQ-UBTVQH15],<br>[570/NQ-UBTVQH15][decree 570/NQ-UBTVQH15]|11/08/2022|01/10/2022 |v1.0.3.1
|[510/NQ-UBTVQH15][decree 510/NQ-UBTVQH15]|12/05/2022|01/07/2022|v1.0.2
|[469/NQ-UBTVQH15][decree 469/NQ-UBTVQH15]|15/02/2022|10/04/2022|v1.0.1
|[387/NQ-UBTVQH15][decree 387/NQ-UBTVQH15]|22/09/2021|01/11/2021|v1.0.0


### Additional change make by this repository

- Add `administrative_regions` table
- Add `administrative_units` table
- Define the administrative unit and associated region for province, district and ward data
- Generate the English name for the provinces, districts and wards, offer both full and short forms
- Generate the code name for the provinces, districts and wards

## Installation

### Postgresql

Either use your existing database, or create a new one:

```sql
CREATE DATABASE vietnamese_administrative_units;
```

Execute the `CreateTable_vn_units.sql` in the [postgresql directory](postgresql) first in the target database to generate all the table structure.

Then follow up by executing the `ImportData_vn_units.sql` to import data to these generated tables.


### MySQL - MariaDB

Either use your existing database, or create a new one:

```sql
CREATE DATABASE vietnamese_administrative_units;
```

Execute the `CreateTable_vn_units.sql` in the [mysql directory](mysql) first in the target database to generate all the table structure.

Then follow up by executing the `ImportData_vn_units.sql` to import data to these generated tables.


### Microsoft SQL Server

Either use your existing database, or create a new one:

```sql
CREATE DATABASE vietnamese_administrative_units;
```

Execute the `CreateTable_vn_units.sql` in the [sqlserver directory](sqlserver) first in the target database to generate all the table structure.

Then follow up by executing the `ImportData_vn_units.sql` to import data to these generated tables.

### Oracle

Either use your existing database, or create a new one

Execute the `CreateTable_vn_units.sql` in the [oracle directory](oracle) first in the target database to generate all the table structure.

Then follow up by executing the `ImportData_vn_units.sql` to import data to these generated tables.

## Tables Schema

![VN_administrative_units db](https://i.imgur.com/sa0k4rt.png)

### `administrative_regions` table

![VN Geographical Regions](https://i.imgur.com/CiyxQi0.png)  
The `administrative_regions` table contains the list of **8** Vietnamese geographical regions with the `id` increment following the region location from North to South.

#### Table definition

|Column|Data type|Meaning|Constraint|
|------|-----------|---------|------------|
|`id`|integer|Id of the region|Primary Key|
|`name`|varchar(255)|Region name in Vietnamese||
|`name_en`|varchar(255)|Region name in English||
|`code_name`|varchar(255)|Code name, derived from Vietnamese name, written in lowercase, underscored||
|`code_name_en`|varchar(255)|Code name, derived from English name, written in lowercase, underscored||

#### Data preview

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

### `administrative_units` table

![VN Units](https://i.imgur.com/j35ELsL.png)  

The `administrative_units` table contains a list of administrative units with `id` sorted by the tier level from biggest to the smallest unit.

#### Table definition

|Column|Data type|Meaning|Constraint|
|------|-----------|---------|------------|
|`id`|integer|Id of the administrative unit|Primary Key|
|`full_name`|varchar(255)|Full name of the administrative unit in Vietnamese||
|`full_name_en`|varchar(255)|Full name of the administrative unit in English||
|`short_name`|varchar(255)|Short name of the administrative unit in Vietnamese||
|`short_name_en`|varchar(255)|Short name of the administrative unit in English||
|`code_name`|varchar(255)|Code name, derived from Vietnamese `full_name`, written in lowercase, underscored||
|`code_name_en`|varchar(255)|Code name, derived from English `full_name_en`, written in lowercase, underscored||

#### Data preview

|id|full_name|full_name_en|short_name|short_name_en|code_name|code_name_en|
|--|---------|------------|----------|-------------|---------|------------|
|1|Thành phố trực thuộc trung ương|Municipality|Thành phố|City|thanh_pho_truc_thuoc_trung_uong|municipality|
|2|Tỉnh|Province|Tỉnh|Province|tinh|province|
|3|Thành phố thuộc thành phố trực thuộc trung ương|Municipal city|Thành phố|City|thanh_pho_thuoc_thanh_pho_truc_thuoc_trung_uong|municipal_city|
|4|Thành phố thuộc tỉnh|Provincial city|Thành phố|City|thanh_pho_thuoc_tinh|provincial_city|
|5|Quận|Urban district|Quận|District|quan|urban_district|
|6|Thị xã|District-level town|Thị xã|Town|thi_xa|district_level_town|
|7|Huyện|District|Huyện|District|huyen|district|
|8|Phường|Ward|Phường|Ward|phuong|ward|
|9|Thị trấn|Commune-level town|Thị trấn|Township|thi_tran|commune_level_town|
|10|Xã|Commune|Xã|Commune|xa|commune|

### `provinces` table
![Provincial level](https://i.imgur.com/wNgbRqb.jpg)  
The `provinces` table contains a list of **first administrative tier - the provincial level** units, includes **63** municipalities and provinces.  
The `code` key and `full_name` are based on the original CSV file.

#### Table definition

|Column|Data type|Meaning|Constraint|
|------|-----------|---------|------------|
|`code`|varchar(20)|The official unit code, defined by government |Primary Key|
|`name`|varchar(255)|Name in Vietnamese||
|`name_en`|varchar(255)|Name of in English||
|`full_name`|varchar(255)|Full name in Vietnamese, includes the administrative unit name||
|`full_name_en`|varchar(255)|Full name in English, includes the administrative unit name||
|`code_name`|varchar(255)|Code name, derived from `name`, written in lowercase, underscored||
|`administrative_unit_id`|integer|The administrative unit id of this record|Foreign Key, references to `administrative_units.id` |
|`administrative_region_id`|integer|The geographical region this this record belongs to|Foreign Key, references to `administrative_regions.id`|

#### Data preview

|code|name|name_en|full_name|full_name_en|code_name|administrative_unit_id|administrative_region_id|
|----|----|-------|---------|------------|---------|----------------------|------------------------|
|01|Hà Nội|Ha Noi|Thành phố Hà Nội|Ha Noi City|ha_noi|1|3|
|30|Hải Dương|Hai Duong|Tỉnh Hải Dương|Hai Duong Province|hai_duong|2|3|
|46|Thừa Thiên Huế|Thua Thien Hue|Tỉnh Thừa Thiên Huế|Thua Thien Hue Province|thua_thien_hue|2|4|
|48|Đà Nẵng|Da Nang|Thành phố Đà Nẵng|Da Nang City|da_nang|1|5|
|79|Hồ Chí Minh|Ho Chi Minh|Thành phố Hồ Chí Minh|Ho Chi Minh City|ho_chi_minh|1|7|
|..|...........|...........|.....................|................|...........|..|..|

### `districts` table

![District level](https://i.imgur.com/B0OKHvB.jpg)
The `districts` table contains a list of **second administrative tier - the district level** units, includes **705** municipal city, urban districts, district-level towns, districts and provincial cities.  
The `code` key and `full_name` are based on the original CSV file.

#### Table definition

|Column|Data type|Meaning|Constraint|
|------|-----------|---------|------------|
|`code`|varchar(20)|The official unit code, defined by government |Primary Key|
|`name`|varchar(255)|Name in Vietnamese||
|`name_en`|varchar(255)|Name of in English||
|`full_name`|varchar(255)|Full name in Vietnamese, includes the administrative unit name||
|`full_name_en`|varchar(255)|Full name in English, includes the administrative unit name||
|`code_name`|varchar(255)|Code name, derived from `name`, written in lowercase, underscored||
|`province_code`|integer|The `province` this record belongs to|Foreign Key, references to `provinces.code`|
|`administrative_unit_id`|integer|The administrative unit id of this record|Foreign Key, references to `administrative_units.id` |

#### Data preview

|code|name|name_en|full_name|full_name_en|code_name|province_code|administrative_unit_id|
|----|----|-------|---------|------------|---------|-------------|----------------------|
|001|Ba Đình|Ba Dinh|Quận Ba Đình|Ba Dinh District|ba_dinh|01|5|
|002|Hoàn Kiếm|Hoan Kiem|Quận Hoàn Kiếm|Hoan Kiem District|hoan_kiem|01|5|
|003|Tây Hồ|Tay Ho|Quận Tây Hồ|Tay Ho District|tay_ho|01|5|
|004|Long Biên|Long Bien|Quận Long Biên|Long Bien District|long_bien|01|5|
|005|Cầu Giấy|Cau Giay|Quận Cầu Giấy|Cau Giay District|cau_giay|01|5|
|...|........|........|.............|.................|........|..|..|

### `wards` table
![Commune level](https://i.imgur.com/B5w1adp.jpg)
The `wards` table contains a list of **third administrative tier - the commune level** units, includes **10599** wards, communes and commune-level towns.  
The `code` key and `full_name` are based on the original CSV file.

#### Table definition

|Column|Data type|Meaning|Constraint|
|------|-----------|---------|------------|
|`code`|varchar(20)|The official unit code, defined by government |Primary Key|
|`name`|varchar(255)|Name in Vietnamese||
|`name_en`|varchar(255)|Name of in English||
|`full_name`|varchar(255)|Full name in Vietnamese, includes the administrative unit name||
|`full_name_en`|varchar(255)|Full name in English, includes the administrative unit name||
|`code_name`|varchar(255)|Code name, derived from `name`, written in lowercase, underscored||
|`district_code`|integer|The `district` this record belongs to|Foreign Key, references to `districts.code`|
|`administrative_unit_id`|integer|The administrative unit id of this record|Foreign Key, references to `administrative_units.id` |

#### Data preview

|code|name|name_en|full_name|full_name_en|code_name|district_code|administrative_unit_id|
|----|----|-------|---------|------------|---------|-------------|----------------------|
|25942|Dĩ An|Di An|Phường Dĩ An|Di An Ward|di_an|724|8|
|25945|Tân Bình|Tan Binh|Phường Tân Bình|Tan Binh Ward|tan_binh|724|8|
|25948|Tân Đông Hiệp|Tan Dong Hiep|Phường Tân Đông Hiệp|Tan Dong Hiep Ward|tan_dong_hiep|724|8|
|25951|Bình An|Binh An|Phường Bình An|Binh An Ward|binh_an|724|8|
|25954|Bình Thắng|Binh Thang|Phường Bình Thắng|Binh Thang Ward|binh_thang|724|8|
|-----|-----|-------|---------|-----------|----------|---|--|

## Sample Queries

You can easily create query to get all the kind of data you need since the tables are clearly referenced between each others.  
Here is some sample queries to start with:

### Get all the provinces in a geographical region

Get all provinces in **South Central Coast region** (`id` = 5)

```sql
SELECT p.code, p."name" , p.full_name , p.full_name_en ,au.full_name as administrative_unit_name
FROM provinces p
INNER JOIN administrative_units au 
ON p.administrative_unit_id = au.id 
WHERE p.administrative_region_id = 5
ORDER BY code;
```

|code|name|full_name|full_name_en|administrative_unit_name|
|----|----|---------|------------|------------------------|
|48|Đà Nẵng|Thành phố Đà Nẵng|Da Nang City|Thành phố trực thuộc trung ương|
|49|Quảng Nam|Tỉnh Quảng Nam|Quang Nam Province|Tỉnh|
|51|Quảng Ngãi|Tỉnh Quảng Ngãi|Quang Ngai Province|Tỉnh|
|52|Bình Định|Tỉnh Bình Định|Binh Dinh Province|Tỉnh|
|54|Phú Yên|Tỉnh Phú Yên|Phu Yen Province|Tỉnh|
|56|Khánh Hòa|Tỉnh Khánh Hòa|Khanh Hoa Province|Tỉnh|
|58|Ninh Thuận|Tỉnh Ninh Thuận|Ninh Thuan Province|Tỉnh|
|60|Bình Thuận|Tỉnh Bình Thuận|Binh Thuan Province|Tỉnh|


### Get all districts under a province

Get all districts under **Khánh Hoà province**

```sql
SELECT d.code, d."name" , d.full_name , d.full_name_en ,au.full_name as administrative_unit_name
FROM districts d 
INNER JOIN administrative_units au 
ON d.administrative_unit_id = au.id
WHERE d.province_code = '56' -- Khanh Hoa province code
ORDER BY d.code;
```

|code|name|full_name|full_name_en|administrative_unit_name|
|----|----|---------|------------|------------------------|
|568|Nha Trang|Thành phố Nha Trang|Nha Trang City|Thành phố thuộc tỉnh|
|569|Cam Ranh|Thành phố Cam Ranh|Cam Ranh City|Thành phố thuộc tỉnh|
|570|Cam Lâm|Huyện Cam Lâm|Cam Lam District|Huyện|
|571|Vạn Ninh|Huyện Vạn Ninh|Van Ninh District|Huyện|
|572|Ninh Hòa|Thị xã Ninh Hòa|Ninh Hoa Town|Thị xã|
|573|Khánh Vĩnh|Huyện Khánh Vĩnh|Khanh Vinh District|Huyện|
|574|Diên Khánh|Huyện Diên Khánh|Dien Khanh District|Huyện|
|575|Khánh Sơn|Huyện Khánh Sơn|Khanh Son District|Huyện|
|576|Trường Sa|Huyện Trường Sa|Truong Sa District|Huyện|

### Get wards under a district

Get all wards of **Ninh Hoa town**
```sql
SELECT w.code, w."name" , w.full_name , w.full_name_en ,au.full_name as administrative_unit_name
FROM wards w 
INNER JOIN administrative_units au 
ON w.administrative_unit_id = au.id
WHERE w.district_code = '572' -- Ninh Hoa town code
ORDER BY w.code;
```

|code|name|full_name|full_name_en|administrative_unit_name|
|----|----|---------|------------|------------------------|
|22528|Ninh Hiệp|Phường Ninh Hiệp|Ninh Hiep Ward|Phường|
|22531|Ninh Sơn|Xã Ninh Sơn|Ninh Son Commune|Xã|
|22534|Ninh Tây|Xã Ninh Tây|Ninh Tay Commune|Xã|
|22537|Ninh Thượng|Xã Ninh Thượng|Ninh Thuong Commune|Xã|
|22540|Ninh An|Xã Ninh An|Ninh An Commune|Xã|
|22543|Ninh Hải|Phường Ninh Hải|Ninh Hai Ward|Phường|
|22546|Ninh Thọ|Xã Ninh Thọ|Ninh Tho Commune|Xã|
|-----|--------|-----------|------|-----|
(the rest of rows are removed for brevity)

## FAQ

### What is the original data source that this repository develope from?

The original data source is the CSV file from the [General Statistics Office of Vietnam website][source danhmuchanhchinh gov].  
You can go to the site, tick on the checkbox **Quận Huyện, Phường Xã**, then click the **Xuất Excel** button to download the CSV file.

### How are the primary keys defined?

|Table|Primary Key|
|-----|-----------|
|`administrative_regions`|Key: `id`. Starting from `1` to `8`, follow the geographical location order from North to South
|`administrative_units`|Key: `id`. Starting from `1` to `10`, follow the tier order from biggest unit to smallest unit
|`provinces`|Key: `code`. Officially referenced from government unit code
|`districts`|Key: `code`. Officially referenced from government unit code
|`wards`|Key: `code`. Officially referenced from government unit code

### I cannot find districts 2, 9 and Thu Duc?

Districts 2, 9 and Thu Duc were merged into a single Thu Duc Municipal city, directly under **Ho Chi Minh city** recently. Hence all their
wards are binded directly under district entity `Thu Duc city`, code `769`.

### How to update the existing dataset?

The government may gradually public new decree to change the administrative unit structure overtime.

If you want to check the current up-to-date status of your provinces dataset, simply execute the [patch_checker.sql](patch/patch_checker.sql) script, and apply the missing decree patch (if any) in the [patch directory](./patch/).

Sample Output from the `patch_checker` script:

|nghidinh_469_nq_ubtvqh15|vietnamese_provinces_dataset_up_to_date|
|------------------------|---------------------------------------|
|true|true|

Or you can completely refresh the existing Vietnamese provinces table in your database by unlink all of your references, then drop these provinces tables and re-create and import all the provinces dataset, then re-establish the relationship between your data and the provinces table again.

### I saw some issues in the SQL patch?

If you see any improvement that can be made, please kindly [Open a issue](https://github.com/ThangLeQuoc/VietnameseProvincesDatabase/issues) and write down your finding. Or even better by [Create a Pull Request](https://github.com/ThangLeQuoc/VietnameseProvincesDatabase/pulls).
Any contribution is welcomed.

##### Reference
Vietnam Map in the banner by [vietcentertourist](https://vietcentertourist.com/assets/images/vietnam.png)


[source danhmuchanhchinh gov]: https://danhmuchanhchinh.gso.gov.vn/
[source government decree]: https://danhmuchanhchinh.gso.gov.vn/NghiDinh.aspx
[source goverment API]: https://danhmuchanhchinh.gso.gov.vn/DMDVHC.asmx
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
