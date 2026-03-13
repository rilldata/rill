# Tech Draft: Declarative Template System for Add Data Modal

## Motivation

The Add Data modal's connector and source forms were defined by hardcoded TypeScript schemas (`web-common/src/features/templates/schemas/*.ts`). Each schema duplicated property metadata already present in Go driver specs, and adding a new connector required changes in both Go and TypeScript. The DuckDB SQL generation and env var extraction logic were also scattered across frontend utilities.

This migration replaces the hardcoded schemas with a **backend-driven, declarative template system**. Template definitions live as JSON files in the Go runtime, are served via API, and power both form rendering and YAML generation. Adding a new connector now requires only a single JSON file.

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                        Frontend                               │
│                                                               │
│  AddDataModal ─── createConnectorSchemas() ──► ListTemplates  │
│       │                                           RPC         │
│       ▼                                                       │
│  Connector Grid  (icons + categories from json_schema)        │
│       │                                                       │
│       ▼                                                       │
│  AddDataForm ─── generateTemplate() ──────► GenerateFile RPC  │
│       │           (debounced, preview=true)       │            │
│       ▼                                          │            │
│  YAML Preview  ◄─────────────────────────────────┘            │
│       │                                                       │
│  Submit ──► GenerateFile(preview=false) ──► writes files      │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                     Backend (Go runtime)                       │
│                                                               │
│  runtime/templates/                                           │
│    registry.go ── //go:embed definitions/ ── loads 30 JSONs   │
│    render.go ──── property pre-processing + Go text/template  │
│    duckdb.go ──── read_csv, read_parquet, read_json SQL       │
│    clickhouse.go ── s3(), gcs(), mysql(), postgresql() SQL    │
│    env.go ──────── secret extraction + conflict resolution    │
│    headers.go ──── HTTP header secrets                        │
│                                                               │
│  runtime/server/templates.go ── ListTemplates, GenerateFile   │
└──────────────────────────────────────────────────────────────┘
```

## What Was Built

### 1. `runtime/templates/` Go Package (7 files)

Core types, registry, and rendering engine.

**`template.go`** — Types:
- `Template`: name, display_name, description, docs_url, driver, olap, tags, json_schema, files
- `File`: name ("connector" or "model"), path_template, code_template
- `ProcessedProp`: key, value (with secret refs), quoted flag

**`registry.go`** — Loads all embedded JSON definitions via `//go:embed`. Methods:
- `List()` — all templates, sorted by name
- `Get(name)` — lookup by exact name
- `ListByTags(tags)` — filter templates matching ALL tags
- `LookupByDriver(driver, resourceType)` — backward-compat mapping for legacy `GenerateTemplate` RPC

**`render.go`** — Rendering pipeline:
1. Pre-process properties: filter empties, extract `x-secret` fields to env vars, skip `x-ui-only` fields
2. Split properties by `x-step` (connector vs source vs explorer)
3. Compute derived fields: DuckDB SQL from path, ClickHouse SQL from driver-specific properties
4. Render each file's path and code templates using Go `text/template` with `[[ ]]` delimiters

Two processing paths:
- **Schema-based** (new): reads `x-secret`, `x-env-var`, `x-ui-only`, `x-step` from `json_schema`
- **Driver-spec** (legacy): reads from `drivers.PropertySpec` for templates without `json_schema`

**`duckdb.go`** — `BuildDuckDBQuery(path, defaultToJSON)`:
- Infers format from file extension (`.csv` → `read_csv`, `.parquet` → `read_parquet`, `.json` → `read_json`)
- Checks basename suffix to avoid false positives (`parquet-archive/readme.txt` is not parquet)

**`clickhouse.go`** — ClickHouse table function SQL builders:
- `BuildClickHouseObjectStoreQuery()` — `s3()` / `gcs()` with optional credentials
- `BuildClickHouseAzureQuery()` — `azureBlobStorage()` with parsed endpoint
- `BuildClickHouseDatabaseQuery()` — `mysql()` / `postgresql()` with connection params
- `BuildClickHouseURLQuery()` — `url()` with format inference
- `BuildClickHouseFileQuery()` — `file()` with format inference
- `BuildClickHouseSQLiteQuery()` — `sqlite()` with db path and table

