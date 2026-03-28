-- SQL Script to create temporary tables for GIS data comparison
-- This script creates temporary tables that will hold the data from SQL dump files
-- for comparison with the current sapnhap_provinces_gis and sapnhap_wards_gis tables

-- Drop temp tables if they exist
DROP TABLE IF EXISTS temp_sapnhap_provinces_gis_dump;
DROP TABLE IF EXISTS temp_sapnhap_wards_gis_dump;

-- Create temporary table for provinces dump
-- This mirrors the structure of sapnhap_provinces_gis but without generated columns
CREATE TEMP TABLE temp_sapnhap_provinces_gis_dump (
	stt INTEGER,
	ten VARCHAR(255),
	truocsn TEXT,
	gis_server_id VARCHAR(20),
	sapnhap_province_matinh INTEGER,
	bbox_wkt TEXT,
	geom_wkt TEXT
);

-- Create temporary table for wards dump
-- This mirrors the structure of sapnhap_wards_gis but without generated columns
CREATE TEMP TABLE temp_sapnhap_wards_gis_dump (
	stt INTEGER,
	ten VARCHAR(255),
	truocsn TEXT,
	gis_server_id VARCHAR(20),
	sapnhap_ward_maxa INTEGER,
	bbox_wkt TEXT,
	geom_wkt TEXT
);

-- Create indexes for efficient comparison
CREATE INDEX idx_temp_provinces_gis_id ON temp_sapnhap_provinces_gis_dump(gis_server_id);
CREATE INDEX idx_temp_wards_gis_id ON temp_sapnhap_wards_gis_dump(gis_server_id);

-- Note: Data will be loaded from dump files using separate INSERT statements
-- The dump files should be modified to insert into these temp tables instead of the original tables
