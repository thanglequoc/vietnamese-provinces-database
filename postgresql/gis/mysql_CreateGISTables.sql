DROP TABLE IF EXISTS gis_wards;
DROP TABLE IF EXISTS gis_provinces;

CREATE TABLE gis_provinces (
  id integer primary key generated always as identity,
  province_code varchar(20) NOT NULL,
  gis_server_id varchar(20),
  area_km2 numeric(12,5),
  bbox geometry(Polygon, 4326),
  geom geometry(Multipolygon, 4326)
);

ALTER TABLE gis_provinces ADD CONSTRAINT gis_provinces_province_code_fkey FOREIGN KEY (province_code) REFERENCES provinces(code);
CREATE INDEX idx_gis_provinces_province_code ON gis_provinces(province_code);
CREATE INDEX idx_gis_provinces_bbox ON gis_provinces USING gist (bbox);
CREATE INDEX idx_gis_provinces_geom ON gis_provinces USING gist (geom);

CREATE TABLE gis_wards (
  id integer primary key generated always as identity,
  ward_code varchar(20) NOT NULL,
  gis_server_id varchar(20),
  area_km2 numeric(12,5),
  bbox geometry(Polygon, 4326),
  geom geometry(Multipolygon, 4326)
);

ALTER TABLE gis_wards ADD CONSTRAINT gis_wards_ward_code_fkey FOREIGN KEY (ward_code) REFERENCES wards(code);
CREATE INDEX idx_gis_wards_ward_code ON gis_wards(ward_code);
CREATE INDEX idx_gis_wards_bbox ON gis_wards USING gist (bbox);
CREATE INDEX idx_gis_wards_geom ON gis_wards USING gist (geom);
