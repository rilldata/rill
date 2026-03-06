---
title: "feat: Backend GenerateTemplate RPC for declarative YAML generation"
type: feat
date: 2026-02-16
---

# Backend GenerateTemplate RPC for Declarative YAML Generation

## Overview

Replace the frontend's imperative `compileSourceYAML()` and `compileConnectorYAML()` string-building functions with a backend `GenerateTemplate` RPC endpoint. The backend becomes the single source of truth for YAML file format, using driver `PropertySpec` metadata to render YAML from structured form data. This includes absorbing the `maybeRewriteToDuckDb()` logic — the backend uses `Spec.ImplementsObjectStore` / `ImplementsFileStore` to detect when a driver should be rewritten to a DuckDB model, builds the appropriate SQL query, and returns the rewritten resource type. Scoped to connectors and sources/models — skeleton resource YAML (explores, dashboards, etc.) remains as frontend constants.

## Problem Statement

The frontend has two imperative YAML builder functions that share overlapping logic but diverge in subtle ways:

| Shared Concern | `compileSourceYAML` | `compileConnectorYAML` |
|---|---|---|
| YAML header + doc link | Yes | Yes |
| Secret -> `{{ .env.VAR }}` | Yes | Yes |
| String quoting | Yes | Yes |
| Empty value filtering | Yes | Yes |
| SQL multi-line formatting | Yes | No |
| Headers map formatting | No | Yes |
| Property ordering | Implicit | Explicit (`orderedProperties`) |
| Dev section | Yes (except Redshift) | No |

**Pain points:**
1. **Duplication** — both functions reimplement secret handling, quoting, and value filtering
2. **Maintenance burden** — adding a new connector requires touching complex conditional logic with special cases (ClickHouse `managed: false`, DuckDB SQL rewriting, HTTP auth scheme splitting)
3. **Frontend owns format** — the frontend knows too much about what valid YAML looks like; this knowledge should live server-side
4. **Frontend owns driver rewriting** — `maybeRewriteToDuckDb()` makes resource-type decisions (connector → model) and driver decisions (s3 → duckdb) that belong server-side. The backend already has the signals it needs (`Spec.ImplementsObjectStore`, `Spec.ImplementsFileStore`)

## Proposed Solution

**Before:**
```
Form Data -> compileSourceYAML() / compileConnectorYAML() -> YAML string -> PutFile RPC -> Backend
                     (imperative string builder)
```

**After:**
```
Form Data -> GenerateTemplate RPC -> { blob, env_vars } -> PutFile RPC (blob) + merge .env (env_vars)
                  (backend)
```

## Technical Approach

### Proto Definition

```protobuf
rpc GenerateTemplate(GenerateTemplateRequest) returns (GenerateTemplateResponse) {
  option (google.api.http) = {
    post: "/v1/instances/{instance_id}/generate/template",
    body: "*"
  };
}

message GenerateTemplateRequest {
  string instance_id = 1 [(validate.rules).string = {pattern: "^[_\\-a-zA-Z0-9]+$"}];
  string resource_type = 2 [(validate.rules).string = {in: ["connector", "model"]}];
  string driver = 3 [(validate.rules).string = {pattern: "^[a-z][a-z0-9_]*$"}];
  google.protobuf.Struct properties = 4;
  string connector_name = 5;
}

message GenerateTemplateResponse {
  string blob = 1;
  map<string, string> env_vars = 2;
  // Actual resource type used — may differ from request when driver is rewritten
  // (e.g., s3 model request → duckdb model with SQL wrapping the s3 path)
  string resource_type = 3;
  // Actual driver used — may differ from request when rewritten to duckdb
  string driver = 4;
}
```

`google.protobuf.Struct` matches existing codebase patterns (used in 11 places across `api.proto`). `resource_type` scoped to `"connector"` and `"model"` only. The response echoes the actual `resource_type` and `driver` used after any rewrites, so the frontend knows the correct file path and directory.

### Backend Handler

New file: `runtime/server/generate_template.go`

```go
func (s *Server) GenerateTemplate(ctx context.Context, req *runtimev1.GenerateTemplateRequest) (*runtimev1.GenerateTemplateResponse, error) {
    // 1. Permission check
    if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditRepo) {
        return nil, ErrForbidden
    }

    // 2. Validate driver exists
    driver, ok := drivers.Connectors[req.Driver]
    if !ok {
        return nil, status.Errorf(codes.InvalidArgument, "unknown driver: %s", req.Driver)
    }
    spec := driver.Spec()

    // 3. Validate properties against original driver spec (reject unknown keys)
    props := req.Properties.AsMap()
    if err := validateProperties(spec, req.ResourceType, props); err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "%s", err)
    }

    // 4. DuckDB rewrite for object store / file store / sqlite drivers
    //    Rewrite happens AFTER validation against the original driver spec,
    //    since the request properties match the original driver (e.g., s3's "path").
    actualDriver, actualResourceType := req.Driver, req.ResourceType
    if req.ResourceType == "model" {
        actualDriver, props = maybeRewriteToDuckDB(spec, req.Driver, props, req.ConnectorName)
    }

    // 5. Read existing .env for env var conflict resolution
    repo, release, err := s.runtime.Repo(ctx, req.InstanceId)
    if err != nil {
        return nil, err
    }
    defer release()
    existingEnv := readEnvKeys(repo, ctx)

    // 6. Render YAML + extract secrets (using rewritten driver/props)
    blob, envVars := renderYAML(spec, actualDriver, actualResourceType, props, existingEnv)

    return &runtimev1.GenerateTemplateResponse{
        Blob:         blob,
        EnvVars:      envVars,
        ResourceType: actualResourceType,
        Driver:       actualDriver,
    }, nil
}
```

