# Tech Design: YAML Generation — Imperative Functions vs. GenerateTemplate API

**Date:** 2026-02-16
**Status:** Brainstorm
**Author:** Cyrus Goh

---

## Problem Statement

The frontend has two imperative YAML builder functions — `compileSourceYAML()` and `compileConnectorYAML()` — that share overlapping logic but diverge in subtle ways. This creates:

1. **Duplication:** Both functions handle secret extraction (`{{ .env.* }}` placeholders), string quoting, empty-value filtering, and property ordering — but with separate, slightly different implementations.
2. **Maintenance burden:** Adding a new connector or field type requires touching complex conditional logic with many special cases (ClickHouse `managed: false` exclusion, DuckDB SQL rewriting, HTTP header formatting, etc.).
3. **Frontend owns YAML format:** The frontend knows too much about what valid YAML looks like for each resource type. This knowledge should live closer to the backend, which actually parses and validates these files.

### Current Architecture

```
Form Data → compileSourceYAML() / compileConnectorYAML() → YAML string → PutFile RPC → Backend
                  (imperative string builder)
```

**Key files:**
- `web-common/src/features/sources/sourceUtils.ts` — `compileSourceYAML()` (~90 lines)
- `web-common/src/features/connectors/code-utils.ts` — `compileConnectorYAML()` (~100 lines)
- `web-common/src/features/sources/modal/submitAddDataForm.ts` — primary caller
- `web-common/src/features/sources/modal/AddDataFormManager.ts` — YAML preview caller

### Shared Logic (duplicated between the two functions)

| Concern | compileSourceYAML | compileConnectorYAML |
|---------|-------------------|----------------------|
| YAML header with doc link | Yes | Yes |
| Secret → `{{ .env.VAR }}` | Yes | Yes |
| String property quoting | Yes | Yes |
| Empty value filtering | Yes | Yes |
| Env var name generation | Via `makeEnvVarKey()` | Via `makeEnvVarKey()` |
| SQL multi-line formatting | Yes | No |
| Headers map formatting | No | Yes |
| Property ordering | Implicit (object key order) | Explicit (orderedProperties) |
| Field filtering | Implicit (step-based) | Explicit (fieldFilter function) |
| Dev section | Yes (with Redshift exception) | No |

---

## Approach A: Consolidate Imperative Functions (Frontend-Only Refactor)

Merge `compileConnectorYAML` and `compileSourceYAML` into a unified `compileResourceYAML()` function that handles all resource types through configuration.

### Design

```typescript
interface ResourceYAMLOptions {
  resourceType: "connector" | "source" | "model";
  driver: string;
  formValues: Record<string, unknown>;
  orderedProperties?: ConnectorDriverProperty[];
  fieldFilter?: (property: ConnectorDriverProperty) => boolean;
  secretKeys?: string[];
  stringKeys?: string[];
  connectorInstanceName?: string;
  schema?: { properties?: Record<string, { "x-env-var-name"?: string }> };
  existingEnvBlob?: string;
  includeDevSection?: boolean;
  originalDriverName?: string;
}

function compileResourceYAML(opts: ResourceYAMLOptions): string {
  // Unified logic:
  // 1. Generate header (type + driver + doc link)
  // 2. Filter properties (by step, field filter, empty values)
  // 3. Order properties (explicit or implicit)
  // 4. Format each property:
  //    - Secrets → {{ .env.VAR }}
  //    - Strings → quoted
  //    - SQL → multi-line
  //    - Headers → YAML map
  //    - Default → raw value
  // 5. Optional dev section
  // 6. Return assembled YAML
}
```

### Migration Path

1. Create `compileResourceYAML()` with all shared logic
2. Re-implement `compileSourceYAML()` and `compileConnectorYAML()` as thin wrappers
3. Gradually migrate callers to the unified function
4. Remove wrappers once all callers are migrated

### Pros

- No backend changes required
- Incremental, low-risk refactor
- Eliminates duplication between the two functions
- Can be done in a single PR

### Cons

- Frontend still owns YAML format knowledge — can drift from backend expectations
- Special cases keep accumulating as connectors are added
- Doesn't extend to other resource types (explores, dashboards, etc.) without growing the function
- YAML string building is inherently fragile (indentation, quoting, escaping)

### Effort Estimate

Small — 1-2 days of frontend work.

---

## Approach B: Backend `GenerateTemplate` RPC (Recommended)

New backend endpoint that accepts structured data and returns rendered YAML. The frontend becomes a thin form layer.

### Design

#### New RPC

```protobuf
// GenerateTemplate renders a YAML file from structured input.
// Supports all resource types: connector, source, model, explore, dashboard, etc.
rpc GenerateTemplate(GenerateTemplateRequest) returns (GenerateTemplateResponse) {
  option (google.api.http) = {
    post: "/v1/instances/{instance_id}/generate/template",
    body: "*"
  };
}

message GenerateTemplateRequest {
  string instance_id = 1;
  // Resource type: "connector", "source", "model", "explore", "metrics_view", etc.
  string resource_type = 2;
  // Driver name (for connectors/sources): "clickhouse", "s3", "duckdb", etc.
  string driver = 3;
  // Structured key-value properties from the form
  google.protobuf.Struct properties = 4;
  // Optional: connector instance name (for sources that reference a connector)
  string connector_name = 5;
  // Optional: fields that should be treated as secrets (extracted to .env)
  repeated string secret_keys = 6;
}

message GenerateTemplateResponse {
  // Rendered YAML blob, ready to write via PutFile
  string blob = 1;
  // Environment variables to write to .env (key → value)
  map<string, string> env_vars = 2;
}
```

