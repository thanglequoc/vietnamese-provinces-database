DROP TABLE IF EXISTS sapnhap_provinces;

CREATE TABLE sapnhap_provinces (
  id INTEGER PRIMARY KEY,
  mahc INTEGER,
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
