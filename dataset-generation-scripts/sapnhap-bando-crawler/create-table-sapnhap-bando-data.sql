CREATE TABLE bando_co_dvch (
    ma TEXT PRIMARY KEY,
    ten TEXT NOT NULL,
    magoc TEXT,
    malk TEXT,
    truocsapnhap TEXT,
    
    CONSTRAINT fk_bando_co_dvch_parent
        FOREIGN KEY (magoc)
        REFERENCES bando_co_dvch(ma)
        ON DELETE SET NULL
        ON UPDATE CASCADE
);
