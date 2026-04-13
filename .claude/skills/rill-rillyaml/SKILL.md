---
name: rill-rillyaml
description: Detailed instructions and examples for developing the rill.yaml file
---

# Instructions for developing `rill.yaml`

## Introduction

`rill.yaml` is a required configuration file located at the root of every Rill project. It defines project-wide settings, similar to `package.json` in Node.js or `dbt_project.yml` in dbt.

## Core Concepts

### Project metadata

There are no required properties in `rill.yaml`, but it is common to configure:

- `display_name`: Human-readable name shown in the UI
- `description`: Brief description of the project's purpose
- `compiler`: Deprecated property that is commonly found in old projects

### Default OLAP connector

The `olap_connector` property sets the default OLAP database for the project. Models output to this connector by default, and metrics views query from it unless explicitly overridden.

Common values are `duckdb` or `clickhouse`. If not specified, Rill initializes a managed DuckDB database and uses it as the default OLAP connector. 

### Mock users for security testing

The `mock_users` property defines test users for validating security policies during local development. Each mock user can have:

- `email` (required): The user's email address
- `name`: Display name
- `admin`: Boolean indicating admin privileges
- `groups`: List of group memberships
- Custom attributes for use in security policy expressions

When mock users are defined and security policies exist, a "View as" dropdown appears in the dashboard preview.

### Environment variables

The `env` property sets default values for non-sensitive variables. These can be referenced in resource files using templating syntax (`{{ .env.<variable> }}`). Sensitive secrets should go in `.env` instead.

### Resource type defaults

Project-wide defaults can be set for resource types using plural keys:

- `models`: Default settings for all models (e.g., refresh schedules)
- `metrics_views`: Default settings for all metrics views (e.g., `first_day_of_week`)
- `explores`: Default settings for explore dashboards (e.g., `time_ranges`, `time_zones`)
- `canvases`: Default settings for canvas dashboards

Individual resources can override these defaults.

### Path management

- `ignore_paths`: List of paths to exclude from parsing (use leading `/`)
- `public_paths`: List of paths to expose over HTTP (defaults to `['./public']`)

### Environment overrides

The `dev` and `prod` properties allow environment-specific configuration overrides.

## Minimal Example

A minimal `rill.yaml` for a new project:

```yaml
display_name: My Analytics Project
```

## Complete Example

A comprehensive `rill.yaml` demonstrating common configurations:

```yaml
display_name: Sales Analytics
description: Sales performance dashboards with partner access controls

olap_connector: duckdb

# Non-sensitive environment variables
env:
  default_lookback: P30D
  data_bucket: gs://my-company-data

# Mock users for testing security policies locally
mock_users:
  - email: admin@mycompany.com
    name: Admin User
    admin: true
  - email: partner@external.com
    groups:
      - partners
  - email: viewer@mycompany.com
    tenant_id: xyz

# Project-wide defaults for models
models:
  refresh:
    cron: 0 0 * * *

# Project-wide defaults for metrics views
metrics_views:
  smallest_time_grain: day

# Project-wide defaults for explore dashboards
explores:
  defaults:
    time_range: P3M
  time_zones:
    - UTC
    - America/New_York
    - Europe/London
  time_ranges:
    - PT24H
    - P7D
    - P30D
    - P3M
    - P12M

# Exclude non-Rill files from parsing
ignore_paths:
  - /docs
```

## Reference documentation

Here is a full JSON schema for the `rill.yaml` syntax:

