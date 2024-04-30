---
title: "Connect Sources"
description: Import local files or remote data sources
sidebar_label: "Connect Sources"
sidebar_position: 00
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->


Rill supports a multitude of connectors to ingest data from various sources: local files, S3 or GCS buckets, download using HTTP(S), databases, data warehouses, and the list goes on. Rill can ingest `.csv`, `.tsv`, `.json`,and `.parquet` files, which may be compressed (`.gz`). This can be done either through the UI directly, when working with Rill Developer, or by pushing the logic into the [source YAML](../../reference/project-files/sources.md) definition directly (see _Using Code_ sections below).

To provide a non-exhaustive list, Rill supports the following connectors:
- [Google Cloud Storage](/reference/connectors/gcs.md)
- [S3](/reference/connectors/s3.md)
- [Azure Blob Storage](/reference/connectors/azure.md)
- [BigQuery](/reference/connectors/bigquery.md)
- [Athena](/reference/connectors/athena.md)
- [Redshift](/reference/connectors/redshift.md)
- [Kafka](/reference/connectors/kafka.md)
- [DuckDB and MotherDuck](/reference/connectors/motherduck.md)
- [PostgreSQL](/reference/connectors/postgres.md)
- [MySQL](/reference/connectors/mysql.md)
- [SQLite](/reference/connectors/sqlite.md)
- [Snowflake](/reference/connectors/snowflake.md)
- [Salesforce](/reference/connectors/salesforce.md)
- [Google Sheets](/reference/connectors/googlesheets.md)

:::info Full List Of Connectors

Rill is continually adding new sources and connectors in our releases. For a comprehensive list, you can refer to our [Connectors](/reference/connectors) page. Please don't hesitate to [reach out](contact.md) either if there's a connector you'd like us to add!

:::

:::tip Avoid Pre-aggregated Metrics

Rill works best for slicing and dicing data meaning keeping data closer to raw to retain that granularity for flexible analysis. When loading data - be careful with adding pre-aggregated metrics like averages as that could lead to unintended results like a sum of an average. Instead load the two raw metrics and calculate the derived metric in your model or dashboard.

:::

## Adding a local file

### Using the UI

To import a file using the UI, click "+" by Sources in the left hand navigation pane, select "Local File", and navigate to the specific file. Alternately, try dragging and dropping the file directly onto the Rill interface.

### Using code
When you add a source using the UI, a code definition will automatically be created as a `.yaml` file in your Rill project in the `sources` directory. However, you can also create sources more directly by creating the artifact.

In your Rill project directory, create a `source_name.yaml` file in the `sources` directory with the following content:

```yaml
type: source
connector: local_file
path: /path/to/local/data.csv
```

Rill will ingest the data next time you run `rill start`.

Note that if you provide a relative path, _the path should be relative to your Rill project root_ (where your `rill.yaml` file is located), **not** relative to the `sources` directory.

:::tip Import from multiple files
To import data from multiple files, you can use a glob pattern to specify the files you want to include. To learn more about the syntax and details of glob patterns, please refer to the documentation on [glob patterns](../connect/glob-patterns.md).
:::

:::note Source Properties

For more details about available configurations and properties, check our [Source YAML](../../reference/project-files/sources) reference page.

:::

## Adding a remote source

### Using the UI
To add a remote source using the UI, click "+" by Sources in the left hand navigation pane and select the location where your remote files are stored ("Google Cloud Storage", "Amazon S3", or "http(s)"). Enter your file's URI and click "Add Source".

After import, you can reimport your data whenever you want by clicking the "refresh source" button in the Rill UI.

### Using code
When you add a source using the UI or CLI, a code definition will automatically be created as a `.yaml` file in your Rill project in the `sources` directory.

For example, to create a remote http(s) source, create a `source_name.yaml` file in the `sources` directory with the following contents:

```yaml
type: source
connector: https
uri: https://data.example.org/path/to/file.parquet
```

:::info

For a full list of connector types available, please see our [Connectors](/reference/connectors/connectors.md) and [Source YAML](/reference/project-files/sources.md#properties) reference pages.

:::

You can also push filters to your source definition using inline editing. Common use cases for inline editing:

- Filter data to only use part of source files
- Push transformations for key fields to source (particularly casting time fields, data types)
- Resolve ingestion issues by declaring types (examples: STRUCT with different values to VARCHAR, fields mixed with INT and VARCHAR values)

:::tip Import from multiple files
To import data from multiple files, you can use a glob pattern to specify the files you want to include. To learn more about the syntax and details of glob patterns, please refer to the documentation on [glob patterns](glob-patterns.md).
:::

:::note Source Properties

For more details about available configurations and properties, check our [Source YAML](../../reference/project-files/sources) reference page.

:::

## Authenticating remote sources

Rill requires an appropriate set of <u>credentials</u> to connect to remote data sources, whether those are buckets (e.g. S3 or GCS) or data warehouses (e.g. Snowflake). When running Rill locally, Rill Developer attempts to find existing credentials that have been configured on your machine. When deploying projects to Rill Cloud, you must explicitly provide service account credentials with correct access permissions.

:::note Setting up credentials

Please see our [Configuring Credentials](../credentials/credentials.md) and [Deployment Credentials](../../deploy/deploy-credentials.md) for more information about setting up and using credentials in Rill.

:::

## External OLAP tables

Rill also has the ability to set up a "live connection" with an [OLAP engine](../olap/olap.md) to discover existing tables and execute OLAP queries directly on the engine without having to transfer data to another OLAP engine. By default, the embedded OLAP engine that comes with Rill is [DuckDB](/reference/olap-engines/duckdb.md).

:::note Configuring the OLAP engine

For more details about configuring and/or changing the OLAP engine used by Rill, please see our [OLAP Engines](/reference/olap-engines/olap-engines.md) reference documentation.

:::

## Rill Developer vs Rill Cloud

_There is a difference_ between Rill Developer and Rill Cloud and they work hand-in-hand to provide a shared experience. For distributed teams, Rill Developer is primarily meant for local development and modeling purposes while Rill Cloud is where the primary dashboard consumption occurs and helps to enable shared collaboration at scale. For Rill Developer, as the size or volume of source data continues to grow (or reaches a certain size), it is strongly recommended to [work with a segment of the data for modeling purposes](../../deploy/performance.md#work-with-a-subset-of-your-source-data-for-local-development-and-modeling) instead of the full dataset (i.e. think of it as a "dev partition"), which is meant to help the developer validate the model logic and verify that the correct results are being produced. Then, after the [model](../models/models.md) and [dashboard](../dashboards/dashboards.md) configurations have been finalized, the project can be [deployed to Rill Cloud](../../deploy/existing-project/existing-project.md) against the full range of data and dashboards can be [explored](../../explore/dashboard-101.md) by other end users.

:::info Have questions?

We are one Slack, email, or chat message away. Please feel free to [contact us](contact.md) - we'd love to help!

:::
