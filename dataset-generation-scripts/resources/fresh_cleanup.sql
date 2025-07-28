-- DROP the GIS relevant table first to avoid constraint problem
DROP TABLE IF EXISTS sapnhap_wards;
DROP TABLE IF EXISTS sapnhap_provinces; 

DROP TABLE IF EXISTS wards_tmp;
DROP TABLE IF EXISTS districts_tmp;
DROP TABLE IF EXISTS provinces_tmp;

DROP TABLE IF EXISTS wards_tmp_seed;
DROP TABLE IF EXISTS provinces_tmp_seed;

DROP TABLE IF EXISTS wards;
DROP TABLE IF EXISTS districts;
DROP TABLE IF EXISTS provinces;
DROP TABLE IF EXISTS administrative_units;
DROP TABLE IF EXISTS administrative_regions;