**`env.go`** — Environment variable handling:
- `ResolveEnvVarName()` / `ResolveEnvVarNameForKey()` — determine env var name from driver + property; append `_1`, `_2` for conflicts
- `ReadEnvKeys()` — parse existing `.env` to detect conflicts

**`headers.go`** — HTTP header secret extraction:
- `IsSensitiveHeaderKey()` — detects Authorization, X-API-Key, etc.
- `SplitAuthSchemePrefix()` — extracts Bearer/Basic/Token prefix
- `ResolveHeaderEnvVarName()` — generates `connector.{name}.{segment}` env var name

**`funcmap.go`** — Template functions available in `[[ ]]` templates:
- `renderProps` — renders `[]ProcessedProp` as YAML key-value lines with proper quoting
- `indent` — prepends N spaces per line (for SQL in YAML)
- `quote` — wraps string in double quotes

### 2. Template Definitions (30 JSON files)

Located in `runtime/templates/definitions/`:

```
definitions/
├── olap/                      # OLAP connector templates (6)
│   ├── duckdb.json
│   ├── clickhouse.json
│   ├── motherduck.json
│   ├── druid.json
│   ├── pinot.json
│   └── starrocks.json
├── duckdb-models/             # Source → DuckDB model templates (15)
│   ├── s3-duckdb.json
│   ├── gcs-duckdb.json
│   ├── azure-duckdb.json
│   ├── https-duckdb.json
│   ├── local-file-duckdb.json
│   ├── sqlite-duckdb.json
│   ├── postgres-duckdb.json
│   ├── mysql-duckdb.json
│   ├── bigquery-duckdb.json
│   ├── snowflake-duckdb.json
│   ├── athena-duckdb.json
│   ├── redshift-duckdb.json
│   ├── salesforce-duckdb.json
│   ├── clickhouse-duckdb.json
│   ├── duckdb-duckdb.json
│   └── iceberg-duckdb.json    # NEW (motivating use case)
└── clickhouse-models/         # Source → ClickHouse model templates (8)
    ├── s3-clickhouse.json
    ├── gcs-clickhouse.json
    ├── azure-clickhouse.json
    ├── https-clickhouse.json
    ├── local-file-clickhouse.json
    ├── postgres-clickhouse.json
    ├── mysql-clickhouse.json
    └── sqlite-clickhouse.json
```

Each template JSON contains:
- Metadata: `name`, `display_name`, `description`, `docs_url`, `driver`, `olap`, `tags`
- `json_schema`: JSON Schema (draft-07) with custom `x-*` extensions for UI and backend behavior
- `files`: array of output file definitions (path + code templates)

**Custom `x-*` extensions on `json_schema`:**

| Extension | Scope | Description |
|-----------|-------|-------------|
| `x-category` | schema | UI category: `olap`, `objectStore`, `fileStore`, `sqlStore`, `warehouse`, `source_only` |
| `x-icon` | schema | Full-size icon component name (for connector grid) |
| `x-small-icon` | schema | Small icon component name (for nav, cards, headers) |
| `x-form-width` | schema | Form width: `wide` or default |
| `x-form-height` | schema | Form height: `tall` or default |
| `x-step` | property | Form step routing: `connector`, `source`, `explorer` |
| `x-secret` | property | Extract value to `.env` as env var |
| `x-env-var` | property | Explicit env var name (else defaults to `DRIVER_KEY`) |
| `x-ui-only` | property | Skip in backend rendering (e.g. radio button selectors) |
| `x-placeholder` | property | Input placeholder text |
| `x-display` | property | Display type: `radio`, `select`, `tabs`, `text` |
| `x-visible-if` | property | Conditional visibility: `{ field: "auth_method", value: "access_keys" }` |
| `x-grouped-fields` | property | Map enum value → array of visible field names |
| `x-tab-group` | property | Tab group name for tabbed field display |
| `x-enum-labels` | property | Display labels for enum values |
| `x-enum-descriptions` | property | Descriptions for each enum value |

### 3. Proto Definitions + Server Handlers

**New RPCs** (in `proto/rill/runtime/v1/api.proto`):

```protobuf
rpc ListTemplates(ListTemplatesRequest) returns (ListTemplatesResponse);
rpc GenerateFile(GenerateFileRequest) returns (GenerateFileResponse);
```

