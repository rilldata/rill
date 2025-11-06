---
title: "Your First Rill Project"
description: "Get started with Rill by understanding what happens when you create a project"
sidebar_label: "Getting Started"
sidebar_position: 1
---

# Your First Rill Project

When you create a new Rill project, you're setting up a complete data pipeline that transforms raw data into interactive dashboards. This guide explains what happens during project initialization and introduces you to the key files you'll work with.

## What Happens When You Create a Rill Project?

A Rill project consists of several data assets that work together to create a dashboard. The data pipeline begins with [connecting](/build/connectors) to your data sources, transforms raw data through [models](/build/models), defines metrics and dimensions in [metrics views](/build/metrics-view), and results in interactive [dashboards](/build/dashboards) for data analysis.

When you create a new Rill project, the following files are automatically generated:

- **`rill.yaml`** - Central configuration hub for your entire project
- **`connectors/<connector>.yaml`** - Connector configuration files for the project's default OLAP engine (e.g., `duckdb.yaml`, `clickhouse.yaml`)
- **`.gitignore`** - Git ignore rules for the project

## Initial Project Files

These are the foundational files that Rill creates for you. As you build your project, you'll add many more files including sources, models, metrics views, and dashboards.

### `rill.yaml`
The central configuration file that controls project-wide settings. You rarely need to modify `rill.yaml` when starting out - the defaults work great! This file enables you to set [project-wide defaults](/build/project-configuration#model-defaults), [configure variables](/build/project-configuration#variable-management), define [connector settings](/build/project-configuration#olap-connector), create [test users](/build/project-configuration#testing-security), and establish [security policies](/build/project-configuration#metrics-views-security-policy).

### `connectors/<connector>.yaml`
Configuration for your default OLAP engine (DuckDB, ClickHouse, Druid, or Pinot). When starting a blank project, this always defaults to `duckdb.yaml`. This file defines how Rill connects to your analytical database.

Connectors enable Rill to connect to various data sources and OLAP engines. You can configure [data source connectors](/build/connectors/data-source) (like S3, GCS, BigQuery, Snowflake) to ingest data, and [OLAP connectors](/build/connectors/olap) to power your analytics. See our [connectors documentation](/build/connectors) for the full list of supported connections.


### `.gitignore`
Specifies which files and directories should be ignored by Git version control. Rill automatically generates this file with appropriate rules to exclude sensitive files like `.env`, temporary files, and build artifacts from being committed to your repository.

## Next Steps

Now that you understand the basics, you can:
1. **[Connect to your data](/build/connectors)** - Set up connections to your data sources
1. **[Create your first model](/build/models)** - Transform your raw data
2. **[Build a metrics view](/build/metrics-view)** - Define your metrics and dimensions
3. **[Create a dashboard](/build/dashboards)** - Visualize your data
4. **[Deploy to Rill Cloud](/deploy)** - Deploy your project to Rill Cloud

When you need more control over your project configuration, see our [project configuration guide](/build/project-configuration) for advanced settings and customization options.
