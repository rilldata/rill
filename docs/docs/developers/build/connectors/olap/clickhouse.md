---
title: ClickHouse
description: Power Rill dashboards using ClickHouse
sidebar_label: ClickHouse
sidebar_position: 0
---

import LoomVideo from '@site/src/components/LoomVideo'; // Adjust the path as needed


[ClickHouse](https://clickhouse.com/docs/en/intro) is an open-source, column-oriented OLAP database management system known for its ability to perform real-time analytical queries on large-scale datasets. Its architecture is optimized for high performance, leveraging columnar storage and advanced compression techniques to speed up data reads and significantly reduce storage costs. ClickHouse's efficiency in query execution, scalability, and ability to handle even petabytes of data make it an excellent choice for real-time analytic use cases.

<LoomVideo loomId='b96143c386104576bcfe6cabe1038c38' /> <br />

Rill supports ClickHouse in three ways:

- **ClickHouse Cloud** — Connect to a managed [ClickHouse Cloud](https://clickhouse.com/cloud) instance.
- **Rill Managed ClickHouse** — Rill provisions and manages a ClickHouse instance for you. No infrastructure to set up.
- **Self-Managed ClickHouse** — Connect to your own self-hosted ClickHouse instance. Rill queries your cluster directly via a live connector with no data ingestion.

:::note Supported Versions

Rill supports connecting to ClickHouse v22.7 or newer versions.

:::

## ClickHouse Cloud

Connect to an existing [ClickHouse Cloud](https://clickhouse.com/cloud) instance. You can retrieve connection details by clicking the `Connect` tab from within the admin settings navigation page. This will provide the hostname, port, and username for your instance.

![ClickHouse Cloud](/img/build/connectors/olap-engines/clickhouse/clickhouse-cloud.png)

After selecting "Add Data", select ClickHouse and fill in your connection parameters. This will automatically create the `clickhouse.yaml` file in your `connectors` directory and populate the `.env` file with `CLICKHOUSE_PASSWORD` or `CLICKHOUSE_DSN` depending on which you select in the UI.

For more information on supported parameters, see our [ClickHouse connector YAML reference docs](/reference/project-files/connectors#clickhouse).

```yaml
type: connector
driver: clickhouse

host: <HOSTNAME>
port: <PORT>
username: <USERNAME>
password: "{{ .env.CLICKHOUSE_PASSWORD }}"
ssl: true
```

### Connection String (DSN)

Because ClickHouse Cloud requires a secure connection over [https](https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#http-support-experimental), you will need to pass in `secure=true` and `skip_verify=true` as additional URL parameters as part of your https URL (for your DSN).

```yaml
https://<hostname>:<port>?username=<username>&password=<password>&secure=true&skip_verify=true
```

:::info Need help connecting to ClickHouse?

If you would like to connect Rill to an existing ClickHouse instance, please don't hesitate to [contact us](/contact). We'd love to help!

:::

## Rill Managed ClickHouse

By setting `managed: true` in your ClickHouse connector, Rill will spin up an embedded ClickHouse server. This allows you to import data directly without managing an external database.

```yaml
type: connector

driver: clickhouse
managed: true
```

:::warning Managed ClickHouse is in Testing

Rill Managed ClickHouse is currently in testing. If you encounter any issues, please [contact us](/contact).

:::

:::tip Ingesting Data
For a full list of supported data sources and configuration examples, see the [ClickHouse data sources](/developers/build/connectors/data-source#clickhouse) documentation.
:::

## Self-Managed ClickHouse

Connect to your own self-hosted ClickHouse instance using connection parameters or a DSN. Rill uses ClickHouse as an OLAP engine built against [external tables](/developers/build/connectors/olap#external-olap-tables) to power dashboards. This is particularly useful when working with extremely large datasets (hundreds of GBs or even TB+ in size).

After selecting "Add Data", select ClickHouse and fill in your connection parameters. This will automatically create the `clickhouse.yaml` file in your `connectors` directory and populate the `.env` file with `CLICKHOUSE_PASSWORD` or `CLICKHOUSE_DSN` depending on which you select in the UI.

For more information on supported parameters, see our [ClickHouse connector YAML reference docs](/reference/project-files/connectors#clickhouse).

```yaml
type: connector
driver: clickhouse

host: <HOSTNAME>
port: <PORT>
username: <USERNAME>
password: "{{ .env.CLICKHOUSE_PASSWORD }}"
ssl: false
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

For a full list of supported data sources and configuration examples, see the [ClickHouse data sources](/developers/build/connectors/data-source#clickhouse) documentation.

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
- Data ingestion into ClickHouse is configured through model YAML files — see the [ClickHouse data sources](/developers/build/connectors/data-source#clickhouse) documentation for supported sources and examples.
