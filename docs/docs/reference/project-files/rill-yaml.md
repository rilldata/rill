---
note: GENERATED. DO NOT EDIT.
title: Project YAML
sidebar_position: 42
---

The `rill.yaml` file contains metadata about your project.

## Properties

### `compiler`

_[string]_ - Specifies the parser version to use for compiling resources 

### `display_name`

_[string]_ - The display name of the project, shown in the upper-left corner of the UI 

### `description`

_[string]_ - A brief description of the project 

### `features`

_[object]_ - Optional feature flags. Can be specified as a map of feature names to booleans. 

### `ai_connector`

_[string]_ - Specifies the default AI connector for the project. Defaults to Rill's internal AI connector if not set. 

### `ai_instructions`

_[string]_ - Extra instructions for LLM/AI features. Used to guide natural language question answering and routing. 

## Configuring the default OLAP Engine

Rill allows you to specify the default OLAP engine to use in your project via `rill.yaml`.
:::info Curious about OLAP Engines?
Please see our reference documentation on [OLAP Engines](/developers/build/connectors/olap).
:::


### `olap_connector`

_[string]_ - Specifies the default OLAP engine for the project. Defaults to duckdb if not set. 

```yaml
olap_connector: clickhouse
```

## Project-wide defaults

In `rill.yaml`, project-wide defaults can be specified for a resource type within a project. Unless otherwise specified, _individual resources will inherit any defaults_ that have been specified in `rill.yaml`. For available properties that can be configured, please refer to the YAML specification for each individual resource type - [model](models.md), [metrics_view](metrics-views.md), and [explore](explore-dashboards.md)

:::note Use plurals when specifying project-wide defaults
In your `rill.yaml`, the top level property for the resource type needs to be **plural**, such as `models`, `metrics_views` and `explores`.
:::

