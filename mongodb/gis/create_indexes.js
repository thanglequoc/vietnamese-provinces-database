// MongoDB GIS Index Creation Script
// Run with: mongosh vn_provinces create_indexes.js

// provinces-gis collection indexes
db.getCollection('provinces-gis').createIndex({ "Code": 1 }, { unique: true });
db.getCollection('provinces-gis').createIndex({ "GIS.Geometry": "2dsphere" });
db.getCollection('provinces-gis').createIndex({ "GIS.Center": "2dsphere" });
db.getCollection('provinces-gis').createIndex({ "SearchKeywords": 1 });

// wards-gis collection indexes
db.getCollection('wards-gis').createIndex({ "Code": 1 }, { unique: true });
db.getCollection('wards-gis').createIndex({ "ProvinceCode": 1 });
db.getCollection('wards-gis').createIndex({ "GIS.Geometry": "2dsphere" });
db.getCollection('wards-gis').createIndex({ "GIS.Center": "2dsphere" });
db.getCollection('wards-gis').createIndex({ "SearchKeywords": 1 });
