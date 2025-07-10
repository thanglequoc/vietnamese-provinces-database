# Improvement plan

- Revise the script execution
- Introduce unit testing?
-

## Compare existing dataset and new dataset
- Compare strategy
- Pagination on the existing dataset, get a batch of n items.
- Query on the new dataset & compare the data.
- If the record does not found in the dataset -> unit has been remove (merge to another unit)
- There's a very rare chance that a new unit is introduced? How to prevent this?

## Integration Test?
https://www.ardanlabs.com/blog/2019/03/integration-testing-in-go-executing-tests-with-docker.html


GIS Data resource

Table
code - primary key, string
level - province/district, required string
bbox - required, polygon
gis_geom - multipolygon

Thing to do

- Generate bbox data first
- 

Format(Lat Long)
105.174641982 21.7004831250001 --> south west
105.292898491 21.8649141770001 --> north east

105.174641982 21.7004831250001
105.292898491 21.7004831250001
105.292898491 21.8649141770001
105.174641982 21.8649141770001
105.174641982 21.7004831250001

To avoid ring not closed exception, first point must repeated
 
https://gis.stackexchange.com/questions/321385/converting-multipolygon-field-stored-as-text-to-geometry-data-type-in-postgis-an
https://dbschema.com/2023/07/16/sqlserver/spatial-data-types/
Oracle database: https://hub.docker.com/r/gvenzl/oracle-free


https://oralytics.com/23c/23c-free-on-docker/

-- https://learn.microsoft.com/en-us/sql/t-sql/spatial-geography/spatial-types-geography?view=sql-server-ver16
-- long, lat
INSERT INTO vn_gis(code,level,bbox)
VALUES (
'02',
'province',
 geography::STGeomFromText('POLYGON((105.285004 20.564251, 106.020155 20.564251, 106.020155 21.385514, 105.285004 21.385514, 105.285004 20.564251))', 4326)
);

BEGIN
	DECLARE @point geography;
	SET @point = geography::STGeomFromText('POINT(106.020155 20.564251)', 4326);
	SELECT p.name, vg.*
	FROM vn_gis vg
	INNER JOIN provinces p 
	ON vg.code = p.code
	WHERE bbox.STIntersects(@point) = 1 OR bbox.STIntersects(@point) = 1
END

-- https://stackoverflow.com/questions/71585217/when-writing-a-sql-server-geography-value-the-shell-of-a-polygon-must-be-orient

-- SQL Server ring counter clockwise problem ?

```sql
BEGIN
	DECLARE @point geography;
	SET @point = geography::STGeomFromText('POINT(106.020155 20.564251)', 4326);
	-- SET @point = geography::Point(20.564251, 106.020155, 4326); -- Alternative
	SELECT p.name, vg.*
	FROM vn_gis vg
	INNER JOIN provinces p 
	ON vg.code = p.code
	WHERE bbox.STIntersects(@point) = 1 OR bbox.STIntersects(@point) = 1
END
```




CREATE ROLE thanglequoc WITH LOGIN SUPERUSER PASSWORD 'root';
CREATE USER 'thanglequoc'@'localhost' IDENTIFIED BY 'Root@1234';
GRANT ALL PRIVILEGES ON *.* TO 'thanglequoc'@'localhost' WITH GRANT OPTION;


--------------------------

Observe change:

Remove district 358: Huyen My Loc

Update Nam Phong

Mỹ Hưng 13729 + Mỹ Lộc 13708 = Hưng Lộc 13708

Lộc An 13702 + Văn Miếu 13675 = Trương Thi (13657)

Hạ Long 13633 + Thống Nhất 13681 = Quang Trung

Trần Tế Xương 13636 + Vị Hoàng 13639 = Vị Xuyên

Phan Đình Phùng 13660 + Nguyễn Du 13651 = Trần Hưng Đạo

Ngô Quyền 13663 + Trần Quang Khải 13678 = Năng Tĩnh 13672

Bà Triệu 13654 + Trần Đăng Ninh 13669 = Cửa Bắc (13669)

Mỹ Thành (13738) -> Mỹ Lộc (13735)



--- Vụ Bản
Tân Khánh + Minh Thuận -> Minh Tân

Tân Thành + Liên Bảo -> Thành Lợi

Yên Thành + Yên Nghĩa + Yên Trung -> Xã Trung Nghĩa

Yên Hưng + Yên Phú + Yên Phương -> Phú Hưng

Yên Minh + Yên Lợi + Yên Tân -> Tân Minh

Yên Quang + Yên Hồng + Yên Bằng -> Hồng Quang

--- Nam Trực
Nam Toàn + Nam Mỹ + Điền Xá -> Nam Điền

-- Xuân Trường
Xuân Đài + Xuân Phong + Xuân Thuỷ -> Xuân Giang
Xuân Hoà + Xuân Kiên + Xuân Tiến -> Xuân Phúc

Xuân Bắc + 

=====
