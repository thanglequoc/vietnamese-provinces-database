# Vietnamese Province Dataset for Redis

This folder contains a Redis-ready export of the Vietnamese administrative units dataset. It stores the current **34 provinces** and **3,321 wards** as Redis hashes, plus precomputed province-to-ward lookup keys for fast reads.

## Files

- `redis_vn_provinces_dataset.redis`: checked-in Redis command script that can be loaded directly with `redis-cli`

If you regenerate the dataset from `dataset-generation-scripts`, the generator may produce a timestamped filename such as `redis_vn_provinces_dataset_<timestamp>.redis`. The key structure is the same.

## Load the dataset into Redis

### Prerequisites

- A running Redis server
- `redis-cli` installed locally
- Access to the dataset file in this directory

### Import into a local Redis instance

```bash
cd redis
redis-cli < redis_vn_provinces_dataset.redis
```

### Import into a remote Redis instance

```bash
cd redis
redis-cli -h 127.0.0.1 -p 6379 -a '<password>' -n 0 < redis_vn_provinces_dataset.redis
```

The checked-in file is a plain Redis command script (`HSET`, `SADD`, ...), so standard stdin redirection is the correct import method for this dataset format.

### Verify the import

```bash
redis-cli DBSIZE
redis-cli EXISTS province:01
redis-cli EXISTS ward:00004
redis-cli SCARD province:01:wards
```

Expected result:

- `DBSIZE` should return a non-zero number
- `EXISTS province:01` should return `1`
- `EXISTS ward:00004` should return `1`
- `SCARD province:01:wards` should return the number of wards stored for province `01`

## Dataset architecture

The Redis export is a key-value representation of the dataset, not a relational schema. Each entity is stored as a hash, and province-to-ward traversal is precomputed with sets and lookup hashes.

### Key families

| Key pattern | Redis type | Purpose |
|-------------|------------|---------|
| `administrativeUnit:{id}` | Hash | Administrative unit lookup such as municipality, province, ward, commune |
| `region:{id}` | Hash | Vietnamese geographical region lookup |
| `province:{code}` | Hash | Province or municipality record |
| `ward:{code}` | Hash | Ward, commune, or special administrative region record |
| `province:{province_code}:wards` | Set | All ward codes that belong to a province |
| `province:{province_code}:wards:vn` | Hash | Ward code -> Vietnamese full name within a province |
| `province:{province_code}:wards:en` | Hash | Ward code -> English full name within a province |

### Stored fields

#### `administrativeUnit:{id}`

| Field | Meaning |
|-------|---------|
| `id` | Administrative unit id |
| `fullName` | Full Vietnamese name |
| `fullNameEn` | Full English name |
| `shortName` | Short Vietnamese name |
| `shortNameEn` | Short English name |
| `codeName` | Vietnamese code name |

#### `region:{id}`

| Field | Meaning |
|-------|---------|
| `name` | Vietnamese region name |
| `nameEn` | English region name |
| `codeName` | Vietnamese code name |

#### `province:{code}`

| Field | Meaning |
|-------|---------|
| `code` | Official province code |
| `name` | Vietnamese short name |
| `nameEn` | English short name |
| `fullName` | Vietnamese full name |
| `fullNameEn` | English full name |
| `codeName` | Vietnamese code name |
| `administrativeUnitId` | Link to `administrativeUnit:{id}` |

#### `ward:{code}`

| Field | Meaning |
|-------|---------|
| `code` | Official ward code |
| `name` | Vietnamese short name |
| `nameEn` | English short name |
| `fullName` | Vietnamese full name |
| `fullNameEn` | English full name |
| `codeName` | Vietnamese code name |
| `administrativeUnitId` | Link to `administrativeUnit:{id}` |
| `districtCode` | Currently stores the parent province code in this export |

## Relationship diagram

```mermaid
flowchart TD
    AU[administrativeUnit:{id}\nHash]
    RG[region:{id}\nHash]
    P[province:{code}\nHash]
    W[ward:{code}\nHash]
    PS[province:{code}:wards\nSet of ward codes]
    PVN[province:{code}:wards:vn\nHash code -> Vietnamese full name]
    PEN[province:{code}:wards:en\nHash code -> English full name]

    P -->|administrativeUnitId| AU
    W -->|administrativeUnitId| AU
    P --> PS
    PS --> W
    P --> PVN
    P --> PEN
    PVN --> W
    PEN --> W
```

`region:{id}` keys are included as standalone reference data in the current Redis export.

## Basic Redis operations

### Get a province

```bash
redis-cli HGETALL province:01
```

Example output fields include:

- `code`
- `name`
- `nameEn`
- `fullName`
- `fullNameEn`
- `codeName`
- `administrativeUnitId`

### Get a ward

```bash
redis-cli HGETALL ward:00004
```

### Get all ward codes under a province

```bash
redis-cli SMEMBERS province:01:wards
```

### Get all ward names under a province

Vietnamese names:

```bash
redis-cli HGETALL province:01:wards:vn
```

English names:

```bash
redis-cli HGETALL province:01:wards:en
```

### Get full ward records under a province

Redis does not support SQL-style joins, so the usual pattern is:

1. Read ward codes from the province set.
2. Fetch each `ward:{code}` hash.

Example shell workflow:

```bash
redis-cli --raw SMEMBERS province:01:wards | while read -r ward_code; do
  echo "ward:$ward_code"
  redis-cli HGETALL "ward:$ward_code"
done
```

### Inspect administrative unit metadata

```bash
redis-cli HGETALL administrativeUnit:1
```

## Data model notes

- The Redis export is optimized for key-based reads, not relational querying.
- Province-to-ward membership is precomputed, so listing wards under a province is fast.
- There is no separate reverse index from ward name to ward code in the current export.
- `region:{id}` keys are included as reference data, but province hashes do not currently store a direct region field.
- Some field names reflect the current generator implementation rather than an ideal schema. In particular, `districtCode` inside `ward:{code}` currently contains the parent province code.
