-- Persist the region & administrative records, since this data does not comes from DVHCVN API

-- DATA for administrative_regions
INSERT INTO administrative_regions (id,"name",name_en,code_name,code_name_en) VALUES
	 (1,'Đông Bắc Bộ','Northeast','dong_bac_bo','northest'),
	 (2,'Tây Bắc Bộ','Northwest','tay_bac_bo','northwest'),
	 (3,'Đồng bằng sông Hồng','Red River Delta','dong_bang_song_hong','red_river_delta'),
	 (4,'Bắc Trung Bộ','North Central Coast','bac_trung_bo','north_central_coast'),
	 (5,'Duyên hải Nam Trung Bộ','South Central Coast','duyen_hai_nam_trung_bo','south_central_coast'),
	 (6,'Tây Nguyên','Central Highlands','tay_nguyen','central_highlands'),
	 (7,'Đông Nam Bộ','Southeast','dong_nam_bo','southeast'),
	 (8,'Đồng bằng sông Cửu Long','Mekong River Delta','dong_bang_song_cuu_long','southwest');

-- DATA for administrative_units
INSERT INTO administrative_units (id,full_name,full_name_en,short_name,short_name_en,code_name,code_name_en) VALUES
	 -- Level 1
	 (1,'Thành phố trực thuộc trung ương','Municipality','Thành phố','City','thanh_pho_truc_thuoc_trung_uong','municipality'),
	 (2,'Tỉnh','Province','Tỉnh','Province','tinh','province'),
  -- Level 2
	 (3,'Phường','Ward','Phường','Ward','phuong','ward'),
	 (4,'Xã','Commune','Xã','Commune','xa','commune'),
	 (5,'Đặc khu tại hải đảo','Special administrative region','Đặc khu','Special administrative region','dac_khu','special_administrative_region');
