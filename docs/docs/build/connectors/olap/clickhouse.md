---
title: ClickHouse
description: Power Rill dashboards using ClickHouse
sidebar_label: ClickHouse
sidebar_position: 0
---

import LoomVideo from '@site/src/components/LoomVideo'; // Adjust the path as needed


[ClickHouse](https://clickhouse.com/docs/en/intro) is an open-source, column-oriented OLAP database management system known for its ability to perform real-time analytical queries on large-scale datasets. Its architecture is optimized for high performance, leveraging columnar storage and advanced compression techniques to speed up data reads and significantly reduce storage costs. ClickHouse's efficiency in query execution, scalability, and ability to handle even petabytes of data make it an excellent choice for real-time analytic use cases.

<LoomVideo loomId='b96143c386104576bcfe6cabe1038c38' /> <br />

Rill supports connecting to an existing ClickHouse cluster via a "live connector" and using it as an OLAP engine  built against [external tables](/build/connectors/olap#external-olap-tables) to power Rill dashboards. This is particularly useful when working with extremely large datasets (hundreds of GBs or even TB+ in size).


:::note Supported Versions

Rill supports connecting to ClickHouse v22.7 or newer versions.

:::

## Connect to ClickHouse

When using ClickHouse for local development, you can connect via connection parameters or by using the DSN. Both local instances of ClickHouse and ClickHouse Cloud are supported.

After selecting "Add Data", select ClickHouse and fill in your connection parameters. This will automatically create the `clickhouse.yaml` file in your `connectors` directory and populate the `.env` file with `connector.clickhouse.password` or `connector.clickhouse.dsn` depending on which you select in the UI.

For more information on supported parameters, see our [ClickHouse connector YAML reference docs](/reference/project-files/connectors#clickhouse).

```yaml
type: connector
driver: clickhouse

host: <HOSTNAME>
port: <PORT>
username: <USERNAME>
password: "{{ .env.connector.clickhouse.password }}"
ssl: true # required for ClickHouse Cloud
```

After creating the connector, you can edit the `.env` file manually in the project directory, or the connectors/clickhouse.yaml file.

:::tip Getting DSN errors in dashboards after setting `.env`?

If you are facing issues related to DSN connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.

:::

### Connection String (DSN)

Rill is able to connect to ClickHouse using the [ClickHouse Go Driver](https://clickhouse.com/docs/en/integrations/go). An appropriate connection string (DSN) will need to be set through the `connector.clickhouse.dsn` property in Rill.

```bash
connector.clickhouse.dsn="clickhouse://<hostname>:<port>?username=<username>&password=<password>"
```

Once the file is created, it will be added directly to the `.env` file in the project directory. To make changes to this connector, modify `connector.clickhouse.dsn`.

```yaml
type: connector
driver: clickhouse

dsn: "{{ .env.connector.clickhouse.dsn }}"
```

:::info Check your port

In most situations, the default port is 9440 for TLS and 9000 when not using TLS. However, it is worth double-checking the port that your ClickHouse instance is configured to use when setting up your connection string.

:::

:::note DSN properties

For more information about available DSN properties and setting an appropriate connection string, please refer to ClickHouse's [documentation](https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#dsn).

:::

## Connect to ClickHouse Cloud

If you are connecting to an existing [ClickHouse Cloud](https://clickhouse.com/cloud) instance, you can retrieve connection details about your instance by clicking on the `Connect` tab from within the admin settings navigation page. This will provide relevant information, such as the hostname, port, and username being used for your instance that you can then use to construct your DSN.

<img src='/img/build/connectors/olap-engines/clickhouse/clickhouse-cloud.png' class='rounded-gif' />
<br />

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

Data ingestion features are not yet available in the UI, please refer to our [model documentation](/reference/project-files/models) on how to ingest data into ClickHouse. For a guide, see [ingesting data directly into ClickHouse](/guides/rill-clickhouse/r_ch_ingest).


## Read Only Connector

You can configure your ClickHouse connector to operate in read-only mode by adding a mode parameter. This provides an additional security layer by ensuring your connector can only read data and cannot perform write operations.

```yaml
mode: read # readwrite
```

## Advanced Configuration Options

### Optimize Temporary Tables Before Partition Replace

When using incremental models with partition overwrite strategies, you can enable automatic optimization of temporary tables before partition replacement operations. This can improve query performance by reducing the number of parts in each partition, but may increase processing time during model refreshes.

```yaml
optimize_temporary_tables_before_partition_replace: true # default: false
```

## Configuring Rill Cloud

When deploying a ClickHouse-backed project to Rill Cloud, you have the following options to pass the appropriate connection string to Rill Cloud:
1. If you have followed the UI to create your ClickHouse connector, the password or DSN should already exist in the .env file. During the deployment process, this `.env` file is automatically pushed with the deployment.
2. If `connector.clickhouse.dsn` has already been set in your project `.env`, you can push and update these variables directly in your cloud deployment by using the `rill env push` command.
3. If you manually passed the connector when running `rill start`, you will need to use the `rill env configure` command to set `connector.clickhouse.dsn` onto Rill Cloud as well.

:::warning Local ClickHouse Server

If you are developing on a locally running ClickHouse server, this will not be deployed with your project. You will either need to use ClickHouse Cloud or Managed ClickHouse.

:::

## Setting the Default OLAP Connection

Creating a connection to an OLAP engine will automatically add the `olap_connector` property in your project's [rill.yaml](/reference/project-files/rill-yaml) and change the default OLAP engine to ClickHouse. Once this is changed, you'll notice that some of the UI features are removed as we currently do not support modeling and direct source ingestion in ClickHouse. However, this is behind a feature flag, [contact us](/contact) for more information!

```yaml
olap_connector: clickhouse
```

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](/build/connectors/olap/multiple-olap) page.

:::

## Reading from Multiple Schemas

Rill supports reading from multiple schemas in ClickHouse from within the same project in Rill Developer, and all accessible tables (given the permission set of the underlying user) should automatically be listed in the lower left-hand tab, which can then be used to [create dashboards](/build/dashboards).

## Additional Notes

- At the moment, we do not officially support modeling with ClickHouse; however, this is available via a feature flag. If this is something you're interested in, please [contact us](/contact).
- For dashboards powered by ClickHouse, [measure definitions](/build/metrics-view/#measures) are required to follow standard [ClickHouse SQL](https://clickhouse.com/docs/en/sql-reference) syntax.
- Because string columns in ClickHouse can theoretically contain [arbitrary binary data](https://github.com/ClickHouse/ClickHouse/issues/2976#issuecomment-416694860), if your column contains invalid UTF-8 characters, you may want to first cast the column by applying the `toValidUTF8` function ([see ClickHouse documentation](https://clickhouse.com/docs/en/sql-reference/functions/string-functions#tovalidutf8)) before reading the table into Rill to avoid any downstream issues.