- `ListTemplates` — returns templates filtered by tags; powers the connector grid
- `GenerateFile` — renders a named template with properties; supports `preview` mode (render without writing) and `output` filter ("connector" or "model")

**New messages**: `Template`, `TemplateFile`, `GeneratedFile`, `ListTemplatesRequest/Response`, `GenerateFileRequest/Response`

**Server handlers** (`runtime/server/templates.go`):
- `ListTemplates` — delegates to registry, converts to proto
- `GenerateFile` — looks up template, reads `.env` for conflict resolution, calls `templates.Render()`, optionally writes files and merges env vars

**Legacy**: `GenerateTemplate` RPC retained for backward compatibility; delegates to `GenerateFile` internally.

### 4. Frontend Changes

**`connector-schemas.ts`** — Schema registry, completely rewritten:
- `createConnectorSchemas(instanceId)` — TanStack Query that calls `ListTemplates` + `GetInstance` RPCs
- `buildSchemaRegistry(templates, olap)` — transforms API templates into local cache, OLAP-aware
- `normalizeOlapForTemplate()` — maps instance OLAP to template suffix (clickhouse → "clickhouse", else → "duckdb")
- Icon auto-discovery via `import.meta.glob("../../../components/icons/connectors/*.svelte")` — no manual imports
- Exported `ICONS` and `connectorIconMapping` maps rebuilt from schema `x-icon` / `x-small-icon`

**`generate-template.ts`** — RPC wrapper for YAML generation:
- `resolveTemplateName(driver, olap)` — OLAP engines use standalone name; sources use `{driver}-{olap}`
- `generateTemplate()` — calls `GenerateFile` with `preview: true`; caches OLAP per instance
- `mergeEnvVars()` — merges env vars into `.env` file

**`AddDataModal.svelte`** — passes `instanceId` to `createConnectorSchemas()`

**`AddDataForm.svelte`** — debounced YAML preview via `generateTemplate()` on every form keystroke

**Removed**: All hardcoded TypeScript schema files (`web-common/src/features/templates/schemas/s3.ts`, `gcs.ts`, etc.) deleted; replaced by API-driven JSON schemas.

### 5. OLAP-Aware Template Selection

When a project uses ClickHouse as its OLAP engine:
- `createConnectorSchemas()` queries `GetInstance` with `sensitive: true` to get `olapConnector`
- `normalizeOlapForTemplate()` maps it to "clickhouse"
- `buildSchemaRegistry()` filters templates by `t.olap === olap`
- Sources without a ClickHouse template (athena, bigquery, redshift, salesforce, snowflake, iceberg, duckdb) are naturally hidden

ClickHouse source templates use `"x-category": "source_only"` — single-page forms without the multi-step connector flow, since ClickHouse table functions embed credentials directly in SQL.

### 6. Icon System

Icons are resolved by string name from template JSON → Svelte component:
- `x-icon` — full-size icon for connector grid
- `x-small-icon` — small icon for nav/cards/headers; falls back to `x-icon`

New small icon components created: `SQLiteIcon.svelte`, `LocalFileIcon.svelte`, `HTTPSIcon.svelte`

Existing icons updated with `size` prop: `GoogleCloudStorageIcon.svelte`, `MicrosoftAzureBlobStorageIcon.svelte`

## Data Flow: YAML Preview

```
User types in form field
    ↓ (debounced 150ms)
AddDataForm.computeYamlPreview()
    ↓
generateTemplate(instanceId, { driver, resourceType, properties })
    ↓
resolveTemplateName() → e.g. "s3-duckdb"
    ↓
runtimeServiceGenerateFile(instanceId, { templateName, output, properties, preview: true })
    ↓ (HTTP POST to backend)
Server.GenerateFile()
    ↓
templates.Render()
  1. processPropertiesFromSchema() → extract secrets, skip empties/ui-only
  2. splitPropsByStep() → route to connector vs model file
  3. applyDuckDBDerivedFields() or applyClickHouseDerivedFields() → compute SQL
  4. renderString() → execute Go text/template with [[ ]] delimiters
    ↓
Response: { files: [{ path, blob }], envVars }
    ↓
Display YAML in preview pane
```

## Template Rendering Details

Templates use Go `text/template` with `[[ ]]` delimiters to avoid collision with Rill's `{{ .env.VAR }}` runtime syntax.

