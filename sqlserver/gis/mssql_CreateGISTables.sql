-- Drop tables if they exist
DROP TABLE IF EXISTS gis_wards;
DROP TABLE IF EXISTS gis_provinces;

-- Provinces table
CREATE TABLE gis_provinces (
  id INT IDENTITY(1, 1) PRIMARY KEY,
  province_code NVARCHAR(20) NOT NULL,
  gis_server_id VARCHAR(20),
  area_km2 FLOAT,
  bbox geometry NOT NULL,
  -- polygon
  geom geometry NOT NULL,
  -- multipolygon
  CONSTRAINT fk_gis_provinces_province_code FOREIGN KEY (province_code) REFERENCES provinces(code)
);

-- Indexes for provinces
CREATE INDEX idx_gis_provinces_province_code ON gis_provinces(province_code);

CREATE SPATIAL INDEX idx_gis_provinces_bbox ON gis_provinces(bbox) USING GEOMETRY_GRID WITH (
  BOUNDING_BOX = (
    102.14148702,
    8.41365345,
    109.45948839,
    23.39330777
  ),
  GRIDS = (HIGH, HIGH, HIGH, HIGH),
  CELLS_PER_OBJECT = 16
);

CREATE SPATIAL INDEX idx_gis_provinces_geom ON gis_provinces(geom) USING GEOMETRY_GRID WITH (
  BOUNDING_BOX = (
    102.14148702,
    8.41365345,
    109.45948839,
    23.39330777
  ),
  GRIDS = (HIGH, HIGH, HIGH, HIGH),
  CELLS_PER_OBJECT = 16
);

-- Wards table
CREATE TABLE gis_wards (
  id INT IDENTITY(1, 1) PRIMARY KEY,
  ward_code NVARCHAR(20) NOT NULL,
  gis_server_id VARCHAR(20),
  area_km2 DECIMAL(12, 5),
  bbox geometry NOT NULL,
  -- polygon
  geom geometry NOT NULL,
  -- multipolygon
  CONSTRAINT fk_gis_wards_ward_code FOREIGN KEY (ward_code) REFERENCES wards(code)
);

-- Indexes for wards
CREATE INDEX idx_gis_wards_ward_code ON gis_wards(ward_code);

CREATE SPATIAL INDEX idx_gis_wards_bbox ON gis_wards(bbox) USING GEOMETRY_GRID WITH (
  BOUNDING_BOX = (
    102.14148702,
    8.41365345,
    109.45948839,
    23.39330777
  ),
  GRIDS = (HIGH, HIGH, HIGH, HIGH),
  CELLS_PER_OBJECT = 16
);

-- Spatial index for gis_wards
CREATE SPATIAL INDEX idx_gis_wards_geom ON gis_wards(geom) USING GEOMETRY_GRID WITH (
  BOUNDING_BOX = (
    102.14148702,
    8.41365345,
    109.45948839,
    23.39330777
  ),
  GRIDS = (HIGH, HIGH, HIGH, HIGH),
  CELLS_PER_OBJECT = 16
);