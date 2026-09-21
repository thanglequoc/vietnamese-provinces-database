/*
 * Release: v5.2.0
 *
 * 1) Promote Bắc Ninh (province code 24) to a centrally-governed city
 *    (Thành phố Bắc Ninh) per Nghị quyết 39/2026/QH16.
 *
 *    Data change: provinces.code = '24'
 *      full_name              : 'Tỉnh Bắc Ninh'    -> 'Thành phố Bắc Ninh'
 *      full_name_en           : 'Bac Ninh Province' -> 'Bac Ninh City'
 *      administrative_unit_id : 2 (Tỉnh) -> 1 (Thành phố trực thuộc trung ương)
 *
 * 2) Establish 12 wards of Bắc Ninh province as phường (wards) per
 *    Nghị quyết 388/NQ-UBTVQH16. The following existing xã (communes) are
 *    reclassified from commune (administrative_unit_id = 4) to ward
 *    (administrative_unit_id = 3); their code, name, name_en, code_name,
 *    province_code and postal_code are unchanged:
 *
 *      07294 Bố Hạ          07840 Hiệp Hoà
 *      07375 Lạng Giang     09193 Yên Phong
 *      07399 Kép            09292 Phù Lãng
 *      07444 Lục Nam        09313 Chi Lăng
 *      09454 Gia Bình       09319 Tiên Du
 *      09475 Nhân Thắng     09496 Lương Tài
 *
 * 3) Refresh the single-row dataset metadata table `vn_provinces_metadata`
 *    with this release's version/decree/timestamp.
 *
 * PostgreSQL only. GIS data is not covered by this patch.
 */

-- 1) Promote Bắc Ninh (province 24) to a centrally-governed city
UPDATE provinces
SET full_name = 'Thành phố Bắc Ninh',
    full_name_en = 'Bac Ninh City',
    administrative_unit_id = 1
WHERE code = '24';

-- 2) Reclassify 12 Bắc Ninh communes as wards (phường)
UPDATE wards
SET full_name = 'Phường Bố Hạ',
    full_name_en = 'Bo Ha Ward',
    administrative_unit_id = 3
WHERE code = '07294';

UPDATE wards
SET full_name = 'Phường Lạng Giang',
    full_name_en = 'Lang Giang Ward',
    administrative_unit_id = 3
WHERE code = '07375';

UPDATE wards
SET full_name = 'Phường Kép',
    full_name_en = 'Kep Ward',
    administrative_unit_id = 3
WHERE code = '07399';

UPDATE wards
SET full_name = 'Phường Lục Nam',
    full_name_en = 'Luc Nam Ward',
    administrative_unit_id = 3
WHERE code = '07444';

UPDATE wards
SET full_name = 'Phường Hiệp Hoà',
    full_name_en = 'Hiep Hoa Ward',
    administrative_unit_id = 3
WHERE code = '07840';

UPDATE wards
SET full_name = 'Phường Yên Phong',
    full_name_en = 'Yen Phong Ward',
    administrative_unit_id = 3
WHERE code = '09193';

UPDATE wards
SET full_name = 'Phường Phù Lãng',
    full_name_en = 'Phu Lang Ward',
    administrative_unit_id = 3
WHERE code = '09292';

UPDATE wards
SET full_name = 'Phường Chi Lăng',
    full_name_en = 'Chi Lang Ward',
    administrative_unit_id = 3
WHERE code = '09313';

UPDATE wards
SET full_name = 'Phường Tiên Du',
    full_name_en = 'Tien Du Ward',
    administrative_unit_id = 3
WHERE code = '09319';

UPDATE wards
SET full_name = 'Phường Gia Bình',
    full_name_en = 'Gia Binh Ward',
    administrative_unit_id = 3
WHERE code = '09454';

UPDATE wards
SET full_name = 'Phường Nhân Thắng',
    full_name_en = 'Nhan Thang Ward',
    administrative_unit_id = 3
WHERE code = '09475';

UPDATE wards
SET full_name = 'Phường Lương Tài',
    full_name_en = 'Luong Tai Ward',
    administrative_unit_id = 3
WHERE code = '09496';

-- 3) Refresh dataset metadata (single-row release descriptor)
DELETE FROM vn_provinces_metadata;
INSERT INTO vn_provinces_metadata(dataset_version, latest_decree, generated_at)
VALUES ('v5.2.0', '388/NQ-UBTVQH16', '2026-09-20 13:13:17');