The handler reads `.env` via `repo.Get()` rather than accepting it as a request field. This eliminates sending ALL project secrets over the wire and removes the TOCTOU race between frontend read and backend use.

### YAML Rendering Engine

The backend renders YAML using the driver's `PropertySpec` metadata:

1. **Property ordering**: Use `ConfigProperties` / `SourceProperties` order from the driver `Spec`
2. **Secret detection**: Use `PropertySpec.Secret` (backend is source of truth)
3. **Env var naming**: Use new `PropertySpec.EnvVarName` field (falls back to `DRIVER_KEY` format)
4. **String quoting**: Use `PropertySpec.Type == StringPropertyType`
5. **Value filtering**: Skip empty/nil values, skip properties not in request
6. **Property validation**: Reject unknown property keys not in driver's `PropertySpec`
7. **Header comment**: Generate `# Connector YAML\n# Reference documentation: <DocsURL>`
8. **Dev section**: Auto-generate for warehouse drivers (except Redshift) with `limit 10000`

Build the `yaml.Node` tree directly rather than the marshal-unmarshal-encode triple-pass used in `generate_metrics_view.go:549-582`:

```go
func buildConnectorYAML(spec drivers.Spec, driverName string, props map[string]any, envVarMap map[string]string) *yaml.Node {
    doc := &yaml.Node{Kind: yaml.DocumentNode}
    mapping := &yaml.Node{Kind: yaml.MappingNode}
    mapping.HeadComment = fmt.Sprintf("Connector YAML\nReference documentation: %s", spec.DocsURL)

    addScalarPair(mapping, "type", "connector")
    addScalarPair(mapping, "driver", driverName)

    for _, propSpec := range spec.ConfigProperties {
        val, ok := props[propSpec.Key]
        if !ok || isEmpty(val) {
            continue
        }
        if propSpec.Secret {
            addScalarPair(mapping, propSpec.Key, fmt.Sprintf("{{ .env.%s }}", envVarMap[propSpec.Key]))
        } else if propSpec.Type == drivers.StringPropertyType {
            addQuotedPair(mapping, propSpec.Key, fmt.Sprintf("%v", val))
        } else {
            addScalarPair(mapping, propSpec.Key, fmt.Sprintf("%v", val))
        }
    }

    doc.Content = append(doc.Content, mapping)
    return doc
}

func addScalarPair(m *yaml.Node, key, value string) {
    m.Content = append(m.Content,
        &yaml.Node{Kind: yaml.ScalarNode, Value: key},
        &yaml.Node{Kind: yaml.ScalarNode, Value: value},
    )
}

func addQuotedPair(m *yaml.Node, key, value string) {
    m.Content = append(m.Content,
        &yaml.Node{Kind: yaml.ScalarNode, Value: key},
        &yaml.Node{Kind: yaml.ScalarNode, Value: value, Style: yaml.DoubleQuotedStyle},
    )
}

func validateProperties(spec drivers.Spec, resourceType string, properties map[string]any) error {
    allowed := make(map[string]bool)
    props := spec.ConfigProperties
    if resourceType == "model" {
        props = spec.SourceProperties
    }
    for _, p := range props {
        allowed[p.Key] = true
    }
    for key := range properties {
        if !allowed[key] {
            return fmt.Errorf("unknown property %q for driver", key)
        }
    }
    return nil
}
```

### Key Design Decisions

#### 1. DuckDB Rewrite Moves to Backend

`maybeRewriteToDuckDb()` transforms S3/GCS/Azure/HTTPS/SQLite/local_file into DuckDB model files. This logic moves to the backend because:

- The backend already has the signals: `Spec.ImplementsObjectStore` (s3, gcs, azure), `Spec.ImplementsFileStore` (https, local_file), and driver name (`sqlite`)
- The rewrite is a resource-type decision that the backend should own
- The `buildDuckDbQuery()` file-extension → read function mapping is ~20 lines of Go
- The `create_secrets_from_connectors` wiring is ~5 lines

The frontend sends the **original** driver and properties (e.g., `driver: "s3"`, `properties: {path: "s3://bucket/file.parquet"}`). The backend detects the rewrite case, builds the DuckDB SQL, and returns the actual driver/resource_type used in the response so the frontend knows the correct file path.

