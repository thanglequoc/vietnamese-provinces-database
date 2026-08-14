# Redis Dataset — Vietnamese Provinces Database

**Generated at: Fri, 14 Aug 2026 08:42:28 +0700**

Redis commands loading all Vietnamese provinces, wards, regions, and administrative units.

## Files

- `redis_vn_provinces_dataset.redis` — Redis HSET/SADD commands (1.11 MB)

## Data Structure

- `province:<code>` — province hash (name, nameEn, fullName, codeName, postalCodePrefix, administrativeUnitId)
- `ward:<code>` — ward hash (name, fullName, codeName, postalCode, administrativeUnitId, districtCode)
- `administrativeUnit:<id>` — unit type hash
- `region:<id>` — region hash
- `province:<code>:wards` — SET of ward codes
- `province:<code>:wards:vn` / `province:<code>:wards:en` — ward code → name hashes

## Sample Queries

```bash
redis-cli HGETALL province:01
redis-cli SMEMBERS province:01:wards
redis-cli HGET ward:00004 fullName
```
