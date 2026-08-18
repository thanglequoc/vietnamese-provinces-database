# Design: Remove Meta (Dataset Version) from ES & MongoDB Documents

**Date:** 2026-08-16
**Branch:** `134_AddPostalCode`
**Status:** Approved

## Objective

Remove the `Meta` dataset-version object from Elasticsearch and MongoDB documents.
The version metadata will later be exposed as a dedicated index/collection rather
than bundled into every document, avoiding consumer-side parsing friction.

## Problem

Both Elasticsearch (`provinces`, `provinces-gis` NDJSON) and MongoDB
(`provinces-gis`, `wards-gis`) documents embed a `Meta` object with
`DatasetVersion`, `AdministrativeRevision`, and `GeneratedAt`. The
`GeneratedAt` timestamp changes on every generation run, so the entire data files
appear modified in git each time, and consumers must tolerate the nested object.

## Requirements (confirmed with user)

1. Remove `Meta` from Elasticsearch documents and MongoDB GIS documents.
2. Keep the dataset-version constants (`esDatasetVer`, `esAdminRev`,
   `mongoDatasetVer`, `mongoAdminRev`) for the future dedicated index.
3. Update generated READMEs, tests, and `AGENTS.md` schema references.

## Changes

### 1. Elasticsearch

- `dto/elasticsearch_dto.go`: remove `Meta` field from
  `ElasticsearchProvinceDocument` and `ElasticsearchWardDocument`; delete
  `ElasticsearchMeta` type.
- `elasticsearch_file_writer.go`:
  - `WriteToFile`: remove the Meta-attach loop and `generatedAt` (lines ~52-60).
  - `WriteElasticsearchGISDataToFile`: remove the per-doc `Meta` and `generatedAt`
    (lines ~100, ~115-120).
  - Remove `Meta` from both mapping JSON bodies (`provinces` and `provinces-gis`).
- Keep `esDatasetVer` / `esAdminRev` constants.

### 2. MongoDB (GIS)

- `dto/mongo_gis_dto.go`: remove `Meta` field from `MongoGISProvinceDocument`
  and `MongoGISWardDocument`; delete `MongoMeta` type.
- `helper/mongo_gis_mapper.go`: `ConvertToMongoGISProvinceDocuments` /
  `ConvertToMongoGISWardDocuments` drop the `(datasetVersion, adminRevision,
  generatedAt string)` parameters and no longer set `Meta`.
- `mongodb_gis_file_writer.go`: remove `generatedAt`; keep `mongoDatasetVer` /
  `mongoAdminRev` constants.
- Base `mongo_data_vn_unit.json` already has no `Meta` — unchanged.

### 3. READMEs, tests, docs

- `elasticsearch_file_writer.go` `writeESReadme`: remove the `Meta` bullet and
  the `Notes` line about `Meta`.
- `mongodb_file_writer.go` `writeMongoReadme`: remove the `Meta` bullet.
- Tests: `elasticsearch_file_writer_test.go` (remove `doc.Meta` assertions),
  `mongo_gis_mapper_test.go` (remove `doc.Meta` assertions, update mapper call
  signatures).
- `AGENTS.md`: remove `Meta` from the Elasticsearch and MongoDB schema
  references.

### 4. Regenerate + publish

Regenerate ES NDJSON and MongoDB GIS JSON, run the copy script, and commit the
updated dataset artifacts. This also eliminates the every-run `GeneratedAt`
churn in git.

## Edge Cases

- **Unused constants**: `esDatasetVer`/`esAdminRev`/`mongoDatasetVer`/
  `mongoAdminRev` become unreferenced package-level constants (legal in Go) and
  are intentionally kept for the future version index.
- **Consumer contract**: removing `Meta` is a breaking change for any consumer
  reading that field; it is intentional per the user's plan to expose version
  metadata separately.