```go
func maybeRewriteToDuckDB(spec drivers.Spec, driverName string, props map[string]any, connectorName string) (string, map[string]any) {
    if !spec.ImplementsObjectStore && !spec.ImplementsFileStore && driverName != "sqlite" {
        return driverName, props
    }

    rewritten := make(map[string]any, len(props))
    for k, v := range props {
        rewritten[k] = v
    }

    switch {
    case spec.ImplementsObjectStore: // s3, gcs, azure
        if connectorName != "" {
            rewritten["create_secrets_from_connectors"] = connectorName
        }
        rewritten["sql"] = buildDuckDBQuery(strVal(props["path"]), false)
        delete(rewritten, "path")

    case spec.ImplementsFileStore && driverName == "https":
        if connectorName != "" {
            rewritten["create_secrets_from_connectors"] = connectorName
        }
        rewritten["sql"] = buildDuckDBQuery(strVal(props["path"]), true)
        delete(rewritten, "path")

    case spec.ImplementsFileStore: // local_file
        rewritten["sql"] = buildDuckDBQuery(strVal(props["path"]), false)
        delete(rewritten, "path")

    case driverName == "sqlite":
        rewritten["sql"] = fmt.Sprintf("SELECT * FROM sqlite_scan('%s', '%s');",
            strVal(props["db"]), strVal(props["table"]))
        delete(rewritten, "db")
        delete(rewritten, "table")
    }

    return "duckdb", rewritten
}

func buildDuckDBQuery(path string, defaultToJSON bool) string {
    ext := strings.ToLower(filepath.Ext(path))
    switch {
    case containsExt(ext, ".csv", ".tsv", ".txt"):
        return fmt.Sprintf("select * from read_csv('%s', auto_detect=true, ignore_errors=1, header=true)", path)
    case containsExt(ext, ".parquet"):
        return fmt.Sprintf("select * from read_parquet('%s')", path)
    case containsExt(ext, ".json", ".ndjson"):
        return fmt.Sprintf("select * from read_json('%s', auto_detect=true, format='auto')", path)
    default:
        if defaultToJSON {
            return fmt.Sprintf("select * from read_json('%s', auto_detect=true, format='auto')", path)
        }
        return fmt.Sprintf("select * from '%s'", path)
    }
}
```

Note: `containsExt` checks if the full extension (e.g., `.v1.parquet.gz`) contains the target part, matching the current frontend behavior of `extensionContainsParts()`.

#### 2. Backend Is Source of Truth for Secrets

Backend uses `PropertySpec.Secret` to identify secrets. No `secret_keys` in the request. Automated Go test asserts Secret flags match expected values for all drivers before shipping.

#### 3. Env Var Naming Uses New `EnvVarName` Field

