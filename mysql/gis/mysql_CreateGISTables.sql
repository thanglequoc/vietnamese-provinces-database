-- Drop tables if exist
DROP TABLE IF EXISTS gis_wards;
DROP TABLE IF EXISTS gis_provinces;

-- Provinces table
CREATE TABLE gis_provinces (
  id INT PRIMARY KEY AUTO_INCREMENT,
  province_code VARCHAR(20) NOT NULL,
  gis_server_id VARCHAR(20),
  area_km2 DOUBLE,
  bbox POLYGON NOT NULL SRID 4326,
  geom MULTIPOLYGON NOT NULL SRID 4326,
  CONSTRAINT gis_provinces_province_code_fkey FOREIGN KEY (province_code) REFERENCES provinces(code)
);

-- Indexes for provinces
CREATE INDEX idx_gis_provinces_province_code ON gis_provinces(province_code);
CREATE SPATIAL INDEX idx_gis_provinces_bbox ON gis_provinces(bbox);
CREATE SPATIAL INDEX idx_gis_provinces_geom ON gis_provinces(geom);

-- Wards table
CREATE TABLE gis_wards (
  id INT PRIMARY KEY AUTO_INCREMENT,
  ward_code VARCHAR(20) NOT NULL,
  gis_server_id VARCHAR(20),
  area_km2 DECIMAL(12,5),
  bbox POLYGON NOT NULL SRID 4326,
  geom MULTIPOLYGON NOT NULL SRID 4326,
  CONSTRAINT gis_wards_ward_code_fkey FOREIGN KEY (ward_code) REFERENCES wards(code)
);

-- Indexes for wards
CREATE INDEX idx_gis_wards_ward_code ON gis_wards(ward_code);
CREATE SPATIAL INDEX idx_gis_wards_bbox ON gis_wards(bbox);
CREATE SPATIAL INDEX idx_gis_wards_geom ON gis_wards(geom);