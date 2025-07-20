![Repository Banner](https://i.imgur.com/6s3XsAA.png)
![Made in Vietnam](https://raw.githubusercontent.com/webuild-community/badge/master/svg/made.svg)

[Đọc phiên bản tiếng Việt](README_vi.md)

# Vietnamese Provinces Database

A complete SQL (and also non-SQL) databases of Vietnamese administrative units, includes all **34 Vietnamese provinces** and associated districts, wards sub-divisions.  
Data is updated as of the most recent effective decree: [19/2025/QĐ-TTg][source government decree]  

If you find this repository helpful, please consider giving it a ⭐ — it helps us stay motivated to keep improving and delivering valuable tools for the community. Also, starring the repo makes it easier to stay updated with future releases.

## Overview

The author(s) of this repository is not associated with the **General Statistics Office of Vietnam**, nor the Vietnamese government.  
The data of provinces and wards are created base on the [API province data provided by the General Statistics Office of Vietnam website][source goverment API].  
**Disclaimer:** Since the GSO SOAP API is not updated with the most recent 34 provinces breaking change. The latest data are purely rely on the document decree [19/2025/QĐ-TTg][decree 19/2025/QĐ-TTg]

This dataset also include additional information apart from the original provinces, wards data from the original data. Please see section [Additional change make by this repository](#additional-change-make-by-this-repository).  


### Dataset releases and Government issued decrees

The Vietnamese Government may issue decree from time to time to adjust the administrative unit structure. You can track the latest issued decrees [here][decree issued page].  

The following table contains a list of issued decrees and their effective dates, tracked from the earliest version of this dataset.

|Issued Decree|Issued on |Effect from|Release Version|
|-------------|-----------|-------------|---------------|
|[19/2025/QĐ-TTg][decree 19/2025/QĐ-TTg]|30/06/2025|01/07/2025|v3.0.0|
|Grammar-correction, data cutoff before 19/2025/QĐ-TTg|15/01/2025|01/03/2025|v2.4.1
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


### Additional Changes Made by This Repository

- Added `administrative_regions` table  
- Added `administrative_units` table  
- Assigned administrative units to province and ward data  
- Generated English names for provinces and wards, offering both full and short forms  
- Generated code names (slugs) for provinces and wards  

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

![VN_administrative_units db](https://i.imgur.com/XEIgaXV.png)

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

![VN Units](https://i.imgur.com/U0Warh3.png)  

The `administrative_units` table contains a list of administrative units with `id` sorted by two tier levels.

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
|3|Phường|Ward|Phường|Ward|phuong|ward|
|4|Xã|Commune|Xã|Commune|xa|commune|
|5|Đặc khu tại hải đảo|Special administrative region|Đặc khu|Special administrative region|dac_khu|special_administrative_region|

### `provinces` table
![Provincial level](https://i.imgur.com/cLTRHkf.png)  

The `provinces` table contains a list of **first administrative tier - the provincial level** units, includes **34** municipalities and provinces.  

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

#### Data preview

|code|name|name_en|full_name|full_name_en|code_name|administrative_unit_id|
|----|----|-------|---------|------------|---------|----------------------|
|01|Hà Nội|Ha Noi|Thành phố Hà Nội|Ha Noi City|ha_noi|1|
|56|Khánh Hòa|Khanh Hoa|Tỉnh Khánh Hòa|Khanh Hoa Province|khanh_hoa|2|
|79|Hồ Chí Minh|Ho Chi Minh|Thành phố Hồ Chí Minh|Ho Chi Minh City|ho_chi_minh|1|
|96|Cà Mau|Ca Mau|Tỉnh Cà Mau|Ca Mau Province|ca_mau|2|
|..|...........|...........|.....................|................|...........|..|..|

### `wards` table
![Commune level](https://i.imgur.com/B5w1adp.jpg)
The `wards` table contains a list of **second administrative tier - the commune level** units, includes **3321** wards, communes and special administrative region.  

#### Table definition

|Column|Data type|Meaning|Constraint|
|------|-----------|---------|------------|
|`code`|varchar(20)|The official unit code, defined by government |Primary Key|
|`name`|varchar(255)|Name in Vietnamese||
|`name_en`|varchar(255)|Name of in English||
|`full_name`|varchar(255)|Full name in Vietnamese, includes the administrative unit name||
|`full_name_en`|varchar(255)|Full name in English, includes the administrative unit name||
|`code_name`|varchar(255)|Code name, derived from `name`, written in lowercase, underscored||
|`province_code`|varchar(20)|The `province` this record belongs to|Foreign Key, references to `provinces.code`|
|`administrative_unit_id`|integer|The administrative unit id of this record|Foreign Key, references to `administrative_units.id` |

#### Data preview

|code|name|name_en|full_name|full_name_en|code_name|province_code|administrative_unit_id|
|----|----|-------|---------|------------|---------|-------------|----------------------|
|25920|Tân Hiệp|Tan Hiep|Phường Tân Hiệp|Tan Hiep Ward|tan_hiep|79|3|
|25942|Dĩ An|Di An|Phường Dĩ An|Di An Ward|di_an|79|3|
|25945|Tân Đông Hiệp|Tan Dong Hiep|Phường Tân Đông Hiệp|Tan Dong Hiep Ward|tan_dong_hiep|79|3|
|25951|Đông Hòa|Dong Hoa|Phường Đông Hòa|Dong Hoa Ward|dong_hoa|79|3|
|25966|Lái Thiêu|Lai Thieu|Phường Lái Thiêu|Lai Thieu Ward|lai_thieu|79|3|
|25969|Thuận Giao|Thuan Giao|Phường Thuận Giao|Thuan Giao Ward|thuan_giao|79|3|
|25975|An Phú|An Phu|Phường An Phú|An Phu Ward|an_phu|79|3|


## Sample Queries

You can easily create query to get all the kind of data you need since the tables are clearly referenced between each others.  
Here is some sample queries to start with:

### Get all wards under a province

Get all wards under **Khánh Hoà province**

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

## Non-SQL Formats

Along with traditional SQL dataset, the Vietnamese Provinces Database also comes with non-sql data format, this includes

- **JSON** format (feature full, simplified and vn_only_simplified versions)
- **MongoDB**
- **Redis**

## FAQ

### What is the original data source that this repository develope from?
The data of provinces and wards are created base on the [API province data provided by the General Statistics Office of Vietnam website][source goverment API].  
**Disclaimer:** Since the GSO SOAP API is not updated with the most recent 34 provinces breaking change. The latest data are purely rely on the document [19/2025/QĐ-TTg][decree 19/2025/QĐ-TTg]

### How are the primary keys defined?

|Table|Primary Key|
|-----|-----------|
|`administrative_regions`|Key: `id`. Starting from `1` to `8`, follow the geographical location order from North to South
|`administrative_units`|Key: `id`. Starting from `1` to `5`, follow the tier order from biggest unit to smallest unit
|`provinces`|Key: `code`. Officially referenced from government unit code
|`wards`|Key: `code`. Officially referenced from government unit code

### The province - administrative region relationship is dropped from v3.0.0?

After the provinces merge down to 37 provinces, the new province e.g: Phú Thọ, which is formed from Vĩnh Phúc, Phú Thọ and Hoà Bình that previously span across 3 different regions, so it's no longer applicable to determine which region does the new province belongs to.

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
[decree 19/2025/QĐ-TTg]: https://www.nso.gov.vn/default/2025/07/quyet-dinh-ban-hanh-bang-danh-muc-va-ma-so-cac-don-vi-hanh-chinh-viet-nam/
