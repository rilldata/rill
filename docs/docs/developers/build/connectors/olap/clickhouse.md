---
title: ClickHouse
description: Power Rill dashboards using ClickHouse
sidebar_label: ClickHouse
sidebar_position: 0
---

import LoomVideo from '@site/src/components/LoomVideo'; // Adjust the path as needed


[ClickHouse](https://clickhouse.com/docs/en/intro) is an open-source, column-oriented OLAP database management system known for its ability to perform real-time analytical queries on large-scale datasets. Its architecture is optimized for high performance, leveraging columnar storage and advanced compression techniques to speed up data reads and significantly reduce storage costs. ClickHouse's efficiency in query execution, scalability, and ability to handle even petabytes of data make it an excellent choice for real-time analytic use cases.

<LoomVideo loomId='b96143c386104576bcfe6cabe1038c38' /> <br />

Rill supports connecting to an existing ClickHouse cluster via a "live connector" and using it as an OLAP engine  built against [external tables](/developers/build/connectors/olap#external-olap-tables) to power Rill dashboards. This is particularly useful when working with extremely large datasets (hundreds of GBs or even TB+ in size).


:::note Supported Versions

Rill supports connecting to ClickHouse v22.7 or newer versions.

:::

## Connect to ClickHouse

When using ClickHouse for local development, you can connect via connection parameters or by using the DSN. Both local instances of ClickHouse and ClickHouse Cloud are supported.

After selecting "Add Data", select ClickHouse and fill in your connection parameters. This will automatically create the `clickhouse.yaml` file in your `connectors` directory and populate the `.env` file with `CLICKHOUSE_PASSWORD` or `CLICKHOUSE_DSN` depending on which you select in the UI.

For more information on supported parameters, see our [ClickHouse connector YAML reference docs](/reference/project-files/connectors#clickhouse).

```yaml
type: connector
driver: clickhouse

host: <HOSTNAME>
port: <PORT>
username: <USERNAME>
password: "{{ .env.CLICKHOUSE_PASSWORD }}"
ssl: true # required for ClickHouse Cloud
```

After creating the connector, you can edit the `.env` file manually in the project directory, or the connectors/clickhouse.yaml file.

:::tip Getting DSN errors in dashboards after setting `.env`?

If you are facing issues related to DSN connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.

:::

### Connection String (DSN)

Rill is able to connect to ClickHouse using the [ClickHouse Go Driver](https://clickhouse.com/docs/en/integrations/go). An appropriate connection string (DSN) will need to be set through the `CLICKHOUSE_DSN` property in Rill.

```bash
CLICKHOUSE_DSN="clickhouse://<hostname>:<port>?username=<username>&password=<password>"
```

Once the file is created, it will be added directly to the `.env` file in the project directory. To make changes to this connector, modify `CLICKHOUSE_DSN`.

```yaml
type: connector
driver: clickhouse

dsn: "{{ .env.CLICKHOUSE_DSN }}"
```

:::info Check your port

In most situations, the default port is 9440 for TLS and 9000 when not using TLS. However, it is worth double-checking the port that your ClickHouse instance is configured to use when setting up your connection string.

:::

:::note DSN properties

For more information about available DSN properties and setting an appropriate connection string, please refer to ClickHouse's [documentation](https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#dsn).

:::

## Connect to ClickHouse Cloud

If you are connecting to an existing [ClickHouse Cloud](https://clickhouse.com/cloud) instance, you can retrieve connection details about your instance by clicking on the `Connect` tab from within the admin settings navigation page. This will provide relevant information, such as the hostname, port, and username being used for your instance that you can then use to construct your DSN.

![ClickHouse Cloud](/img/build/connectors/olap-engines/clickhouse/clickhouse-cloud.png)

Using the information in the ClickHouse UI, populate the parameters of your connection. 

### Connection String (DSN)

Because ClickHouse Cloud requires a secure connection over [https](https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#http-support-experimental), you will need to pass in `secure=true` and `skip_verify=true` as additional URL parameters as part of your https URL (for your DSN).

```yaml
https://<hostname>:<port>?username=<username>&password=<password>&secure=true&skip_verify=true
```

:::info Need help connecting to ClickHouse?

If you would like to connect Rill to an existing ClickHouse instance, please don't hesitate to [contact us](/contact). We'd love to help!

:::

## Rill Managed ClickHouse

By setting `managed: true` in your ClickHouse connector, you will enable an embedded ClickHouse server to spin up with Rill. This will allow you to import data directly into this ClickHouse server without having to worry about managing an external database.

```yaml
type: connector

driver: clickhouse
managed: true
```

:::warning Managed ClickHouse is in Testing

Rill Managed ClickHouse is currently in testing. If you encounter any issues, please [contact us](/contact).

:::

Data ingestion is configured through [model YAML files](/reference/project-files/models) — there is no UI support for ingesting data into ClickHouse at this time. You write SQL that uses ClickHouse [table functions](https://clickhouse.com/docs/en/sql-reference/table-functions) to read from external sources, and Rill materializes the results into your ClickHouse instance.

For a list of supported data sources, see [Ingesting Data into ClickHouse](#ingesting-data-into-clickhouse) below.

## Connector Mode

By default, ClickHouse connectors operate in read-only mode. You can explicitly set the mode using the `mode` parameter:

```yaml
mode: read
```

:::note Read-Write Mode in Development

Read-write mode (`mode: readwrite`) for self-managed ClickHouse is currently in development. For now, data ingestion is supported through Rill Managed ClickHouse.

:::

## Advanced Configuration Options

### Optimize Temporary Tables Before Partition Replace

When using incremental models with partition overwrite strategies, you can enable automatic optimization of temporary tables before partition replacement operations. This can improve query performance by reducing the number of parts in each partition, but may increase processing time during model refreshes.

```yaml
optimize_temporary_tables_before_partition_replace: true # default: false
```

## Ingesting Data into ClickHouse

When using ClickHouse as your OLAP engine, you can ingest data from external sources by writing models that use ClickHouse's built-in [table functions](https://clickhouse.com/docs/en/sql-reference/table-functions). Each model is a YAML file with a SQL query that Rill executes against your ClickHouse connector.

There is no UI support for configuring ClickHouse data sources at this time — all configuration is done through model YAML files. Credentials are stored in your project's `.env` file and referenced using `{{ .env.VARIABLE_NAME }}` [template syntax](/developers/build/connectors/templating).

### Supported Data Sources

**Object Storage**
- [Amazon S3](/developers/build/connectors/data-source/clickhouse/s3) — `s3()` table function
- [Google Cloud Storage](/developers/build/connectors/data-source/clickhouse/gcs) — `s3()` with GCS HMAC keys
- [Azure Blob Storage](/developers/build/connectors/data-source/clickhouse/azure) — `azureBlobStorage()` table function
- [HTTPS](/developers/build/connectors/data-source/clickhouse/https) — `url()` table function

**Databases**
- [PostgreSQL](/developers/build/connectors/data-source/clickhouse/postgres) — `postgresql()` table function
- [MySQL](/developers/build/connectors/data-source/clickhouse/mysql) — `mysql()` table function
- [MongoDB](/developers/build/connectors/data-source/clickhouse/mongodb) — `mongodb()` table function
- [Supabase](/developers/build/connectors/data-source/clickhouse/supabase) — `postgresql()` table function
- [Remote ClickHouse](/developers/build/connectors/data-source/clickhouse/remote-clickhouse) — `remoteSecure()` / `remote()` table functions

**Table Formats**
- [Apache Iceberg](/developers/build/connectors/data-source/clickhouse/iceberg) — `icebergS3()` / `icebergAzure()` table functions
- [Delta Lake](/developers/build/connectors/data-source/clickhouse/delta-lake) — `deltaLake()` table function
- [Apache Hudi](/developers/build/connectors/data-source/clickhouse/hudi) — `hudi()` table function

**Other**
- [HDFS](/developers/build/connectors/data-source/clickhouse/hdfs) — `hdfs()` table function

### Example

Create `models/s3_events.yaml`:

```yaml
type: model
connector: my_clickhouse

sql: |
  SELECT *
  FROM s3(
    'https://my-bucket.s3.amazonaws.com/events/*.parquet',
    '{{ .env.AWS_ACCESS_KEY_ID }}',
    '{{ .env.AWS_SECRET_ACCESS_KEY }}',
    'Parquet'
  )
```

For dev/prod environment handling, see [Model Environment Templating](/developers/build/models/templating).

## Configuring Rill Cloud

When deploying a ClickHouse-backed project to Rill Cloud, you have the following options to pass the appropriate connection string to Rill Cloud:
1. If you have followed the UI to create your ClickHouse connector, the password or DSN should already exist in the .env file. During the deployment process, this `.env` file is automatically pushed with the deployment.
2. If `CLICKHOUSE_DSN` has already been set in your project `.env`, you can push and update these variables directly in your cloud deployment by using the `rill env push` command.


:::warning Local ClickHouse Server

If you are developing on a locally running ClickHouse server, this will not be deployed with your project. You will either need to use ClickHouse Cloud or Managed ClickHouse.

:::

## Setting the Default OLAP Connection

Creating a connection to ClickHouse will automatically add the `olap_connector` property in your project's [rill.yaml](/reference/project-files/rill-yaml) and change the default OLAP engine to ClickHouse.

```yaml
olap_connector: clickhouse
```

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](/developers/build/connectors/olap/multiple-olap) page.

:::

## Reading from Multiple Schemas

Rill supports reading from multiple schemas in ClickHouse from within the same project in Rill Developer, and all accessible tables (given the permission set of the underlying user) should automatically be listed in the lower left-hand tab, which can then be used to [create dashboards](/developers/build/dashboards).

## Additional Notes

- For dashboards powered by ClickHouse, [measure definitions](/developers/build/metrics-view/#measures) are required to follow standard [ClickHouse SQL](https://clickhouse.com/docs/en/sql-reference) syntax.
- Because string columns in ClickHouse can theoretically contain [arbitrary binary data](https://github.com/ClickHouse/ClickHouse/issues/2976#issuecomment-416694860), if your column contains invalid UTF-8 characters, you may want to first cast the column by applying the `toValidUTF8` function ([see ClickHouse documentation](https://clickhouse.com/docs/en/sql-reference/functions/string-functions#tovalidutf8)) before reading the table into Rill to avoid any downstream issues.
- Data ingestion into ClickHouse is configured through model YAML files — see [Ingesting Data into ClickHouse](#ingesting-data-into-clickhouse) for supported sources and examples.
