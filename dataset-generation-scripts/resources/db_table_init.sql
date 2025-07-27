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

-- CREATE provinces_tmp TABLE
CREATE TABLE provinces_tmp (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NOT NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	administrative_unit_id integer NULL,
	CONSTRAINT provinces_tmp_pkey PRIMARY KEY (code)
);


-- provinces foreign keys
ALTER TABLE provinces_tmp ADD CONSTRAINT provinces_tmp_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);
CREATE INDEX idx_provinces_tmp_unit ON provinces_tmp(administrative_unit_id);

-- CREATE wards_tmp TABLE
CREATE TABLE wards_tmp (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	province_code varchar(20) NULL,
	administrative_unit_id integer NULL,
	CONSTRAINT wards_tmp_pkey PRIMARY KEY (code)
);

-- wards_tmp foreign keys
ALTER TABLE wards_tmp ADD CONSTRAINT wards_tmp_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);
ALTER TABLE wards_tmp ADD CONSTRAINT wards_tmp_province_code_fkey FOREIGN KEY (province_code) REFERENCES provinces_tmp(code);
CREATE INDEX idx_wards_tmp_province ON wards_tmp(province_code);
CREATE INDEX idx_wards_tmp_tmp_unit ON wards_tmp(administrative_unit_id);


-- CREATE provinces_tmp_seed TABLE
CREATE TABLE provinces_tmp_seed (
    code varchar(20) NOT NULL,
    "name" varchar(255) NOT NULL,
    CONSTRAINT provinces_tmp_seed_pkey PRIMARY KEY (code)
);

-- CREATE wards_tmp_seed TABLE
CREATE TABLE wards_tmp_seed (
    code varchar(20) NOT NULL,
    "name" varchar(255) NOT NULL,
		province_code varchar(20) NOT NULL,
    CONSTRAINT wards_tmp_seed_pkey PRIMARY KEY (code)
);

-- wards_tmp_seed foreign keys
ALTER TABLE wards_tmp_seed ADD CONSTRAINT wards_tmp_seed_province_code_fkey FOREIGN KEY (province_code) REFERENCES provinces_tmp_seed(code);

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


-- CREATE wards TABLE
CREATE TABLE wards (
	code varchar(20) NOT NULL,
	"name" varchar(255) NOT NULL,
	name_en varchar(255) NULL,
	full_name varchar(255) NULL,
	full_name_en varchar(255) NULL,
	code_name varchar(255) NULL,
	province_code varchar(20) NULL,
	administrative_unit_id integer NULL,
	CONSTRAINT wards_pkey PRIMARY KEY (code)
);
ALTER TABLE wards ADD CONSTRAINT wards_administrative_unit_id_fkey FOREIGN KEY (administrative_unit_id) REFERENCES administrative_units(id);
ALTER TABLE wards ADD CONSTRAINT wards_province_code_fkey FOREIGN KEY (province_code) REFERENCES provinces(code);
