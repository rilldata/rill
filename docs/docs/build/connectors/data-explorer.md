---
title: Data Explorer
description: Explore your data in Rill Developer
sidebar_label: Data Explorer
sidebar_position: 19
---

Rill includes a built-in Data Explorer that allows you to browse and preview tables from your connected data sources directly within the developer UI. This is helpful for verifying schemas, checking data quality, and exploring available tables before creating [source models](/build/models/source-models).

## Accessing the Data Explorer

You can access the Data Explorer in the bottom-left panel of Rill Developer. This panel lists all configured [connectors](/build/connectors/data-source) in your environment, including [OLAP engines](/build/connectors/olap).

:::tip Default OLAP
Only one OLAP engine can be the default engine of a single Rill project. This is defined in the `rill.yaml` using the `olap_connector` key.

:::

<!-- ![Data Explorer](/img/build/connectors/data-explorer/overview.gif) -->

## Features

### Browse Tables and Schemas

Expanding a connector in the list reveals the available databases, schemas, and tables. Clicking on a table name will show its schema information, including column names and data types.

<!-- ![Data Explorer](/img/build/connectors/data-explorer/browse-table.png) -->


### Data Preview

Select a table to view a preview of its content. Rill displays a sample of rows (up to 150) so you can inspect the actual data values without needing to write a query or ingest the full dataset.

<!-- ![Data Explorer](/img/build/connectors/data-explorer/preview.png) -->


### Import Data (Rill Managed Only)

The Data Explorer is integrated into the data import workflow. When you connect to a new data source, Rill presents a simplified view of the explorer, allowing you to select the specific tables you want to ingest.

<!-- ![Data Explorer](/img/build/connectors/data-explorer/two-part-flow.gif) -->


:::tip Unsure of the table?

If you're in the import flow and not sure you have the correct table selected, you can always close the modal and use the full Data Explorer in the sidebar to inspect schemas and preview data before committing to an import.

:::

### Live Connector

When live connecting to your own OLAP database, Rill will not materialize any of your data into its embedded OLAP engine. Instead, all of the queries will be pushed down to your connected OLAP engine.

:::warning Read only

By default, when creating a live connector to your OLAP database, we set a `readonly` parameter in the `connector.yaml`. You can modify this setting to allow write access but is not recommended as this has the potential to drop data on your OLAP database.

:::

### Generate Metrics View or Dashboard with AI

Rill can automatically generate a metrics view and an exploratory dashboard from any table in the Data Explorer. By clicking the "Generate dashboard" button, Rill uses AI to analyze your table's schema and data to suggest relevant dimensions and measures.

For data source connectors using Rill Managed OLAP Engine, Rill will first create a [source model](/build/models/source-models) to ingest the data into the default OLAP engine, and then build the dashboard on top of that model.

## Supported Connectors

The Data Explorer supports browsing for most database and data warehouse connectors, including:

*   BigQuery
*   Snowflake
*   Athena
*   Redshift
*   PostgreSQL
*   MySQL
*   DuckDB  _(OLAP)_
*   ClickHouse _(OLAP)_
*   Motherduck _(OLAP)_
*   Pinot _(OLAP)_

For file-based sources (like S3 or GCS), exploration is limited to listing buckets and objects if the connector supports it.