Example template code (from `s3-duckdb.json`):
```
# Connector YAML for S3
# Ref: [[ .docs_url ]]
type: connector
driver: s3
[[ renderProps .config_props ]]
```

The `renderProps` function renders `[]ProcessedProp` as YAML:
```yaml
aws_access_key_id: "{{ .env.AWS_ACCESS_KEY_ID }}"
aws_secret_access_key: "{{ .env.AWS_SECRET_ACCESS_KEY }}"
endpoint: "https://custom-endpoint.com"
```

Secrets are replaced with `{{ .env.VAR }}` references; the actual values are returned separately in `envVars` for `.env` file merging.

## Bug Fixes Included

| Bug | Fix |
|-----|-----|
| `containsExt` false positive: `parquet-archive/readme.txt` matched `.parquet` | `matchesExt()` now checks basename suffix only |
| `headerKeyToEnvSegment` regex compiled on every call | Compiled once at package level |
| Template render errors silently returned empty blob | Errors now propagated to caller |
| OLAP connector not detected for managed ClickHouse | `GetInstance` called with `sensitive: true` (required to get `olapConnector` field) |

## Test Coverage

**Go tests** (4 test files, 32+ cases in `runtime/templates/` and `runtime/server/`):
- Registry: loading, duplicate detection, lookup by driver, tag filtering, sorted output, all 30 definitions valid
- Render: S3 connector, S3-DuckDB model, Snowflake warehouse model, Redshift no-dev, Iceberg-DuckDB, env var conflicts, empty filtering, output filtering, local file, SQLite
- Env: explicit name, fallback, single conflict, double conflict
- DuckDB: query building for all formats
- ClickHouse: object store, database, URL, file, SQLite queries
- Headers: sensitive detection, auth scheme splitting, env segment naming

**Frontend tests** (`generate-template.spec.ts`):
- Template name resolution for DuckDB and ClickHouse OLAP
- OLAP engine standalone template names

## Backward Compatibility

- `GenerateTemplate` RPC retained; delegates to `GenerateFile` internally
- `DriverSpec` fallback: templates without `json_schema` use `drivers.PropertySpec` for property metadata
- Frontend auto-generated clients updated via Orval; old `runtimeServiceGenerateTemplate` still available

## Key Files

| File | Description |
|------|-------------|
| `runtime/templates/template.go` | Core types: Template, File, ProcessedProp |
| `runtime/templates/registry.go` | Registry with //go:embed loading |
| `runtime/templates/render.go` | Rendering pipeline with property pre-processing |
| `runtime/templates/duckdb.go` | DuckDB SQL generation (read_csv, read_parquet, etc.) |
| `runtime/templates/clickhouse.go` | ClickHouse table function SQL builders |
| `runtime/templates/env.go` | Env var naming and conflict resolution |
| `runtime/templates/headers.go` | HTTP header secret extraction |
| `runtime/templates/funcmap.go` | Template functions (renderProps, indent, quote) |
| `runtime/templates/definitions/**/*.json` | 30 template definitions |
| `runtime/server/templates.go` | ListTemplates + GenerateFile RPC handlers |
| `runtime/server/generate_template.go` | Legacy GenerateTemplate handler |
| `proto/rill/runtime/v1/api.proto` | Proto definitions for template RPCs |
| `web-common/src/features/sources/modal/connector-schemas.ts` | Frontend schema registry (API-driven) |
| `web-common/src/features/sources/modal/generate-template.ts` | Frontend RPC wrapper |
| `web-common/src/features/sources/modal/AddDataModal.svelte` | Modal entry point |
| `web-common/src/features/sources/modal/AddDataForm.svelte` | Form rendering + YAML preview |
| `web-common/src/features/sources/modal/AddDataFormManager.ts` | Form orchestration |

## How to Add a New Connector

1. Create a JSON template file in the appropriate `definitions/` subdirectory
2. Define `json_schema` with field properties, `x-step` routing, `x-secret` for credentials
3. Set `x-icon` / `x-small-icon` to existing or new icon component names
4. Add any new icon `.svelte` files to `web-common/src/components/icons/connectors/`
5. Add the driver name to the `SOURCES` constant in `web-common/src/features/sources/modal/constants.ts`
6. Run `go test ./runtime/templates/...` and `npm run test -w web-common`