```
$schema: http://json-schema.org/draft-07/schema#
allOf:
    - properties:
        ai_connector:
            description: Specifies the default AI connector for the project. Defaults to Rill's internal AI connector if not set.
            type: string
        ai_instructions:
            description: Extra instructions for LLM/AI features. Used to guide natural language question answering and routing.
            type: string
        compiler:
            description: Specifies the parser version to use for compiling resources
            type: string
        description:
            description: A brief description of the project
            type: string
        display_name:
            description: The display name of the project, shown in the upper-left corner of the UI
            type: string
        features:
            description: Optional feature flags. Can be specified as a map of feature names to booleans.
            type: object
      title: Properties
      type: object
    - description: |
        Rill allows you to specify the default OLAP engine to use in your project via `rill.yaml`.
        :::info Curious about OLAP Engines?
        Please see our reference documentation on [OLAP Engines](/developers/build/connectors/olap).
        :::
      properties:
        olap_connector:
            description: Specifies the default OLAP engine for the project. Defaults to duckdb if not set.
            examples:
                - olap_connector: clickhouse
            type: string
      title: Configuring the default OLAP Engine
      type: object
    - description: |
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
      properties:
        canvases:
            description: Defines project-wide default settings for canvases. Unless overridden, individual canvases will inherit these defaults.
            examples:
                - canvases:
                    defaults:
                        time_range: P7D
                    time_ranges:
                        - PT24H
                        - P7D
                    time_zones:
                        - UTC
                  explores:
                    defaults:
                        time_range: P24M
                    time_ranges:
                        - PT24H
                        - P6M
                    time_zones:
                        - UTC
                  metrics_views:
                    first_day_of_week: 1
                    smallest_time_grain: month
                  models:
                    refresh:
                        cron: 0 * * * *
            type: object
        explores:
            description: Defines project-wide default settings for explores. Unless overridden, individual explores will inherit these defaults.
            type: object
        metrics_views:
            description: Defines project-wide default settings for metrics_views. Unless overridden, individual metrics_views will inherit these defaults.
            type: object
        models:
            description: Defines project-wide default settings for models. Unless overridden, individual models will inherit these defaults.
            type: object
      title: Project-wide defaults
      type: object
    - description: "Primarily useful for [templating](/developers/build/connectors/templating), variables can be set in the `rill.yaml` file directly. This allows variables to be set for your projects deployed to Rill Cloud while still being able to use different variable values locally if you prefer. \n:::info Overriding variables locally\nVariables also follow an order of precedence and can be overridden locally. By default, any variables defined will be inherited from `rill.yaml`. However, if you manually pass in a variable when starting Rill Developer locally via the CLI, this value will be used instead for the current instance of your running project:\n```bash\nrill start --env numeric_var=100 --env string_var=\"different_value\"\n```\n:::\n:::tip Setting variables through `.env`\nVariables can also be set through your project's `<RILL_PROJECT_HOME>/.env` file (or using the `rill env set` CLI command), such as:\n```bash\nvariable=xyz\n```\nSimilar to how [connector credentials can be pushed / pulled](/developers/build/connectors/credentials#pulling-credentials-and-variables-from-a-deployed-project-on-rill-cloud) from local to cloud or vice versa, project variables set locally in Rill Developer can be pushed to Rill Cloud and/or pulled back to your local instance from your deployed project by using the `rill env push` and `rill env pull` commands respectively.\n:::\n"
      properties:
        env:
            description: To define a variable in `rill.yaml`, pass in the appropriate key-value pair for the variable under the `env` key
            examples:
                - env:
                    numeric_var: 10
                    string_var: string_value
            type: object
      title: Setting variables
      type: object
    - description: |
        The public_paths and ignore_paths properties in the rill.yaml file provide control over which files and directories are processed or exposed by Rill. The public_paths property defines a list of file or directory paths to expose over HTTP. By default, it includes ['./public']. The ignore_paths property specifies a list of files or directories that Rill excludes during ingestion and parsing. This prevents unnecessary or incompatible content from affecting the project.
        :::tip
        Don't forget the leading `/` when specifying the path for `ignore_paths` and this path is also assuming the relative path from your project root.
        :::
      properties:
        ignore_paths:
            description: A list of file or directory paths to exclude from parsing. Useful for ignoring extraneous or non-Rill files in the project
            examples:
                - ignore_paths:
                    - /path/to/ignore
                    - /file_to_ignore.yaml
            items:
                type: string
            type: array
        public_paths:
            description: List of file or directory paths to expose over HTTP. Defaults to ['./public']
            items:
                type: string
            type: array
      title: Managing Paths in Rill
      type: object
    - description: "During development, it is always a good idea to check if your [access policies](/developers/build/metrics-view/security) are behaving the way you designed them to before pushing these changes into production. You can set mock users which enables a drop down in the dashboard preview to view as a specific user. \n:::info The View as selector is not visible in my dashboard, why?\nThis feature is _only_ enabled when you have set a security policy on the dashboard. By default, the dashboard and it's contents is viewable by every user.\n:::\n"
      properties:
        mock_users:
            description: A list of mock users used to test dashboard security policies within the project
            examples:
                - mock_users:
                    - admin: true
                      email: john@yourcompany.com
                      name: John Doe
                    - email: jane@partnercompany.com
                      groups:
                        - partners
                    - email: anon@unknown.com
                    - custom_variable_1: Value_1
                      custom_variable_2: Value_2
                      email: embed@rilldata.com
                      name: embed
            items:
                properties:
                    admin:
                        description: Indicates whether the mock user has administrative privileges
                        type: boolean
                    email:
                        description: The email address of the mock user. This field is required
                        type: string
                    groups:
                        description: An array of group names that the mock user is a member of
                        items:
                            type: string
                        type: array
                    name:
                        description: The name of the mock user.
                        type: string
                required:
                    - email
                type: object
            type: array
      title: Testing access policies
      type: object
    - properties:
        dev:
            description: Overrides any properties in development environment.
            type: object
        prod:
            description: Overrides any properties in production environment.
            type: object
      title: Common Properties
      type: object
description: The `rill.yaml` file contains metadata about your project.
id: rill-yaml.schema.yaml
title: Project YAML
type: object
```