:::info Hierarchy of inheritance and property overrides
As a general rule of thumb, properties that have been specified at a more _granular_ level will supercede or override higher level properties that have been inherited. Therefore, in order of inheritance, Rill will prioritize properties in the following order:
1. Individual [models](models.md)/[metrics_views](metrics-views.md)/[explore](explore-dashboards.md) object level properties (e.g. `models.yaml` or `explore-dashboards.yaml`)
2. [Environment](/developers/build/models/templating) level properties (e.g. a specific property that have been set for `dev`)
3. [Project-wide defaults](#project-wide-defaults) for a specific property and resource type
:::


### `models`

_[object]_ - Defines project-wide default settings for models. Unless overridden, individual models will inherit these defaults. 

### `metrics_views`

_[object]_ - Defines project-wide default settings for metrics_views. Unless overridden, individual metrics_views will inherit these defaults. 

### `explores`

_[object]_ - Defines project-wide default settings for explores. Unless overridden, individual explores will inherit these defaults. 

### `canvases`

_[object]_ - Defines project-wide default settings for canvases. Unless overridden, individual canvases will inherit these defaults. 

```yaml
# For complete examples, see: 
# https://docs.rilldata.com/developers/build/rill-project-file#dashboard-defaults
models:
    refresh:
        cron: '0 * * * *'
metrics_views:
    first_day_of_week: 1
    smallest_time_grain: month
explores:
    defaults:
        time_range: P24M
    time_zones:
        - UTC
    time_ranges:
        - PT24H
        - P6M
canvases:
    defaults:
        time_range: P7D
    time_zones:
        - UTC
    time_ranges:
        - PT24H
        - P7D
```

## Setting variables

Primarily useful for [templating](/developers/build/connectors/templating), variables can be set in the `rill.yaml` file directly. This allows variables to be set for your projects deployed to Rill Cloud while still being able to use different variable values locally if you prefer. 
:::info Overriding variables locally
Variables also follow an order of precedence and can be overridden locally. By default, any variables defined will be inherited from `rill.yaml`. However, if you manually pass in a variable when starting Rill Developer locally via the CLI, this value will be used instead for the current instance of your running project:
```bash
rill start --env numeric_var=100 --env string_var="different_value"
```
:::
:::tip Setting variables through `.env`
Variables can also be set through your project's `<RILL_PROJECT_HOME>/.env` file (or using the `rill env set` CLI command), such as:
```bash
variable=xyz
```
Similar to how [connector credentials can be pushed / pulled](/developers/build/connectors/credentials#pulling-credentials-and-variables-from-a-deployed-project-on-rill-cloud) from local to cloud or vice versa, project variables set locally in Rill Developer can be pushed to Rill Cloud and/or pulled back to your local instance from your deployed project by using the `rill env push` and `rill env pull` commands respectively.
:::


### `env`

_[object]_ - A map of key-value pairs for setting variables on your project. It accepts both user-defined variables (for use with templating) and reserved `rill.*` keys that configure project-wide settings. The full set of reserved keys is listed below.
 

  - **`rill.download_limit_bytes`** - _[integer]_ - Limit on the size of an exported file, in bytes. Default: 134217728 (128 MB). 

  - **`rill.interactive_sql_row_limit`** - _[integer]_ - Row limit for interactive SQL queries; does not apply to SQL exports. Default: 10000. 

  - **`rill.models.default_materialize`** - _[boolean]_ - Materialize models as tables by default instead of views. Default: false. 

  - **`rill.models.materialize_delay_seconds`** - _[integer]_ - Delay before materializing models, in seconds. Default: 0. 

  - **`rill.models.concurrent_execution_limit`** - _[integer]_ - Maximum number of concurrent model executions. Default: 5. 

  - **`rill.model.timeout_override`** - _[integer]_ - Timeout for model reconciliation in seconds, used in validation mode. Default: 0 (no override). 

  - **`rill.model.partitions_warn_on_failure`** - _[boolean]_ - When true, partition execution failures are surfaced as non-blocking warnings instead of errors. Default: true in `prod`, false otherwise. 

  - **`rill.model.tests_warn_on_failure`** - _[boolean]_ - When true, model test failures are surfaced as non-blocking warnings instead of errors. Default: true in `prod`, false otherwise. 

  - **`rill.models.disable`** - _[boolean]_ - When true, model execution is disabled. Useful for stopping any ingestion in Rill temporarily. Default: false. 

  - **`rill.metrics.approximate_comparisons`** - _[boolean]_ - Rewrite metrics comparison queries to use an approximate, faster form. Approximate comparisons may not return data points for all values. Default: true. 

  - **`rill.metrics.approximate_comparisons_cte`** - _[boolean]_ - Rewrite metrics comparison queries to use a CTE for the base query. Default: false. 

  - **`rill.metrics.approximate_comparisons_two_phase_limit`** - _[integer]_ - Row-limit threshold under which metrics comparison queries use a two-phase strategy (base values first, comparison values second). Default: 250. 

  - **`rill.metrics.exactify_druid_topn`** - _[boolean]_ - Split Druid TopN queries into two queries to improve measure accuracy, at the cost of performance. Default: false. 

  - **`rill.metrics.timeseries_null_filling_implementation`** - _[string]_ - Null-filling implementation for timeseries queries. One of `none`, `new`, or `pushdown`. Default: `pushdown`. 

  - **`rill.alerts.default_streaming_refresh_cron`** - _[string]_ - Default cron expression for refreshing alerts that depend on streaming refs (for example, external tables in Druid where new data may arrive at any time). Default: `0 0 * * *` (every 24 hours). 

  - **`rill.alerts.fast_streaming_refresh_cron`** - _[string]_ - Cron expression for refreshing streaming alerts on always-on OLAP connectors. Default: `*/10 * * * *` (every 10 minutes). 

  - **`rill.parser.skip_updates_if_parse_errors`** - _[boolean]_ - Short-circuit project parser reconciliation when parse errors exist. Default: false. 

  - **`rill.ai.completion_timeout_seconds`** - _[integer]_ - Maximum duration of a full AI completion request (which may include multiple LLM calls and tool uses), in seconds. Default: 300. 

  - **`rill.ai.llm_timeout_seconds`** - _[integer]_ - Maximum duration of a single LLM completion request, in seconds. Default: 180. Note: when using Rill's hosted AI service (i.e. not a self-configured LLM), the admin server enforces a hard upper bound of 10 minutes, so values above that have no effect. 

  - **`rill.ai.default_query_limit`** - _[integer]_ - Default row limit applied to AI tool queries when no limit is specified. Default: 25. 

  - **`rill.ai.max_query_limit`** - _[integer]_ - Maximum row limit allowed for AI tool queries. Default: 250. 

  - **`rill.ai.require_time_range`** - _[boolean]_ - Require AI tool queries to include a time range filter; reject queries without one. Default: true. 

  - **`rill.ai.max_time_range_days`** - _[integer]_ - Maximum time range allowed for AI tool queries, in days. Set to 0 for no limit. Default: 0. 

  - **`rill.strict_resolver_properties`** - _[boolean]_ - Return an error when a resolver contains properties not recognized by its implementation. Default: false. 

  - **`rill.strict_model_properties`** - _[boolean]_ - Return an error when a model contains unmapped properties. Default: false. 

```yaml
env:
    foo: bar
    rill.interactive_sql_row_limit: 5000
```

## Managing Paths in Rill

The public_paths and ignore_paths properties in the rill.yaml file provide control over which files and directories are processed or exposed by Rill. The public_paths property defines a list of file or directory paths to expose over HTTP. By default, it includes ['./public']. The ignore_paths property specifies a list of files or directories that Rill excludes during ingestion and parsing. This prevents unnecessary or incompatible content from affecting the project.
:::tip
Don't forget the leading `/` when specifying the path for `ignore_paths` and this path is also assuming the relative path from your project root.
:::


### `public_paths`

_[array of string]_ - List of file or directory paths to expose over HTTP. Defaults to ['./public'] 

### `ignore_paths`

_[array of string]_ - A list of file or directory paths to exclude from parsing. Useful for ignoring extraneous or non-Rill files in the project 

```yaml
ignore_paths:
    - /path/to/ignore
    - /file_to_ignore.yaml
```

## Testing access policies

During development, it is always a good idea to check if your [access policies](/developers/build/metrics-view/security) are behaving the way you designed them to before pushing these changes into production. You can set mock users which enables a drop down in the dashboard preview to view as a specific user. 
:::info The View as selector is not visible in my dashboard, why?
This feature is _only_ enabled when you have set a security policy on the dashboard. By default, the dashboard and it's contents is viewable by every user.
:::


### `mock_users`

_[array of object]_ - A list of mock users used to test dashboard security policies within the project 

  - **`email`** - _[string]_ - The email address of the mock user. This field is required _(required)_

  - **`name`** - _[string]_ - The name of the mock user. 

  - **`admin`** - _[boolean]_ - Indicates whether the mock user has administrative privileges 

  - **`groups`** - _[array of string]_ - An array of group names that the mock user is a member of 

```yaml
mock_users:
    - email: john@yourcompany.com
      name: John Doe
      admin: true
    - email: jane@partnercompany.com
      groups:
        - partners
    - email: anon@unknown.com
    - email: embed@rilldata.com
      name: embed
      custom_variable_1: Value_1
      custom_variable_2: Value_2
```

## Common Properties

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 