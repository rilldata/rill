---
title: Project Configuration
sidebar_label: Project Configuration
sidebar_position: 60
---

## Project configuration options

The following is a comprehensive list of options that can be configured via `rill.yaml`.

### Title

The title of the project. If not provided, it will be inferred from the project directory name. Note that this is _NOT_ used to set the title of the deployment for your project (which is set via `rill deploy`).

```yaml
title: My Rill Project
```

### Public

_Only applicable for `rill deploy`_

If set to `true`, the project will be accessible to everyone on the internet (no login required).

```yaml
public: true
```

### OLAP Connector

The OLAP connector to use for the project. The default is `duckdb`.

```yaml
olap_connector: duckdb
```

### Defaults

You can specify default values for all dashboards in the project using the `defaults` key. For example, the following configuration sets the default time range to 7 days and the comparison mode to time for all dashboards:

```yaml
defaults:
  dashboards:
    time_range: P7D
    comparison_mode: time
```

The full list of options that can be configured for dashboards is:
- `dimensions` - default dimensions to show in the leaderboard
- `measures` - default measures to show in the leaderboard
- `time_range` - default time range for the dashboard
- `comparison_mode` - default comparison mode for the dashboard (`time`, `dimension`, or `none`)
- `comparison_dimension` - default comparison dimension (only applicable if `comparison_mode` is `dimension`)
- `available_time_zones` - list of time zones that the user can select from
- `default_time_zone` - default time zone for the dashboard

**Note:** Users can override these defaults in the UI, and the overrides will be persisted in the URL.

### Variables

Variables can be used to parameterize SQL queries and dashboard filters. They can be defined in `rill.yaml` and referenced in SQL queries using the `{{ .vars.variable_name }}` syntax.

```yaml
vars:
  database: my_database
  schema: my_schema
```

You can then reference these variables in your SQL queries:

```sql
SELECT * FROM {{ .vars.database }}.{{ .vars.schema }}.my_table
```

#### Variable interpolation

You can use variable interpolation to reference other variables in your `rill.yaml` file. For example:

```yaml
vars:
  database: my_database
  schema: my_schema
  table: "{{ .vars.database }}.{{ .vars.schema }}.my_table"
```

You can then reference the `table` variable in your SQL queries:

```sql
SELECT * FROM {{ .vars.table }}
```

:::note
You can also interpolate environment variables using the `{{ .env.VAR_NAME }}` syntax.
:::

#### User attributes

Variables can also be set on a per-user basis by admins. See [User Attributes](../manage/user-management.md#user-attributes) for more details.

### Environment-specific configuration

You can specify environment-specific variables in `rill.yaml` using the `env` key. For example, you might want to use a different database in production than in development:

```yaml
env:
  dev:
    vars:
      database: dev_database
  prod:
    vars:
      database: prod_database
```

When running `rill start`, Rill will use the `dev` environment by default. When running `rill deploy`, Rill will use the `prod` environment by default. You can also specify a custom environment using the `--env` flag.

:::note
Environment variables (defined in `.env`) are also environment-specific. See [Credentials](./credentials.md) for more details.
:::

### Catalog

Rill can populate a data catalog from external metadata sources. The catalog can be viewed in the Rill UI on the _Catalog_ page.

To enable catalog population from a connector, use the `catalog` key:

```yaml
catalog:
  my_catalog:
    connector: my_connector
```

For more details, see [Catalog](./catalog.md).

### AI

You can configure AI features for your project using the `ai` key. Currently, the only supported AI feature is the AI chat interface.

```yaml
ai:
  chat:
    max_samples: 5
```

The following options are available for AI chat:
- `max_samples` - the maximum number of data samples to include in AI responses (default: 5, max: 10)

### Features

Features are used to enable experimental or optional functionality in Rill. They can be enabled by adding them to the `features` list in `rill.yaml`.

```yaml
features:
  - feature_name
```

#### Available features

The following features are available:

- `chatCharts` - Enable inline chart visualizations in AI chat responses (default: false)
- `alertsEditAndDelete` - Enable editing and deleting alerts through the UI (default: false)
- `cloudDataViewer` - Enable viewing data from cloud storage in the UI (default: false)

**Example: Enable cloud data viewer**

```yaml
features:
  - cloudDataViewer
```

## Environment variables

You can use environment variables in `rill.yaml` by referencing them using the `{{ .env.VAR_NAME }}` syntax. For example:

```yaml
vars:
  database: "{{ .env.DATABASE_NAME }}"
```

You can set environment variables in a `.env` file in your project directory. See [Credentials](./credentials.md) for more details.

## Migrating from `rill.yaml` to dashboard YAML

In earlier versions of Rill, dashboard defaults were configured in `rill.yaml`. This has been deprecated in favor of configuring defaults directly in the dashboard YAML files.

If you have a `rill.yaml` file with dashboard defaults, you should migrate them to the dashboard YAML files. For example, if you have the following in `rill.yaml`:

```yaml
defaults:
  dashboards:
    time_range: P7D
    comparison_mode: time
```

You should move these options to your dashboard YAML files:

```yaml
type: metrics_view
# ... other dashboard configuration
time_range: P7D
comparison_mode: time
```

After migrating, you can remove the `defaults` key from `rill.yaml`.
