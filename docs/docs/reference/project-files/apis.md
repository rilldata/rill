---
title: API YAML
sidebar_label: API YAML
sidebar_position: 50
---

In your Rill project directory, create a new file name `<api-name>.yaml` in the `apis` directory containing a custom API definition.
See comprehensive documentation on how to define and use [custom APIs](/integrate/custom-apis/index.md)

## Properties

**`type`** — Refers to the resource type and must be `api` _(required)_.

**`connector`** — Refers to the OLAP engine if not already set in rill.yaml or if using [multiple OLAP connectors](/connect/olap/multiple-olap) in a single project. Only applies when using `sql` _(optional)_.

Either one of the following:

- **`sql`** — General SQL query referring a [model](/build/models/sql-models) _(required)_.

- **`metrics_sql`** — SQL query referring metrics definition and dimensions defined in the [metrics view](/build/dashboards/dashboards.md) _(required)_.

**`skip_nested_security`** (boolean) - Ignore any security on referenced metrics views. Default `false` _(optional)_

**`openapi`** - Provide a OpenAPI specification for your endpoint _(optional)_
  - **`summary`** - Summary of your api
  - **`parameters`** - Accepted parameters, see [the parameters specification](https://swagger.io/specification/#parameter-object)
  - **`request_schema`** - Request schema, see the [request schema specification](https://swagger.io/specification/#request-body-object)
  - **`response_schema`** - Response schema, see the [response schema specification](https://swagger.io/specification/#response-object)

**`security`**
  - **`access`** - access policy to access the API endpoint
