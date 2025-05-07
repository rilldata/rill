---
note: GENERATED. DO NOT EDIT.
title: API YAML
sidebar_position: 2
---

In your Rill project directory, create a new file name `<api-name>.yaml` in the `apis` directory containing a custom API definition. See comprehensive documentation on how to define and use [custom APIs](/integrate/custom-apis/index.md)

## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `api`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`refs`**  - _[array of oneOf]_ - List of resource references, each as a string or map. 

  *option 1* - _[object]_ - An object reference with at least a `<name>` and `<type>`.

  - **`type`**  - _[string]_ - type of resource 

  - **`name`**  - _[string]_ - name of resource  _(required)_

  *option 2* - _[string]_ - A string reference like `<resource-name>` or `<type/resource-name>`.

**`dev`**  - _[object]_ - Overrides properties in development 

**`prod`**  - _[object]_ - Overrides properties in production 

**`openapi`**  - _[object]_  

  - **`summary`**  - _[string]_  

  - **`request`**  - _[object]_  

    - **`parameters`**  - _[array of object]_  

  - **`response`**  - _[object]_  

    - **`schema`**  - _[object]_  

**`security`**  - _[object]_  

  - **`access`**  - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

    *option 1* - _[string]_ 

    *option 2* - _[boolean]_ 

  - **`row_filter`**  - _[string]_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause 

  - **`include`**  - _[array of object]_ - List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded 

    - **`if`**  - _[string]_ - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

    - **`names`**  - _[anyOf]_ - List of fields to include. Should match the name of one of the dashboard's dimensions or measures  _(required)_

      *option 1* - _[array of string]_ 

      *option 2* - _[string]_ 

  - **`exclude`**  - _[array of object]_ - List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included 

    - **`if`**  - _[string]_ - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

    - **`names`**  - _[anyOf]_ - List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures  _(required)_

      *option 1* - _[array of string]_ 

      *option 2* - _[string]_ 

  - **`rules`**  - _[array of object]_  

    - **`type`**  - _[string]_   _(required)_

    - **`action`**  - _[string]_  

    - **`if`**  - _[string]_  

    - **`names`**  - _[array of string]_  

    - **`all`**  - _[boolean]_  

    - **`sql`**  - _[string]_  

**`skip_nested_security`**  - _[boolean]_  

## One of Properties Options
- [sql](#sql)
- [metrics_sql](#metrics_sql)
- [api](#api)
- [glob](#glob)
- [resource_status](#resource_status)


### sql



**`sql`**  - _[string]_ - Raw SQL query to run against existing models in the project.  _(required)_

**`connector`**  - _[string]_ - specifies the connector to use when running SQL or glob queries. 

### metrics_sql



**`metrics_sql`**  - _[string]_ - SQL query that targets a metrics view in the project  _(required)_

### api



**`api`**  - _[string]_ - Name of a custom API defined in the project.  _(required)_

**`args`**  - _[object]_ - Arguments to pass to the custom API. 

### glob



**`glob`**  - _[anyOf]_ - Defines the file path or pattern to query from the specified connector.  _(required)_

  *option 1* - _[string]_ 

  *option 2* - _[object]_ 

**`connector`**  - _[string]_ - Specifies the connector to use with the glob input. 

### resource_status



**`resource_status`**  - _[object]_ - Based on resource status  _(required)_

  - **`where_error`**  - _[boolean]_ - Indicates whether the condition should trigger when the resource is in an error state. 

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