Add `EnvVarName string` to `PropertySpec`. When set, use it (e.g., S3's `aws_access_key_id` -> `AWS_ACCESS_KEY_ID`). When empty, fall back to `DRIVER_KEY` format. Must match current frontend naming exactly or existing `.env` files break on upgrade.

#### 4. Backend Reads `.env` Directly

Backend reads `.env` via `repo.Get()` and parses key names into a `map[string]bool` for O(1) conflict detection. Response `env_vars` is a delta. Frontend merges with existing `.env` using `replaceOrAddEnvVariable()`.

#### 5. Frontend Strips x-ui-only Fields

Fields like `deployment_type`, `connection_mode`, `auth_method` are stripped before calling `GenerateTemplate`. Backend validates and rejects unknown keys. Postgres DSN-vs-parameters tabs are handled by `filterSchemaValuesForSubmit()`.

#### 6. `GenerateTemplate` Does Not Write Files

Returns blob + env_vars. Frontend calls `PutFile` separately. This enables rollback (restore original `.env` on reconciliation failure) and the "Save Anyway" bypass flow.

#### 7. No Backward Compatibility Fallback

Runtime deploys before UI. No catch-UNIMPLEMENTED fallback needed. Old frontend compile functions are removed immediately after migration.

### Implementation

**Scope:** All connectors (including HTTPS), all sources/models, and cleanup — shipped as one unit.

**Backend tasks:**
- [ ] Add `EnvVarName string` field to `PropertySpec` in `runtime/drivers/connectors.go`
- [ ] Add `EnvVarName` values to drivers with custom env var names (S3, GCS, BigQuery, etc.)
- [ ] Write automated Go test asserting `PropertySpec.Secret` matches expected values for all drivers; fix discrepancies
- [ ] Define `GenerateTemplateRequest`/`GenerateTemplateResponse` in `proto/rill/runtime/v1/api.proto` with validation annotations
- [ ] Run `buf generate` to regenerate Go and TypeScript proto bindings
- [ ] Implement `GenerateTemplate` handler in `runtime/server/generate_template.go`
  - Permission check (`EditRepo`)
  - Driver validation (must exist in `drivers.Connectors`)
  - Property validation (reject unknown keys)
  - `.env` read via `repo.Get()` for conflict resolution
  - YAML rendering via `yaml.Node` direct construction
  - Secret -> `{{ .env.VAR }}` replacement using `PropertySpec.Secret`
  - Env var naming from `PropertySpec.EnvVarName` with conflict suffix
  - Header comment with driver `DocsURL`
  - Value filtering (empty/nil)
  - String quoting for `StringPropertyType`
  - ClickHouse: exclude `managed: false` when default
  - Redshift: skip dev section
  - Model type: `type: model`, `materialize: true`, `connector:`, `sql:`, dev section
  - HTTPS headers: `formatHeadersAsYamlMap()` equivalent with auth-scheme splitting
  - DuckDB rewrite: detect `ImplementsObjectStore`/`ImplementsFileStore`/sqlite, build SQL from path + file extension, set `create_secrets_from_connectors`
  - Return actual `resource_type` and `driver` in response (may differ from request after rewrite)
- [ ] Write Go unit tests (see Testing Strategy below)
- [ ] Never include property values or env var values in error messages

**Frontend tasks:**
- [ ] Update `submitAddDataForm.ts` to call `GenerateTemplate` for all connectors
- [ ] Update `submitAddDataForm.ts` to call `GenerateTemplate` for source/model creation (send original driver + properties, no preprocessing)
- [ ] Use response `resource_type` and `driver` to determine file path (replaces frontend rewrite logic for path calculation)
- [ ] Strip `x-ui-only` fields from form values before sending (using `filterSchemaValuesForSubmit()`)
- [ ] Update `.env` merge logic to use `env_vars` response delta
- [ ] Remove `compileConnectorYAML()` and `compileSourceYAML()` (dead code)
- [ ] Remove `maybeRewriteToDuckDb()`, `buildDuckDbQuery()`, `extensionContainsParts()`, `prepareSourceFormData()` (dead code)
- [ ] Remove `updateDotEnvWithSecrets()`, `makeEnvVarKey()`, `findAvailableEnvVarName()`, `getGenericEnvVarName()` (dead code)

**Success criteria:**
- [ ] Creating any connector via the form produces identical YAML as before
- [ ] Creating any source/model via the form produces identical YAML as before
- [ ] S3/GCS/Azure/HTTPS/local_file/SQLite sources are correctly rewritten to DuckDB models with proper SQL
- [ ] `create_secrets_from_connectors` is set correctly for object store and HTTPS sources
- [ ] `.env` file is correctly updated with secrets
- [ ] All auth method variants work (S3 access_keys vs public, ClickHouse parameters vs DSN, etc.)
- [ ] HTTPS headers with sensitive auth tokens correctly extracted to `.env`
- [ ] Zero frontend YAML string building or driver rewriting remains

### Testing Strategy

New file: `runtime/server/generate_template_test.go`

Tests follow the codebase's existing pattern: table-driven tests using `testruntime.NewInstanceWithOptions()` and `server.NewServer()`, with `require.Contains()` / `require.Equal()` assertions on YAML output.

Note: The frontend functions being replaced (`compileSourceYAML`, `compileConnectorYAML`, `maybeRewriteToDuckDb`, `buildDuckDbQuery`) have **zero direct unit tests**. Only env var helpers are tested. The backend tests here are net-new coverage — there is no existing frontend test output to port.

#### Test 1: PropertySpec.Secret Flag Assertion

Automated test that asserts `PropertySpec.Secret` matches expected values for every registered driver. Catches drift between backend metadata and frontend `x-secret` annotations.

```go
func TestPropertySpecSecretFlags(t *testing.T) {
    // Expected secret keys per driver, derived from frontend x-secret annotations
    expected := map[string][]string{
        "s3":         {"aws_access_key_id", "aws_secret_access_key", "aws_role_arn", "aws_role_session_name", "aws_external_id"},
        "gcs":        {"google_application_credentials", "key_id", "secret"},
        "azure":      {"azure_storage_account", "azure_storage_key", "azure_storage_sas_token", "azure_storage_connection_string"},
        "clickhouse": {"dsn", "password", "write_dsn"},
        "postgres":   {"dsn", "password"},
        "bigquery":   {"google_application_credentials"},
        "snowflake":  {"dsn", "password", "privateKey"},
        "redshift":   {"aws_access_key_id", "aws_secret_access_key"},
        "motherduck": {"token"},
        "athena":     {"aws_access_key_id", "aws_secret_access_key"},
        "mysql":      {"dsn", "password"},
        "druid":      {"dsn", "password"},
        "pinot":      {"dsn", "password"},
        "starrocks":  {"dsn", "password"},
        "salesforce": {"password", "key"},
    }

    for driverName, driver := range drivers.Connectors {
        spec := driver.Spec()
        expectedKeys, ok := expected[driverName]
        if !ok {
            continue // AI/notifier drivers not relevant to GenerateTemplate
        }
        actualSecrets := secretKeys(spec.ConfigProperties)
        require.ElementsMatch(t, expectedKeys, actualSecrets, "driver %s", driverName)
    }
}
```

#### Test 2: Connector YAML Rendering (Table-Driven)

One test case per driver, covering the most common form submission for each. Validates the full YAML blob output.

```go
func TestBuildConnectorYAML(t *testing.T) {
    tt := []struct {
        name     string
        driver   string
        props    map[string]any
        contains []string // key substrings that must appear in output
        excludes []string // substrings that must NOT appear
    }{
        {
            name:   "clickhouse with parameters",
            driver: "clickhouse",
            props:  map[string]any{"host": "ch.example.com", "port": "9000", "password": "secret123"},
            contains: []string{
                "type: connector",
                "driver: clickhouse",
                `host: "ch.example.com"`,
                "port: 9000",
                `password: "{{ .env.CLICKHOUSE_PASSWORD }}"`,
                "# Connector YAML",
                "Reference documentation: https://docs.rilldata.com",
            },
            excludes: []string{"secret123"}, // actual secret value must never appear
        },
        {
            name:   "clickhouse with dsn",
            driver: "clickhouse",
            props:  map[string]any{"dsn": "clickhouse://user:pass@host:9000/db"},
            contains: []string{
                `dsn: "{{ .env.CLICKHOUSE_DSN }}"`,
            },
            excludes: []string{"clickhouse://user:pass"},
        },
        {
            name:   "s3 connector",
            driver: "s3",
            props:  map[string]any{"aws_access_key_id": "AKIA...", "aws_secret_access_key": "secret"},
            contains: []string{
                "driver: s3",
                `aws_access_key_id: "{{ .env.AWS_ACCESS_KEY_ID }}"`,
                `aws_secret_access_key: "{{ .env.AWS_SECRET_ACCESS_KEY }}"`,
            },
            excludes: []string{"AKIA", "secret"},
        },
        {
            name:   "bigquery connector",
            driver: "bigquery",
            props:  map[string]any{"project_id": "my-project", "google_application_credentials": `{"type":"service_account"}`},
            contains: []string{
                "driver: bigquery",
                `project_id: "my-project"`,
                `google_application_credentials: "{{ .env.GOOGLE_APPLICATION_CREDENTIALS }}"`,
            },
        },
        {
            name:   "postgres with individual params",
            driver: "postgres",
            props:  map[string]any{"host": "db.example.com", "port": "5432", "password": "pass"},
            contains: []string{
                "driver: postgres",
                `host: "db.example.com"`,
                `password: "{{ .env.POSTGRES_PASSWORD }}"`,
            },
        },
        {
            name:     "empty values filtered",
            driver:   "clickhouse",
            props:    map[string]any{"host": "ch.example.com", "port": "", "database": ""},
            contains: []string{"host:"},
            excludes: []string{"port:", "database:"},
        },
        {
            name:     "clickhouse managed false excluded when default",
            driver:   "clickhouse",
            props:    map[string]any{"host": "ch.example.com", "managed": false},
            excludes: []string{"managed"},
        },
    }

    for _, tc := range tt {
        t.Run(tc.name, func(t *testing.T) {
            // ... render and assert
        })
    }
}
```

**Drivers to cover:** clickhouse (params + DSN), postgres (params + DSN), s3, gcs, azure, bigquery, snowflake, redshift, athena, motherduck, duckdb, druid, pinot, starrocks, mysql, salesforce. That's 16 drivers × 1-2 variants each.

#### Test 3: Model YAML Rendering (Table-Driven)

Validates model output including `type: model`, `materialize: true`, `connector:`, SQL formatting, and dev section.

```go
func TestBuildModelYAML(t *testing.T) {
    tt := []struct {
        name     string
        driver   string
        props    map[string]any
        connName string
        contains []string
        excludes []string
    }{
        {
            name:     "clickhouse model with dev section",
            driver:   "clickhouse",
            props:    map[string]any{"sql": "SELECT * FROM events"},
            connName: "clickhouse_prod",
            contains: []string{
                "type: model",
                "materialize: true",
                "connector: clickhouse_prod",
                "sql: |",
                "  SELECT * FROM events",
                "dev:",
                "limit 10000",
                "# Model YAML",
            },
        },
        {
            name:     "redshift model without dev section",
            driver:   "redshift",
            props:    map[string]any{"sql": "SELECT * FROM events"},
            connName: "redshift_prod",
            contains: []string{"type: model", "connector: redshift_prod"},
            excludes: []string{"dev:"},
        },
        {
            name:     "bigquery model with dev section",
            driver:   "bigquery",
            props:    map[string]any{"sql": "SELECT * FROM `project.dataset.table`"},
            connName: "bq_prod",
            contains: []string{"dev:", "limit 10000"},
        },
    }
    // ...
}
```

#### Test 4: DuckDB Rewrite (Table-Driven)

Tests `maybeRewriteToDuckDB` and `buildDuckDBQuery` — the logic moving from frontend.

```go
func TestMaybeRewriteToDuckDB(t *testing.T) {
    tt := []struct {
        name          string
        driver        string
        props         map[string]any
        connectorName string
        wantDriver    string
        wantSQL       string
        wantSecrets   string // expected create_secrets_from_connectors value
        wantDeleted   []string // keys that should be removed from props
    }{
        // Object store drivers
        {
            name:          "s3 csv",
            driver:        "s3",
            props:         map[string]any{"path": "s3://bucket/data.csv"},
            connectorName: "my_s3",
            wantDriver:    "duckdb",
            wantSQL:       "select * from read_csv('s3://bucket/data.csv', auto_detect=true, ignore_errors=1, header=true)",
            wantSecrets:   "my_s3",
            wantDeleted:   []string{"path"},
        },
        {
            name:       "s3 parquet",
            driver:     "s3",
            props:      map[string]any{"path": "s3://bucket/data.parquet"},
            wantDriver: "duckdb",
            wantSQL:    "select * from read_parquet('s3://bucket/data.parquet')",
        },
        {
            name:       "s3 compressed parquet",
            driver:     "s3",
            props:      map[string]any{"path": "s3://bucket/data.v1.parquet.gz"},
            wantDriver: "duckdb",
            wantSQL:    "select * from read_parquet('s3://bucket/data.v1.parquet.gz')",
        },
        {
            name:       "gcs json",
            driver:     "gcs",
            props:      map[string]any{"path": "gs://bucket/data.json"},
            wantDriver: "duckdb",
            wantSQL:    "select * from read_json('gs://bucket/data.json', auto_detect=true, format='auto')",
        },
        {
            name:       "gcs ndjson",
            driver:     "gcs",
            props:      map[string]any{"path": "gs://bucket/data.ndjson"},
            wantDriver: "duckdb",
            wantSQL:    "select * from read_json('gs://bucket/data.ndjson', auto_detect=true, format='auto')",
        },
        {
            name:       "azure tsv",
            driver:     "azure",
            props:      map[string]any{"path": "azure://container/data.tsv"},
            wantDriver: "duckdb",
            wantSQL:    "select * from read_csv('azure://container/data.tsv', auto_detect=true, ignore_errors=1, header=true)",
        },
        {
            name:       "s3 unknown extension falls through",
            driver:     "s3",
            props:      map[string]any{"path": "s3://bucket/data.avro"},
            wantDriver: "duckdb",
            wantSQL:    "select * from 's3://bucket/data.avro'",
        },

        // File store drivers
        {
            name:       "https defaults to json",
            driver:     "https",
            props:      map[string]any{"path": "https://api.example.com/data"},
            wantDriver: "duckdb",
            wantSQL:    "select * from read_json('https://api.example.com/data', auto_detect=true, format='auto')",
        },
        {
            name:       "https with csv extension",
            driver:     "https",
            props:      map[string]any{"path": "https://example.com/data.csv"},
            wantDriver: "duckdb",
            wantSQL:    "select * from read_csv('https://example.com/data.csv', auto_detect=true, ignore_errors=1, header=true)",
        },
        {
            name:          "https with connector name sets secrets",
            driver:        "https",
            props:         map[string]any{"path": "https://api.example.com/data"},
            connectorName: "my_http",
            wantSecrets:   "my_http",
        },
        {
            name:       "local_file csv",
            driver:     "local_file",
            props:      map[string]any{"path": "/data/file.csv"},
            wantDriver: "duckdb",
            wantSQL:    "select * from read_csv('/data/file.csv', auto_detect=true, ignore_errors=1, header=true)",
            wantSecrets: "", // no create_secrets_from_connectors for local_file
        },

        // SQLite
        {
            name:        "sqlite",
            driver:      "sqlite",
            props:       map[string]any{"db": "/data/app.db", "table": "users"},
            wantDriver:  "duckdb",
            wantSQL:     "SELECT * FROM sqlite_scan('/data/app.db', 'users');",
            wantDeleted: []string{"db", "table"},
        },

        // Non-rewritable drivers pass through
        {
            name:       "clickhouse not rewritten",
            driver:     "clickhouse",
            props:      map[string]any{"sql": "SELECT 1"},
            wantDriver: "clickhouse",
        },
        {
            name:       "postgres not rewritten",
            driver:     "postgres",
            props:      map[string]any{"sql": "SELECT 1"},
            wantDriver: "postgres",
        },
    }
    // ...
}
```

#### Test 5: Env Var Naming and Conflict Resolution

Tests the `EnvVarName` field and `_1`, `_2` suffix logic. Mirrors the 48 frontend test cases in `code-utils.spec.ts`.

```go
func TestEnvVarNaming(t *testing.T) {
    tt := []struct {
        name        string
        driver      string
        propKey     string
        envVarName  string // PropertySpec.EnvVarName override
        existingEnv map[string]bool
        want        string
    }{
        // Schema-driven names (EnvVarName set)
        {name: "s3 access key", driver: "s3", propKey: "aws_access_key_id", envVarName: "AWS_ACCESS_KEY_ID", want: "AWS_ACCESS_KEY_ID"},
        {name: "bigquery creds", driver: "bigquery", propKey: "google_application_credentials", envVarName: "GOOGLE_APPLICATION_CREDENTIALS", want: "GOOGLE_APPLICATION_CREDENTIALS"},
        {name: "motherduck token", driver: "motherduck", propKey: "token", envVarName: "MOTHERDUCK_TOKEN", want: "MOTHERDUCK_TOKEN"},
        {name: "clickhouse password", driver: "clickhouse", propKey: "password", envVarName: "CLICKHOUSE_PASSWORD", want: "CLICKHOUSE_PASSWORD"},

        // Fallback naming (EnvVarName empty) → DRIVER_KEY format
        {name: "fallback format", driver: "custom", propKey: "api_token", want: "CUSTOM_API_TOKEN"},

        // Conflict resolution
        {name: "first conflict", driver: "bigquery", propKey: "google_application_credentials", envVarName: "GOOGLE_APPLICATION_CREDENTIALS",
            existingEnv: map[string]bool{"GOOGLE_APPLICATION_CREDENTIALS": true},
            want: "GOOGLE_APPLICATION_CREDENTIALS_1"},
        {name: "second conflict", driver: "bigquery", propKey: "google_application_credentials", envVarName: "GOOGLE_APPLICATION_CREDENTIALS",
            existingEnv: map[string]bool{"GOOGLE_APPLICATION_CREDENTIALS": true, "GOOGLE_APPLICATION_CREDENTIALS_1": true},
            want: "GOOGLE_APPLICATION_CREDENTIALS_2"},
        {name: "multi-key conflict (s3)", driver: "s3", propKey: "aws_access_key_id", envVarName: "AWS_ACCESS_KEY_ID",
            existingEnv: map[string]bool{"AWS_ACCESS_KEY_ID": true},
            want: "AWS_ACCESS_KEY_ID_1"},
    }
    // ...
}
```

#### Test 6: Property Validation

```go
func TestValidateProperties(t *testing.T) {
    tt := []struct {
        name         string
        driver       string
        resourceType string
        props        map[string]any
        wantErr      string
    }{
        {name: "valid connector props", driver: "clickhouse", resourceType: "connector", props: map[string]any{"host": "x"}, wantErr: ""},
        {name: "unknown prop rejected", driver: "clickhouse", resourceType: "connector", props: map[string]any{"bogus": "x"}, wantErr: `unknown property "bogus"`},
        {name: "source prop on connector rejected", driver: "duckdb", resourceType: "connector", props: map[string]any{"sql": "SELECT 1"}, wantErr: `unknown property "sql"`},
        {name: "source prop on model accepted", driver: "duckdb", resourceType: "model", props: map[string]any{"sql": "SELECT 1"}, wantErr: ""},
    }
    // ...
}
```

#### Test 7: Handler Integration Test

End-to-end test through the RPC handler, using the same test infrastructure as `generate_metrics_view_test.go`.

```go
func TestGenerateTemplate(t *testing.T) {
    rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
        Files: map[string]string{"rill.yaml": ""},
    })
    server, err := server.NewServer(ctx, &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient(), nil)
    require.NoError(t, err)

    tt := []struct {
        name         string
        req          *runtimev1.GenerateTemplateRequest
        wantContains []string
        wantEnvVars  map[string]string
        wantDriver   string // response driver (after rewrite)
        wantResType  string // response resource_type
        wantErr      codes.Code
    }{
        {
            name: "clickhouse connector",
            req: &runtimev1.GenerateTemplateRequest{
                InstanceId:   instanceID,
                ResourceType: "connector",
                Driver:       "clickhouse",
                Properties:   structpb("host", "ch.example.com", "password", "secret123"),
            },
            wantContains: []string{"type: connector", "driver: clickhouse", `{{ .env.CLICKHOUSE_PASSWORD }}`},
            wantEnvVars:  map[string]string{"CLICKHOUSE_PASSWORD": "secret123"},
            wantDriver:   "clickhouse",
            wantResType:  "connector",
        },
        {
            name: "s3 model rewritten to duckdb",
            req: &runtimev1.GenerateTemplateRequest{
                InstanceId:    instanceID,
                ResourceType:  "model",
                Driver:        "s3",
                Properties:    structpb("path", "s3://bucket/data.parquet"),
                ConnectorName: "my_s3",
            },
            wantContains: []string{"type: model", "connector: duckdb", "read_parquet", "create_secrets_from_connectors: my_s3"},
            wantDriver:   "duckdb",
            wantResType:  "model",
        },
        {
            name: "unknown driver rejected",
            req: &runtimev1.GenerateTemplateRequest{
                InstanceId:   instanceID,
                ResourceType: "connector",
                Driver:       "nonexistent",
                Properties:   structpb(),
            },
            wantErr: codes.InvalidArgument,
        },
        {
            name: "unknown property rejected",
            req: &runtimev1.GenerateTemplateRequest{
                InstanceId:   instanceID,
                ResourceType: "connector",
                Driver:       "clickhouse",
                Properties:   structpb("bogus_key", "value"),
            },
            wantErr: codes.InvalidArgument,
        },
        {
            name: "secret values never in error messages",
            req: &runtimev1.GenerateTemplateRequest{
                InstanceId:   instanceID,
                ResourceType: "connector",
                Driver:       "clickhouse",
                Properties:   structpb("bogus_key", "super_secret_value"),
            },
            wantErr: codes.InvalidArgument,
            // Additionally assert: !strings.Contains(err.Error(), "super_secret_value")
        },
    }
    // ...
}
```

#### Test 8: Env Var Conflict With Existing `.env`

Tests that the handler reads existing `.env` via `repo.Get()` and resolves conflicts correctly.

```go
func TestGenerateTemplateEnvConflict(t *testing.T) {
    rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
        Files: map[string]string{
            "rill.yaml": "",
            ".env":      "CLICKHOUSE_PASSWORD=old_value\n",
        },
    })
    // ... call GenerateTemplate with clickhouse password
    // Assert env_vars returns CLICKHOUSE_PASSWORD_1 (not CLICKHOUSE_PASSWORD)
}
```

#### Test Matrix Summary

| Test | What it validates | Cases |
|---|---|---|
| PropertySpec.Secret flags | Backend secret metadata matches frontend | 15 drivers |
| Connector YAML rendering | Full connector YAML output per driver | ~20 cases (16 drivers × 1-2 variants) |
| Model YAML rendering | Model output with dev section, SQL formatting | ~8 cases (warehouse drivers + DuckDB) |
| DuckDB rewrite | Driver/property transformation for object/file stores | ~15 cases (6 drivers × extension variants) |
| Env var naming | EnvVarName field, fallback format, conflict suffixes | ~10 cases |
| Property validation | Reject unknown keys, accept valid keys | ~5 cases |
| Handler integration | End-to-end RPC with test runtime | ~8 cases |
| Env conflict resolution | `.env` read + suffix logic | ~3 cases |

**Total: ~85 test cases** in `runtime/server/generate_template_test.go`.

## Alternative Approach Considered

### Consolidate Frontend Functions Only

Merge `compileConnectorYAML` and `compileSourceYAML` into a unified `compileResourceYAML()` on the frontend.

**Why rejected:** Doesn't solve the root problem (frontend owns YAML format knowledge), special cases keep accumulating, YAML string building is inherently fragile.

**When to reconsider:** If backend bandwidth is severely limited and we need a quick win.

## Acceptance Criteria

### Functional Requirements

- [ ] `GenerateTemplate` RPC produces valid YAML for all connectors and sources/models
- [ ] Output matches current frontend output for all connectors (verified by golden file tests)
- [ ] Object store drivers (s3, gcs, azure), file store drivers (https, local_file), and sqlite are rewritten to DuckDB models with correct SQL
- [ ] File extension → DuckDB read function mapping matches current frontend behavior (csv/tsv/txt → `read_csv`, parquet → `read_parquet`, json/ndjson → `read_json`, HTTPS default → `read_json`)
- [ ] Response `resource_type` and `driver` reflect actual values after rewrite
- [ ] Secret values replaced with `{{ .env.VAR }}` placeholders
- [ ] Env var names match current naming (including `EnvVarName` overrides)
- [ ] Env var conflicts resolved with `_1`, `_2`, etc. suffixes
- [ ] Unknown property keys rejected with `InvalidArgument`

### Quality Gates

- [ ] ~85 Go unit tests across 8 test categories (see Testing Strategy)
- [ ] `TestPropertySpecSecretFlags` passes for all 15 drivers
- [ ] `TestBuildConnectorYAML` covers all 16 form-relevant drivers
- [ ] `TestMaybeRewriteToDuckDB` covers all 6 rewritable drivers × file extension variants
- [ ] `TestGenerateTemplate` integration tests pass through full RPC handler
- [ ] Proto definitions pass `buf lint`
- [ ] No regression in existing connector creation flows
- [ ] Error messages never contain property values or secret data

## Dependencies & Prerequisites

- **Backend PropertySpec updates** — `EnvVarName` field and automated Secret flag test must be done before handler implementation
- **Proto generation pipeline** — `buf generate` must produce updated TypeScript bindings before frontend migration
- **Version rollout** — runtime deploys before UI; no backward-compatibility fallback needed

## Risk Analysis & Mitigation

| Risk | Severity | Mitigation |
|---|---|---|
| YAML output differs subtly from frontend | High | Golden file tests for every driver |
| Backend `PropertySpec.Secret` doesn't match frontend `x-secret` | High | Automated Go test asserting Secret flags |
| Env var naming changes break existing `.env` files | High | `EnvVarName` field matches exact current naming; test against real `.env` files |
| Dual-metadata divergence (PropertySpec vs frontend JSON Schemas) | Medium | Pre-ship audit; long-term: generate frontend schemas from PropertySpec |
| DuckDB rewrite SQL differs from frontend | High | Golden file tests for every rewritable driver (s3, gcs, azure, https, local_file, sqlite) with each file extension |
| TOCTOU race in `.env` (concurrent connector creations) | Medium | Backend reads `.env` directly; future: atomic read-modify-write |

## Technical Debt

1. **Dual-metadata problem:** `PropertySpec` (Go) and frontend JSON Schemas (TypeScript) describe overlapping property metadata. Generate frontend schemas from `PropertySpec` to eliminate this divergence.
2. **File path generation stays in frontend:** `getName()` and `getFileAPIPathFromNameAndType()` remain frontend-owned. Could eventually use response `driver` to infer path server-side.
3. **`rill.yaml` OLAP connector update stays in frontend:** Orchestration, not template rendering.

## References

- Brainstorm: `docs/brainstorms/2026-02-16-generate-template-api-brainstorm.md`
- Frontend YAML builders: `web-common/src/features/sources/sourceUtils.ts:24-113`, `web-common/src/features/connectors/code-utils.ts:158-256`
- Submission flow: `web-common/src/features/sources/modal/submitAddDataForm.ts`
- Existing Generate* RPCs: `runtime/server/generate_metrics_view.go`, `runtime/server/generate_resolver.go`, `runtime/server/generate_chart.go`, `runtime/server/generate_canvas_dashboard.go`
- Driver PropertySpec: `runtime/drivers/connectors.go:19-64`
- ClickHouse driver spec: `runtime/drivers/clickhouse/clickhouse.go:43+`
- Frontend schemas: `web-common/src/features/templates/schemas/*.ts`
- DuckDB rewrite (moving to backend): `web-common/src/features/sources/sourceUtils.ts:209-275`
- Backend driver specs with ObjectStore/FileStore flags: `runtime/drivers/s3/s3.go`, `runtime/drivers/gcs/gcs.go`, `runtime/drivers/azure/azure.go`, `runtime/drivers/https/https.go`, `runtime/drivers/file/file.go`
- Env var naming: `web-common/src/features/connectors/code-utils.ts:403-495`
