---
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

## JSON Schema

Here is a full JSON schema for the `rill.yaml` syntax:

```
{% json_schema_for_resource "rill.yaml" %}
```

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
