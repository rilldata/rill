# Template definitions

JSON files in this tree are the source of truth for the connector and OLAP forms shown in the Rill add-data flow. Each file describes one template: a JSON Schema (which drives the form on the frontend) plus one or more output files (rendered on the backend through `text/template`).

The runtime loads every `definitions/*/*.json` file at process start via `embed.FS` (see `registry.go`). Stub files — empty or containing only a `_reason` field — are skipped, so an empty JSON file can be used as a placeholder while a connector is being designed.

## Layout

```
definitions/
  olap/                 connectors that act as the project's OLAP engine
  duckdb-models/        source connectors targeting DuckDB OLAP
  clickhouse-models/    source connectors targeting ClickHouse OLAP
```

Templates are keyed by their `name` field, not by file path. Filenames follow the convention `<driver>-<olap>.json` (e.g. `s3-duckdb.json`, `s3-clickhouse.json`) or just `<driver>.json` for OLAP connectors.

## Template shape

```jsonc
{
  "name": "s3-duckdb",
  "display_name": "Amazon S3",
  "driver": "s3",
  "olap": "duckdb",
  "icon": "AmazonS3",
  "small_icon": "AmazonS3Icon",
  "tags": ["source", "duckdb", "s3", "objectStore"],
  "description": "...",
  "docs_url": "https://docs.rilldata.com/...",
  "json_schema": { /* JSON Schema with x-* extensions, see below */ },
  "files": [
    { "name": "connector", "path_template": "connectors/[[ .connector_name ]].yaml", "code_template": "..." },
    { "name": "model",     "path_template": "models/[[ .model_name ]].yaml",         "code_template": "..." }
  ]
}
```

### File outputs

Each entry in `files` produces one rendered file:

- `name` — `"connector"` or `"model"`. The `GenerateFile` RPC accepts an `output` filter to render a single entry (used for previewing one step of a multi-step flow).
- `path_template` — Go `text/template` for the output path, relative to the project root.
- `code_template` — inline Go `text/template` for the file contents.
- `code_template_file` — alternative to `code_template`: a path (relative to the JSON definition) to a separate `.tmpl` file. Useful for long templates. If both are set, the file wins.

Templates use **`[[ ]]` delimiters** (not `{{ }}`) so they don't collide with Rill's runtime templating syntax (`{{ .env.FOO }}`) inside the rendered YAML.

### Property order

JSON object key order is preserved at load time (`extractPropertyOrder` in `registry.go`) and re-exposed as `x-property-order` for the frontend. This means the order of fields in `json_schema.properties` is the order they appear in the form and the order their values are emitted by `renderProps`.

## JSON Schema extensions

Standard JSON Schema fields (`type`, `properties`, `required`, `enum`, `default`, `description`) work as expected. The following `x-*` extensions are also recognised. Most are interpreted by both the backend renderer and the frontend form — keep them in sync.

### Form structure

| Key | Type | Meaning |
|---|---|---|
| `x-step` | `"connector"` \| `"source"` \| `"explorer"` | Which form step a property belongs to. Connector-step props are rendered into the connector YAML; source-step props into the model YAML. Explorer-step props are accessed directly as template variables (e.g. `[[ .sql ]]`) and are not emitted by `renderProps`. Props without `x-step` go into both connector and source outputs. |
| `x-grouped-fields` | `{ enumValue: [...] }` | When the parent property is a radio/select with enum values, lists the fields shown when each value is selected. Used by the frontend to scope visibility to the active branch. |
| `x-visible-if` | `{ otherKey: [allowedValue, ...] }` | Conditional visibility: show this field only when the named field has one of the listed values. |
| `x-tab-group` | `{ tabName: [fieldKey, ...] }` | Groups fields under named tabs in the form. |
| `x-form-height` | `"tall"` | Frontend hint to use a taller form layout (currently used by Snowflake and Salesforce). |
| `x-form-width` | `"wide"` | Frontend hint to use a wider form layout. |
| `x-category` | `"objectStore"` \| `"warehouse"` \| `"sqlStore"` \| `"olap"` \| `"sourceOnly"` | Connector category, used to organise the picker. |

### Field display

