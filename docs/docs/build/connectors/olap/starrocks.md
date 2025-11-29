---
title: StarRocks
description: Power Rill dashboards using StarRocks
sidebar_label: StarRocks
sidebar_position: 07
---

[StarRocks](https://www.starrocks.io/) is an open-source, high-performance OLAP database designed for real-time analytics on large datasets. It provides a MySQL-compatible interface and excels in analytical workloads with its columnar storage format, vectorized query execution engine, and intelligent query optimization. StarRocks is particularly well-suited for real-time data analytics, ad-hoc queries, and interactive data exploration, making it a powerful choice for business intelligence, user behavior analytics, and data warehouse applications.

Rill supports connecting to an existing StarRocks cluster via a "live connector" and using it as an OLAP engine built against [external tables](/build/connectors/olap#external-olap-tables) to power Rill dashboards. This is particularly useful when working with extremely large datasets (hundreds of GBs or even TB+ in size).

## Configuring Rill Developer with StarRocks

When using Rill for local development, there are a few options to configure Rill to enable StarRocks as an OLAP engine:
1. Connect to an OLAP engine via Add Data. This will automatically create the `starrocks.yaml` file in your `connectors` directory and populate the `.env` file with `connector.starrocks.password` or `connector.starrocks.dsn` depending on which you select in the UI.

For more information on supported parameters, see our [StarRocks connector YAML reference docs](/reference/project-files/connectors#starrocks).

```yaml
type: connector

driver: starrocks
host: <HOSTNAME>
port: 9030
username: <USERNAME>
password: "{{ .env.connector.starrocks.password }}"
database: <DATABASE>
ssl: false

# or

dsn: "{{ .env.connector.starrocks.dsn }}"
```

2. You can manually set `connector.starrocks.dsn` in your project's `.env` file or try pulling existing credentials locally using `rill env pull` if the project has already been deployed to Rill Cloud.

:::tip Getting DSN errors in dashboards after setting `.env`?

If you are facing issues related to DSN connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.

:::

## Connection String (DSN)

Rill connects to StarRocks using the MySQL protocol. StarRocks supports two DSN formats:

### StarRocks URL Format (Recommended)

```bash
connector.starrocks.dsn="starrocks://user:password@host:port/database"
```

This format is more intuitive and will be automatically converted to the MySQL DSN format internally. For example:

```bash
connector.starrocks.dsn="starrocks://root:password@localhost:9030/analytics_db"
```

### MySQL DSN Format

```bash
connector.starrocks.dsn="user:password@tcp(host:port)/database?parseTime=true"
```

This format is also supported for compatibility. For example:

```bash
connector.starrocks.dsn="root:password@tcp(localhost:9030)/analytics_db?parseTime=true"
```

:::note Important Notes

- If `user` or `password` contain special characters, they should be URL encoded (i.e., `p@ssword` -> `p%40ssword`)
- StarRocks uses the MySQL wire protocol, so both DSN formats are compatible
- The `parseTime=true` parameter is automatically added when using the StarRocks URL format
- The DSN property takes precedence over individual connection fields (host, port, username, etc.)

:::

### Connection Parameters

- **host**: Hostname or IP address of the StarRocks FE (Frontend) node
- **port**: MySQL protocol port of the StarRocks FE node (default: 9030)
- **username**: Username for authentication (default: root)
- **password**: Password for authentication
- **database**: Name of the StarRocks database to connect to
- **ssl**: Enable SSL/TLS encryption for the connection (default: false)
- **log_queries**: Enable query logging for debugging purposes (default: false)

:::info Need help connecting to StarRocks?

If you would like to connect Rill to an existing StarRocks instance, please don't hesitate to [contact us](/contact). We'd love to help!

:::

## Setting the Default OLAP Connection

When connecting to StarRocks via the UI, the default OLAP connection will be automatically added to your rill.yaml. This will change the way the UI behaves, such as adding new data sources, as this is not supported with a StarRocks-backed Rill project.

```yaml
olap_connector: starrocks
```

:::note

For more information about available properties in `rill.yaml`, see our [project YAML](/reference/project-files/rill-yaml) documentation.

:::

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](/build/connectors/olap/multiple-olap) page.

:::

## Supported Versions

Rill supports connecting to StarRocks v2.5 or newer versions.

## Additional Notes

- StarRocks supports both materialized views and regular tables. Rill can query both types as external tables.
- For dashboards powered by StarRocks, [measure definitions](/build/metrics-view/#measures) are required to follow standard [StarRocks SQL](https://docs.starrocks.io/docs/sql-reference/sql-statements/account-management/CREATE%20USER/) syntax.
- StarRocks uses a MySQL-compatible protocol, making it easy to integrate with existing MySQL-based tools and workflows.
- The default MySQL protocol port for StarRocks FE nodes is 9030 (not to be confused with the HTTP port 8030).
