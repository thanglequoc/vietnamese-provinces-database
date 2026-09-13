/*
 * Release: v5.1.0
 *
 * 1) Promote Quảng Ninh (province code 22) to a centrally-governed city
 *    (Thành phố Quảng Ninh) per Nghị quyết 36/2026/QH16,
 *    issued 24/08/2026, effective 01/09/2026.
 *
 *    Data change: provinces.code = '22'
 *      full_name              : 'Tỉnh Quảng Ninh'    -> 'Thành phố Quảng Ninh'
 *      full_name_en           : 'Quang Ninh Province' -> 'Quang Ninh City'
 *      administrative_unit_id : 2 (Tỉnh) -> 1 (Thành phố trực thuộc trung ương)
 *
 *    All other columns of the province row are unchanged. No other
 *    administrative_regions, administrative_units, provinces or wards rows
 *    were added, changed, or removed in this release.
 *
 * 2) Add the new single-row dataset metadata table `vn_provinces_metadata`
 *    and seed it with this release's version/decree/timestamp.
 *
 * PostgreSQL only. GIS data is not covered by this patch.
 */

-- 1) Promote Quảng Ninh (province 22) to a centrally-governed city
UPDATE provinces
SET full_name = 'Thành phố Quảng Ninh',
    full_name_en = 'Quang Ninh City',
    administrative_unit_id = 1
WHERE code = '22';

-- 2) Dataset metadata table (single-row release descriptor)
CREATE TABLE IF NOT EXISTS vn_provinces_metadata (
	dataset_version varchar(50) NOT NULL,
	latest_decree varchar(100) NULL,
	generated_at timestamp NOT NULL
);

-- 3) Seed the single metadata row for this release
DELETE FROM vn_provinces_metadata;
INSERT INTO vn_provinces_metadata(dataset_version, latest_decree, generated_at)
VALUES ('v5.1.0', '36/2026/QH16', '2026-09-12 07:42:08');
