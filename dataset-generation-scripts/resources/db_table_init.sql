DROP TABLE IF EXISTS wards_tmp;
DROP TABLE IF EXISTS districts_tmp;
DROP TABLE IF EXISTS provinces_tmp;
DROP TABLE IF EXISTS wards;
DROP TABLE IF EXISTS districts;
DROP TABLE IF EXISTS provinces;

DROP TABLE IF EXISTS administrative_units;
DROP TABLE IF EXISTS administrative_regions;
DROP TABLE IF EXISTS vn_gis;

CREATE EXTENSION IF NOT EXISTS postgis;

-- CREATE administrative_regions TABLE
CREATE TABLE administrative_regions (
	id integer NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NOT NULL,
	code_name varchar(255) NULL,
	code_name_en varchar(255) NULL,
	CONSTRAINT administrative_regions_pkey PRIMARY KEY (id)
);


-- CREATE administrative_units TABLE
CREATE TABLE administrative_units (
	id integer NOT NULL,
	full_name varchar(255) NULL,
	full_name_en varchar(255) NULL,
	short_name varchar(255) NULL,
	short_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	code_name_en varchar(255) NULL,
	CONSTRAINT administrative_units_pkey PRIMARY KEY (id)
);


-- CREATE provinces TABLE
CREATE TABLE provinces_tmp (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NOT NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	administrative_unit_id integer NULL,
	administrative_region_id integer NULL,
	CONSTRAINT provinces_tmp_pkey PRIMARY KEY (code)
);


-- provinces foreign keys

ALTER TABLE provinces_tmp ADD CONSTRAINT provinces_tmp_administrative_region_id_fkey FOREIGN KEY (administrative_region_id) REFERENCES administrative_regions(id);
ALTER TABLE provinces_tmp ADD CONSTRAINT provinces_tmp_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);
CREATE INDEX idx_provinces_tmp_region ON provinces_tmp(administrative_region_id);
CREATE INDEX idx_provinces_tmp_unit ON provinces_tmp(administrative_unit_id);


-- CREATE districts TABLE
CREATE TABLE districts_tmp (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	province_code varchar(20) NULL,
	administrative_unit_id integer NULL,
	CONSTRAINT districts_tmp_pkey PRIMARY KEY (code)
);


-- districts foreign keys

ALTER TABLE districts_tmp ADD CONSTRAINT districts_tmp_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);
ALTER TABLE districts_tmp ADD CONSTRAINT districts_tmp_province_code_fkey FOREIGN KEY (province_code) REFERENCES provinces_tmp(code);

CREATE INDEX idx_districts_tmp_province ON districts_tmp(province_code);
CREATE INDEX idx_districts_tmp_unit ON districts_tmp(administrative_unit_id);



-- CREATE wards TABLE
CREATE TABLE wards_tmp (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	district_code varchar(20) NULL,
	administrative_unit_id integer NULL,
	CONSTRAINT wards_tmp_pkey PRIMARY KEY (code)
);


-- wards foreign keys

ALTER TABLE wards_tmp ADD CONSTRAINT wards_tmp_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);
ALTER TABLE wards_tmp ADD CONSTRAINT wards_tmp_district_code_fkey FOREIGN KEY (district_code) REFERENCES districts_tmp(code);
CREATE INDEX idx_wards_tmp_district ON wards_tmp(district_code);
CREATE INDEX idx_wards_tmp_unit ON wards_tmp(administrative_unit_id);


-- -----------------------------------------------------------------

/* Create current dataset tables */
-- CREATE provinces TABLE
CREATE TABLE provinces (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NOT NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	administrative_unit_id integer NULL,
	administrative_region_id integer NULL,
	CONSTRAINT provinces_pkey PRIMARY KEY (code)
);
ALTER TABLE provinces ADD CONSTRAINT provinces_administrative_region_id_fkey FOREIGN KEY (administrative_region_id) REFERENCES administrative_regions(id);
ALTER TABLE provinces ADD CONSTRAINT provinces_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);

-- CREATE districts TABLE
CREATE TABLE districts (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	province_code varchar(20) NULL,
	administrative_unit_id integer NULL,
	CONSTRAINT districts_pkey PRIMARY KEY (code)
);

ALTER TABLE districts ADD CONSTRAINT districts_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);
ALTER TABLE districts ADD CONSTRAINT districts_province_code_fkey FOREIGN KEY (province_code) REFERENCES provinces(code);

-- CREATE wards TABLE
CREATE TABLE wards (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	district_code varchar(20) NULL,
	administrative_unit_id integer NULL,
	CONSTRAINT wards_pkey PRIMARY KEY (code)
);
ALTER TABLE wards ADD CONSTRAINT wards_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);
ALTER TABLE wards ADD CONSTRAINT wards_district_code_fkey FOREIGN KEY (district_code) REFERENCES districts(code);

-- CREATE vn_gis TABLE
CREATE TABLE vn_gis (
	code varchar(20) NOT NULL,
	level varchar(20) NOT NULL,
	bbox geometry(Polygon, 4326),
	gis_geom geometry(Multipolygon, 4326),
	CONSTRAINT vn_gis_pkey PRIMARY KEY (code)
)
