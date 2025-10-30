---
title: Project Configuration
sidebar_label: Project Configuration
sidebar_position: 10
---

## Configuring `rill.yaml`

The `rill.yaml` file tells Rill which files to parse and which to ignore. By default, `rill.yaml` contains the following:

```yaml
compiler:
  sql_dialect: duckdb # The SQL dialect to use when compiling metrics views and APIs
```

You can specify additional file path patterns to include or exclude from parsing. (For more information on file path patterns, see [doublestar](https://github.com/bmatcuk/doublestar#patterns).)

```yaml
compiler:
  sql_dialect: duckdb
  include_paths:
    - "**/*.sql"
    - "**/*.yaml"
    - "**/*.yml"
  exclude_paths:
    - "tmp/**"
```

:::info

For more information, see [rill.yaml reference page](/reference/project-files/rill-yaml.md).

:::


## Configuring `rill-prod.yaml`

When deploying to Rill Cloud, `rill-prod.yaml` allows you to configure project-level cloud overrides. (This file is currently mainly used for [scheduling reports](/build/alerts/reports#scheduling-reports).)

:::note

The file is named `rill-prod.yaml` for legacy reasons. It applies to all cloud deployments, not just production deployments.

:::


## Specifying an OLAP connector

By default, Rill uses DuckDB as its OLAP engine, with data stored in a local `stage.db` file (or a bucket in Rill Cloud).

Rill also supports plugging in other OLAP engines. You can change the OLAP connector in `rill.yaml`:

```yaml
olap_connector: clickhouse
```

For most OLAP connectors, you also need to provide credentials either in a `.env` file or using `rill env configure` in Rill Cloud:

```bash
# in .env or configured using `rill env configure` (with some variation based on the connector)
olap.host=localhost
olap.port=9000
olap.database=db_name
olap.username=default
olap.password=pass
```

For the available OLAP connectors and their configuration, see the [OLAP catalog](/reference/olap-engines/olap-engines.md).


## Configuring variables in `rill.yaml`

You can configure project-level variables in `rill.yaml` by specifying a `vars` section. These variables can then be interpolated in different resources of your Rill project using templating syntax, such as `{{ .vars.my_variable }}`. For example:

```yaml
# Set project variables in rill.yaml
vars:
  my_variable: "Hello, world!"
```

```yaml
# models/my_model.yaml

# Reference the variable in a model
sql: SELECT '{{ .vars.my_variable }}' AS my_column

```

Note that variables will only work when templating is supported (such as in the `sql` property of models). For a full overview of templating in Rill, see [Templating](/build/templating).


## Configuring deployment defaults for Rill Cloud

In `rill.yaml`, you can add a `defaults` section to configure defaults for your deployment in Rill Cloud. The available defaults are:

- `olap_connector` – the OLAP connector for sources to use by default (useful if you have many sources)
- `slots` – the maximum concurrency of queries against the project's OLAP engine

Example `rill.yaml` with defaults:

```yaml
defaults:
  olap_connector: clickhouse
  slots: 10
```

:::info Want to tune your deployment?

Find more information on [deployment management](/manage/projects#deployment-management).

:::


## Feature flags

Rill uses feature flags to enable experimental or in-development features. Feature flags are boolean switches that can be set in `rill.yaml`.

### Available feature flags

- `chatCharts` – Enable AI chat with inline chart visualizations (default: `false`)
- `cloudDataViewer` – Enable viewing tabular data in the cloud UI (default: `false`)

### Enabling feature flags

To enable feature flags, add the `features` section to your `rill.yaml`:

```yaml
features:
  - chatCharts
  - cloudDataViewer
```

:::note

Feature flags are subject to change as features are developed and stabilized. Check the latest documentation or release notes for the most up-to-date list of available flags.

:::