| Key | Type | Meaning |
|---|---|---|
| `x-display` | `"radio"` \| `"select"` \| `"file"` \| ... | Override the default input control. |
| `x-select-style` | `"rich"` | Render a select with icons / descriptions instead of plain options. |
| `x-placeholder` | string | Placeholder text for text inputs. |
| `x-hint` | string | Helper text shown beneath the input. |
| `x-enum-labels` | array | Display labels for `enum` values, in the same order. |
| `x-enum-descriptions` | array | Sub-labels / descriptions for `enum` values. |
| `x-button-labels` | nested object | Custom labels for nested radio groups (see `clickhouse.json` for an example). |
| `x-informational` | bool | Marks a read-only / explanatory field that is not part of the rendered output. |

### File uploads

| Key | Type | Meaning |
|---|---|---|
| `x-file-accept` | string | Comma-separated `accept` filter for file inputs (e.g. `".json"`, `".pem,.p8"`). |
| `x-file-encoding` | `"base64"` \| `"json"` | How the uploaded file should be encoded before submission. |
| `x-file-extract` | `{ schemaKey: jsonPath }` | When the upload is a JSON file, populate other form fields from values inside it (used by BigQuery's service-account flow). |

### Backend / rendering

| Key | Type | Meaning |
|---|---|---|
| `x-secret` | bool | Value is a credential. The renderer extracts it to a `.env` variable and emits `{{ .env.<NAME> }}` in the YAML. |
| `x-env-var` | string | Override the default env var name (`<DRIVER>_<KEY>`). When the chosen name already exists in `.env`, the renderer appends a numeric suffix (`FOO`, `FOO_1`, `FOO_2`, ...). |
| `x-omit-if-default` | bool | Skip rendering the property when its value equals the schema `default`. Replaces the per-driver "skip if `key == "managed"`" hack used in v1. |
| `x-ui-only` | bool | Property exists only to drive the form (e.g. an `auth_method` radio that selects between branches). It is never written to the rendered YAML. |

## Template helpers

`text/template` data passed to each file includes:

- `.driver`, `.connector_name`, `.docs_url`, `.model_name` — basics.
- `.<schemaKey>` — every property declared in `json_schema.properties` is exposed as a top-level key. Empty values fall back to the schema `default` if defined, otherwise to `""`. This means conditional template branches can rely on the key existing.
- `.props`, `.config_props`, `.source_props` — pre-processed `[]ProcessedProp` slices for use with `renderProps`. `config_props` and `source_props` correspond to `x-step: connector` and `x-step: source`; `.props` is an alias for `config_props`.

Functions registered in `funcmap.go`:

| Name | Purpose |
|---|---|
| `renderProps` | Renders a `[]ProcessedProp` as YAML key/value lines, quoting strings/secrets and leaving numbers and booleans bare. |
| `propVal` | Looks up a single property value by key (used inside conditional branches). |
| `default` | Positional `[[ default (expr) "fallback" ]]` — returns the fallback when the value is empty. **Do not use pipeline syntax** (`[[ expr \| default "fallback" ]]`); `text/template` pipes into the last argument and would swap the values. |
| `indent`, `quote` | String helpers. |
| `duckdbSQL` | Maps a file path to the appropriate DuckDB `read_*` call based on the extension. |
| `s3ToHTTPS`, `gcsToHTTPS` | Convert `s3://` / `gs://` URIs to HTTPS for ClickHouse's `s3()` / `gcs()` functions. |
| `azureContainer`, `azureBlobPath`, `azureEndpoint` | Decompose Azure URIs into the parts ClickHouse's `azureBlobStorage()` function expects. |
| `clickhouseFormat`, `clickhouseURLSuffix` | Map URL extensions to ClickHouse input formats and emit the optional `, Format, headers(...)` suffix when custom HTTPS headers are present. |

See `funcmap.go` for the exact behaviour of each function and `render_test.go` for golden-output examples.

## Authoring a new template

1. Create a JSON file under the appropriate group (`olap/`, `duckdb-models/`, `clickhouse-models/`).
2. Write the JSON Schema for the form. Use `x-step` to split fields between the connector and model steps when the form is multi-step.
3. Add the `files` entries with `path_template` and `code_template` (or `code_template_file`).
4. Add a golden test case in `../render_test.go`.
5. Run `go test ./runtime/templates/...` to verify.