#### New Architecture

```
Form Data → GenerateTemplate RPC → { blob, env_vars } → PutFile RPC (blob) + update .env (env_vars)
                (backend)
```

#### Backend Implementation

The backend would:
1. Look up a Go template for the given `resource_type` + `driver` combination
2. Apply structured properties to the template
3. Handle secret extraction: replace secret values with `{{ .env.VAR }}` and return the real values in `env_vars`
4. Return the rendered YAML blob

Templates could be Go `text/template` files or a simple struct-to-YAML mapper using the `gopkg.in/yaml.v3` library.

#### Frontend Changes

```typescript
// Before (imperative)
const blob = compileConnectorYAML(connector, formValues, { ...options });
await runtimeServicePutFile(instanceId, { path, blob, create: true });
await updateDotEnvWithSecrets(instanceId, connector, formValues, options);

// After (declarative)
const { blob, envVars } = await runtimeServiceGenerateTemplate(instanceId, {
  resourceType: "connector",
  driver: connector.name,
  properties: formValues,
  secretKeys: schemaSecretKeys,
});
await runtimeServicePutFile(instanceId, { path, blob, create: true });
await writeDotEnv(instanceId, envVars);
```

### Migration Path

1. **Phase 1:** Implement `GenerateTemplate` RPC for connectors
2. **Phase 2:** Migrate `compileConnectorYAML()` callers to use the RPC
3. **Phase 3:** Add source/model support, migrate `compileSourceYAML()` callers
4. **Phase 4:** Extend to other resource types (explores, dashboards, etc.)
5. **Phase 5:** Remove frontend compile functions

Each phase can be a separate PR. Old and new paths can coexist during migration.

### Pros

- Backend owns the canonical YAML format — single source of truth
- Frontend complexity drops dramatically — no YAML string building, quoting, or escaping
- Generalizes to all resource types (connectors, sources, models, explores, dashboards, APIs, canvases, themes, reports, alerts)
- Aligns with existing backend generation patterns (`GenerateMetricsViewFile`, `GenerateCanvas`, `GenerateResolver`)
- Secret handling can be co-located with template logic
- Backend can validate property completeness before rendering

### Cons

- Requires backend work (new RPC, template registry, tests)
- Larger scope — multi-phase migration across frontend and backend
- YAML preview in the form is lost (acceptable per discussion — not critical)
- Network round-trip for template rendering (acceptable — only on submit)
- Need to decide where templates live (Go code, embedded files, or database)

### Effort Estimate

Medium — 1-2 weeks across backend + frontend.

---

## Open Questions

1. **Secret handling ownership:** Should the backend fully own env var naming and `.env` file writes? Or should it just return `env_vars` and let the frontend write them?
   - Option A: Backend returns `env_vars` map, frontend writes to `.env` (simpler, frontend already has this code)
   - Option B: Backend writes `.env` directly as part of `GenerateTemplate` (cleaner, but couples template rendering with file I/O)

2. **Template storage:** Where do the YAML templates live?
   - Go struct-to-YAML mapping (type-safe, no template files)
   - Embedded Go `text/template` files (flexible, easy to edit)
   - Hardcoded strings in Go (simple, like current frontend approach but server-side)

3. **DuckDB rewriting:** Currently `maybeRewriteToDuckDb()` transforms S3/GCS/HTTPS sources into DuckDB SQL. Should this logic move to the backend too, or stay as a frontend preprocessing step?

4. **Connector property metadata:** The frontend schemas define `x-secret`, `x-string`, `x-env-var-name` extensions. Should the backend be the source of truth for these, or should the frontend continue to pass `secretKeys`/`stringKeys` in the request?

5. **Backward compatibility:** Do we need to support both old (imperative) and new (RPC) paths simultaneously during migration? Or can we cut over all at once?

---

## Recommendation

**Approach B (Backend `GenerateTemplate` RPC)** is recommended because:

- It addresses the root cause: the frontend shouldn't own YAML format knowledge
- It's a natural extension of existing backend patterns (`GenerateMetricsViewFile`, etc.)
- It generalizes to all resource types, not just connectors/sources
- The migration can be incremental (phase by phase)
- Loss of YAML preview is acceptable

Approach A is a valid stopgap if backend bandwidth is limited, but it doesn't solve the fundamental problem of frontend-owned YAML format.

---

## Key Decisions Made

- **Motivation:** Duplication and maintenance burden in `compileSourceYAML`/`compileConnectorYAML`
- **Direction:** Backend RPC over frontend-only refactor
- **Scope:** All resource types (design wide), implement connectors/sources first
- **YAML Preview:** Can be dropped — not critical for users
- **Secret handling:** Open question to resolve during planning
