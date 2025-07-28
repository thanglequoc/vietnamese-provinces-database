DROP TABLE IF EXISTS sapnhap_wards;
DROP TABLE IF EXISTS sapnhap_provinces;

CREATE TABLE sapnhap_provinces (
  id INTEGER PRIMARY KEY,
  mahc INTEGER UNIQUE,
  tentinh VARCHAR(100) NOT NULL,
  dientichkm2 VARCHAR(50),
  dansonguoi VARCHAR(50),
  trungtamhc VARCHAR(50),
  kinhdo DOUBLE PRECISION,
  vido DOUBLE PRECISION,
  truocsapnhap VARCHAR(255),
  con VARCHAR(50),
  vn_province_code VARCHAR(50) NOT NULL REFERENCES provinces_tmp(code)
);

CREATE TABLE sapnhap_wards (
  id INTEGER PRIMARY KEY,
  matinh INTEGER,
  ma VARCHAR(50),
  tentinh VARCHAR(255),
  loai VARCHAR(50),
  tenhc VARCHAR(255),
  cay VARCHAR(50),
  dientichkm2 DOUBLE PRECISION,
  dansonguoi VARCHAR(50),
  trungtamhc TEXT,
  kinhdo DOUBLE PRECISION,
  vido DOUBLE PRECISION,
  truocsapnhap TEXT,
  maxa INTEGER,
  khoa VARCHAR(255),
  vn_ward_code VARCHAR(50) NOT NULL,

  -- Foreign keys
  CONSTRAINT fk_matinh FOREIGN KEY (matinh) REFERENCES sapnhap_provinces(mahc),
  CONSTRAINT fk_ward_code FOREIGN KEY (vn_ward_code) REFERENCES wards_tmp(code)
);
