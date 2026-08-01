# Add MongoDB 8.3 to VPS Docker Compose — Design Spec

**Date:** 2026-08-01  
**Status:** Draft — awaiting review

---

## 1. Objective

Add a MongoDB 8.3 container to the VPS Docker Compose stack (`/home/thanglequoc/docker/database/vietnamese-provinces-docker-compose.yaml`) following the existing conventions (auth, localhost binding, named volume, restart policy).

## 2. Background

The VPS currently runs 4 database containers under the `database` compose project:

| Container | Image | Port |
|-----------|-------|------|
| `mysql_db` | `mysql:8.0` | `127.0.0.1:3306` |
| `mssql-server-2022` | `mcr.microsoft.com/mssql/server:2022-latest` | `127.0.0.1:1433` |
| `oracle-xe` | `gvenzl/oracle-xe:21-slim` | `127.0.0.1:1521` |
| `elasticsearch` | `docker.elastic.co/elasticsearch/elasticsearch:9.4.3` | `127.0.0.1:9200` |

MongoDB is needed as a target database for the MongoDB GIS dataset (per the `2026-07-30-mongodb-gis-dataset-design.md` spec), enabling local testing and validation of generated MongoDB GIS exports.

## 3. Design

### 3.1 Container Specification

| Aspect | Value | Rationale |
|--------|-------|-----------|
| **Image** | `mongo:8.0` | MongoDB 8.0 official image (mongodb/mongodb-community-server:8.3 not published on Docker Hub) |
| **Container name** | `mongodb` | Consistent with existing naming (`mysql_db`, `elasticsearch`, `oracle-xe`, `mssql-server-2022`) |
| **Port** | `127.0.0.1:27017:27017` | Default MongoDB port, localhost-only (matches all other DB services) |
| **Auth** | Root user + password via env vars | Matches MySQL/SQLServer/Oracle auth pattern |
| **Volume** | `mongodb_data:/data/db` | Named volume for persistence (matches existing volume pattern) |
| **Restart** | `always` | Matches all other services |

### 3.2 Environment Variables

```yaml
MONGO_INITDB_ROOT_USERNAME: root
MONGO_INITDB_ROOT_PASSWORD: Q35iSs8h5Y47VMcxZ5UC  # same as MySQL root password for simplicity
```

`MONGO_INITDB_ROOT_USERNAME` and `MONGO_INITDB_ROOT_PASSWORD` are only applied on first run (when the volume is empty). Changing them after initialization requires manual `db.createUser()` steps.

### 3.3 Connection String

```
mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/vn_provinces?authSource=admin
```

- `authSource=admin` is required because the root user is created in the `admin` database
- `vn_provinces` is the target database name (created on first use, no upfront DB creation needed)

### 3.4 Compose File Changes

**File:** `/home/thanglequoc/docker/database/vietnamese-provinces-docker-compose.yaml`

**New service** (added after `elasticsearch`):

```yaml
  mongodb:
    image: mongo:8.0
    container_name: mongodb
    restart: always
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: Q35iSs8h5Y47VMcxZ5UC
    ports:
      - "127.0.0.1:27017:27017"
    volumes:
      - mongodb_data:/data/db
```

**New volume** (added to `volumes:` section):

```yaml
  mongodb_data:
```

### 3.5 Local Test Compose

The local `dataset-generation-scripts/docker/docker-compose.yaml` is **not modified** in this scope. That change will be addressed when MongoDB integration tests are implemented.

## 4. Deployment Steps

1. SSH into the VPS: `ssh -i ~/.ssh/id_ed25519 thanglequoc@machine.thanglequoc.xyz`
2. Backup the existing compose file: `cp vietnamese-provinces-docker-compose.yaml vietnamese-provinces-docker-compose.yaml.bak`
3. Update the compose file with the new `mongodb` service and `mongodb_data` volume
4. Start the service (image pulls automatically): `docker compose -f /home/thanglequoc/docker/database/vietnamese-provinces-docker-compose.yaml up -d mongodb`
5. Verify: `docker exec mongodb mongosh -u root -p "Q35iSs8h5Y47VMcxZ5UC" --quiet --eval "db.version()"`
6. Verify connectivity from local (requires SSH tunnel to VPS): `mongosh "mongodb://root:Q35iSs8h5Y47VMcxZ5UC@localhost:27017/?authSource=admin&directConnection=true" --eval "db.version()"`

## 5. Post-Deployment

After MongoDB is running, the AGENTS.md should be updated to include a MongoDB query skill section (auto-trigger patterns, connection details, collection reference) — this will be done in a follow-up task after the MongoDB GIS dataset generation is complete and data is imported.

## 6. Rollback

If the deployment fails:
1. Stop and remove the container: `docker compose -f /home/thanglequoc/docker/database/vietnamese-provinces-docker-compose.yaml down mongodb`
2. Remove the volume (if no data to keep): `docker volume rm database_mongodb_data`
3. Restore the backup compose file: `cp vietnamese-provinces-docker-compose.yaml.bak vietnamese-provinces-docker-compose.yaml`

## 7. Out of Scope

- Local test docker-compose (`dataset-generation-scripts/docker/docker-compose.yaml`) — deferred to MongoDB integration test work
- AGENTS.md MongoDB query skill — deferred to post-data-import
- MongoDB replica set or sharding — single-node dev instance only
- Custom `mongod.conf` or startup flags — using defaults