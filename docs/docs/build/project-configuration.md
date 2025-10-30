---
title: Project Configuration
sidebar_label: Project Configuration
sidebar_position: 70
---

## Introduction

You can set project-level configuration using a `rill.yaml` file, which should be in the root of your Rill project. Options set in this file apply to both local development and projects deployed to [Rill Cloud](/deploy/deploy-dashboard/existing-project).

:::info
Changes to your project's `rill.yaml` configuration file can be pushed to your Github repository and _will_ apply immediately in production. Please verify accordingly!
:::

## Olap connector

This option configures the OLAP engine used by the project. By default, Rill uses [DuckDB](https://duckdb.org). 

For deployed projects, we recommend setting `olap_connector` to `clickhouse` or `druid`, which enables access to Rill's [managed OLAP infrastructure](/deploy/deploy-dashboard/existing-project#managed-olap-infrastructure).

```yaml
olap_connector: duckdb
```

All the connectors that can be used as an OLAP connector are :
- duckdb (default in `rill start`)
- clickhouse (default in `rill deploy`)
- druid

## Variables

Use the `variables` setting in `rill.yaml` to hard-code values for project [variables](/build/credentials/variables).

```yaml
variables:
  database_url: "https://example.com"
  default_region: "us-east-1"
```

Variables can be dynamically referenced in other project files using templating syntax like `{{ .vars.database_url }}`.

For sensitive credentials, see [Credentials](/build/credentials).

## UI

:::info
Note this property is only applicable to projects deployed to Rill Cloud. For embedded projects, consult [Embedding Rill Dashboards](/integrate/embedding).
:::

This section is used to customize the appearance and behavior of Rill Cloud's user interface. 

### Public URLs

The `public_urls` property is used to define pages that can be accessed by unauthenticated users. You can configure this property to allow public access to specific dashboards (currently available for [embedded dashboards](/integrate/embedding#public-dashboards) only). 

```yaml
ui:
  public_urls:
    - /dashboard/dashboard_a
    - /dashboard/dashboard_b
```

In this example, a public URL to `dashboard_a` would be in the form: `https://<org_name>.rilldata.io/<project_name>/dashboard/dashboard_a`.

Note that the `public_urls` property accepts wildcards, so you can use a value of `/*` if you want to make all dashboards in your project publicly accessible:

```yaml
ui:
  public_urls:
    - /*
```

:::warning Restricted to embedded projects

Note that Rill Cloud projects with public URLs must have at least one report scheduled (even if the scheduled report isn't actively delivering). This is a mechanism intended to ensure public URLs are only active for paid projects.

Please contact your account executive (or reach out to [Rill Support](contact.md)) for more details.
:::

### Navigation

The `navigation` property is used to change the ordering of dashboards and canvases or introduce sections. The following configuration shows how to group menu items into sections:

```yaml
ui:
  navigation:
    - group: "Section A"
      items:
        - dashboard_a
        - explore_b
        - canvas_a
    - group: "Section B"
      items:
        - dashboard_c
```

You can also use navigation to order menu items without defining sections:

```yaml
ui:
  navigation:
    - dashboard_a
    - explore_b
    - canvas_a
    - dashboard_c
```

Note that dashboards/canvases not specified in the YAML file will appear at the bottom of the sidebar.

## Features

The `features` setting enables or disables features that are currently in development or restricted to specific user groups.

Available feature flags (all default to false):

- **`alertsInCloud`** – When enabled, this allows [alerts](/build/alerts) to be triggered in Rill Cloud (usually only available in `rill start`)
- **`chatCharts`** – Enables inline chart visualizations in AI chat responses
- **`cloudDataViewer`** – Enables the cloud data viewer, which allows viewing tables and models in Rill Cloud
- **`alertsEditInCloud`** – Allows editing of alerts in Rill Cloud (currently restricted to specific users)
- **`defaultEmbedTheme`** – For embedded dashboards, this sets the default theme to use

```yaml
features:
  - alertsInCloud
  - chatCharts
  - cloudDataViewer
```

