---
note: GENERATED. DO NOT EDIT.
title: Project YAML
sidebar_position: 41
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
Please see our reference documentation on [OLAP Engines](/build/connectors/olap).
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
2. [Environment](/build/models/templating) level properties (e.g. a specific property that have been set for `dev`)
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
# https://docs.rilldata.com/build/rill-project-file#dashboard-defaults
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

Primarily useful for [templating](/build/connectors/templating), variables can be set in the `rill.yaml` file directly. This allows variables to be set for your projects deployed to Rill Cloud while still being able to use different variable values locally if you prefer. 
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
Similar to how [connector credentials can be pushed / pulled](/build/connectors/credentials#pulling-credentials-and-variables-from-a-deployed-project-on-rill-cloud) from local to cloud or vice versa, project variables set locally in Rill Developer can be pushed to Rill Cloud and/or pulled back to your local instance from your deployed project by using the `rill env push` and `rill env pull` commands respectively.
:::


### `env`

_[object]_ - To define a variable in `rill.yaml`, pass in the appropriate key-value pair for the variable under the `env` key 

```yaml
env:
    numeric_var: 10
    string_var: "string_value"
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

During development, it is always a good idea to check if your [access policies](/build/metrics-view/security) are behaving the way you designed them to before pushing these changes into production. You can set mock users which enables a drop down in the dashboard preview to view as a specific user. 
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