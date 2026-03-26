-- Enable Postgis Extension
CREATE EXTENSION IF NOT EXISTS postgis;

DROP TABLE IF EXISTS sapnhap_geojson_objects;

-- SQL Script for sapnhap_geojson_objects
-- Generated from: ./donvi_tinhthanh.json
-- Date: 2025-03-15
-- Using PostgreSQL Bulk INSERT for performance

-- Create table if not exists
CREATE TABLE IF NOT EXISTS sapnhap_geojson_objects (
    ma TEXT PRIMARY KEY,
    ten TEXT NOT NULL,
    magoc TEXT,
    malk TEXT,
    truocsapnhap TEXT,
    vn_ds_province_code varchar(20),
    vn_ds_ward_code varchar(20),

    bbox_wkt TEXT,
    bbox geometry(Polygon, 4326) GENERATED ALWAYS AS (ST_GeomFromText(bbox_wkt, 4326)) STORED,
    geom_wkt TEXT,
    geom geometry(Multipolygon, 4326) GENERATED ALWAYS AS (ST_GeomFromText(geom_wkt, 4326)) STORED,
    
    CONSTRAINT fk_sapnhap_geojson_objects_parent
        FOREIGN KEY (magoc)
        REFERENCES sapnhap_geojson_objects(ma)
        ON DELETE SET NULL
        ON UPDATE CASCADE,
    
    CONSTRAINT fk_sapnhap_geojson_objects_province
        FOREIGN KEY (vn_ds_province_code)
        REFERENCES provinces_tmp(code)
        ON DELETE SET NULL
        ON UPDATE CASCADE,
    
    CONSTRAINT fk_sapnhap_geojson_objects_ward
        FOREIGN KEY (vn_ds_ward_code)
        REFERENCES wards_tmp(code)
        ON DELETE SET NULL
        ON UPDATE CASCADE
);
