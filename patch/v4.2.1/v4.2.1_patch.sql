/*
 * Release: v4.2.1
 *
 * Correct the ward code of Xã Ba Chẽ, huyện Ba Chẽ, tỉnh Quảng Ninh (province 22).
 * The correct administrative code is 06970 (previously mis-signed as 06978).
 *
 * Data change: wards.code 06978 -> 06970
 * All other columns of the ward row are unchanged.
 *
 * If any of your tables reference the old code 06978, update those references
 * to 06970 as well.
 */
UPDATE wards SET code = '06970' WHERE code = '06978';
