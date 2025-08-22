---
note: GENERATED. DO NOT EDIT.
title: API YAML
sidebar_position: 32
---

In your Rill project directory, create a new file name `<api-name>.yaml` in the `apis` directory containing a custom API definition. See comprehensive documentation on how to define and use [custom APIs](/build/custom-apis)

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `api` _(required)_

### `openapi`

_[object]_ - OpenAPI specification for the API endpoint 

  - **`summary`** - _[string]_ - A brief description of what the API endpoint does 

  - **`parameters`** - _[array of object]_ - List of parameters that the API endpoint accepts 

  - **`request_schema`** - _[object]_ - JSON schema for the request body (use nested YAML instead of a JSON string) 

  - **`response_schema`** - _[object]_ - JSON schema for the response body (use nested YAML instead of a JSON string) 

### `security`

_[object]_ - Defines security rules and access control policies for resources 

  - **`access`** - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

    - **option 1** - _[string]_ - SQL expression that evaluates to a boolean to determine access

    - **option 2** - _[boolean]_ - Direct boolean value to allow or deny access

  - **`row_filter`** - _[string]_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause 

  - **`include`** - _[array of object]_ - List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded 

    - **`if`** - _[string]_ - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_

    - **`names`** - _[anyOf]_ - List of fields to include. Should match the name of one of the dashboard's dimensions or measures _(required)_

      - **option 1** - _[array of string]_ - List of specific field names to include

      - **option 2** - _[string]_ - Wildcard '*' to include all fields

  - **`exclude`** - _[array of object]_ - List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included 

    - **`if`** - _[string]_ - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean _(required)_

    - **`names`** - _[anyOf]_ - List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures _(required)_

      - **option 1** - _[array of string]_ - List of specific field names to exclude

      - **option 2** - _[string]_ - Wildcard '*' to exclude all fields

  - **`rules`** - _[array of object]_ - List of detailed security rules that can be used to define complex access control policies 

    - **`type`** - _[string]_ - Type of security rule - access (overall access), field_access (field-level access), or row_filter (row-level filtering) _(required)_

    - **`action`** - _[string]_ - Whether to allow or deny access for this rule 

    - **`if`** - _[string]_ - Conditional expression that determines when this rule applies. Must be a valid SQL expression that evaluates to a boolean 

    - **`names`** - _[array of string]_ - List of field names this rule applies to (for field_access type rules) 

    - **`all`** - _[boolean]_ - When true, applies the rule to all fields (for field_access type rules) 

    - **`sql`** - _[string]_ - SQL expression for row filtering (for row_filter type rules) 

### `skip_nested_security`

_[boolean]_ - Flag to control security inheritance 

## One of Properties Options
- [SQL Query](#sql-query)
- [Metrics View Query](#metrics-view-query)
- [Custom API Call](#custom-api-call)
- [File Glob Query](#file-glob-query)
- [Resource Status Check](#resource-status-check)

## SQL Query

Executes a raw SQL query against the project's data models.

### `sql`

_[string]_ - Raw SQL query to run against existing models in the project. _(required)_

### `connector`

_[string]_ - specifies the connector to use when running SQL or glob queries. 

## Metrics View Query

Executes a SQL query that targets a defined metrics view.

### `metrics_sql`

_[string]_ - SQL query that targets a metrics view in the project _(required)_

## Custom API Call

Calls a custom API defined in the project to compute data.

### `api`

_[string]_ - Name of a custom API defined in the project. _(required)_

### `args`

_[object]_ - Arguments to pass to the custom API. 

## File Glob Query

Uses a file-matching pattern (glob) to query data from a connector.

### `glob`

_[anyOf]_ - Defines the file path or pattern to query from the specified connector. _(required)_

  - **option 1** - _[string]_ - A simple file path/glob pattern as a string.

  - **option 2** - _[object]_ - An object-based configuration for specifying a file path/glob pattern with advanced options.

### `connector`

_[string]_ - Specifies the connector to use with the glob input. 

## Resource Status Check

Uses the status of a resource as data.

### `resource_status`

_[object]_ - Based on resource status _(required)_

  - **`where_error`** - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

## Examples

```yaml
# Example: This api returns the top 10 authors by net line changes since the specified date provided in the arguments.
type: api
name: metrics_view_api
metrics_sql: |-
    SELECT author_name, net_line_changes
    FROM advanced_metrics_view
      where author_date > '{{ .args.date }}'
      order by net_line_changes DESC
      limit 10
```