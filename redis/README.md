# Redis Dataset — Vietnamese Provinces Database

**Generated at: Thu, 27 Aug 2026 08:52:14 +0700**

Redis commands (HSET/SADD) loading all Vietnamese provinces, wards, regions, and administrative units.

## Files

- `redis_vn_provinces_dataset.redis` — Redis HSET/SADD commands (1.11 MB)

## Overview

The dataset stores every administrative unit as Redis hashes and sets:

| Key pattern | Type | Count |
|-------------|------|-------|
| `province:<code>` | hash | 34 |
| `ward:<code>` | hash | 3,321 |
| `administrativeUnit:<id>` | hash | 8 |
| `region:<id>` | hash | 8 |
| `province:<code>:wards` | set | 34 |
| `province:<code>:wards:vn` / `:en` | hash | 34 each |

## Data Structure

`province:<code>` fields: `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `postalCodePrefix`, `administrativeUnitId`.

`ward:<code>` fields: `code`, `name`, `nameEn`, `fullName`, `fullNameEn`, `codeName`, `postalCode`, `administrativeUnitId`, `districtCode`.

`province:<code>:wards:vn` / `:en` map ward codes to Vietnamese/English full names.

## Sample Document

```bash
HSET province:01 code "01" name "Hà Nội" nameEn "Ha Noi" fullName "Thành phố Hà Nội" fullNameEn "Ha Noi City" codeName "ha_noi" postalCodePrefix "10, 11, 12, 13, 14" administrativeUnitId 1

SADD province:01:wards "00004"
HSET province:01:wards:vn "00004" "Phường Ba Đình"
```

## Quick Start

```bash
redis-cli --pipe < redis_vn_provinces_dataset.redis
```

## Sample Queries

```bash
redis-cli HGETALL province:01
redis-cli SMEMBERS province:01:wards
redis-cli HGET ward:00004 fullName
redis-cli HGET province:01:wards:vn 00004
```
