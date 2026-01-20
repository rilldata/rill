---
title: StarRocks
description: Power Rill dashboards using StarRocks
sidebar_label: StarRocks
sidebar_position: 25
---

[StarRocks](https://www.starrocks.io/) is an open-source, high-performance analytical database designed for real-time, multi-dimensional analytics on large-scale data. It supports both primary key and aggregate data models, making it suitable for a variety of analytical workloads including real-time dashboards, ad-hoc queries, and complex analytical tasks.

:::note Supported Versions

Rill supports connecting to StarRocks 4.0 or newer versions.

:::

:::info

Rill supports connecting to an existing StarRocks cluster via a read-only OLAP connector and using it to power Rill dashboards with [external tables](/developers/build/connectors/olap#external-olap-tables).

:::

## Connect to StarRocks

When using StarRocks for local development, you can connect via connection parameters or by using a DSN.

After selecting "Add Data", select StarRocks and fill in your connection parameters. This will automatically create the `starrocks.yaml` file in your `connectors` directory and populate the `.env` file with `connector.starrocks.password`.

### Connection Parameters

```yaml
type: connector
driver: starrocks

host: <HOSTNAME>
port: 9030
username: <USERNAME>
password: "{{ .env.connector.starrocks.password }}"
catalog: default_catalog
database: <DATABASE>
ssl: false
```

### Connection String (DSN)

Rill can also connect to StarRocks using a DSN connection string. StarRocks uses MySQL protocol, so the connection string must follow the MySQL DSN format:

```yaml
type: connector
driver: starrocks

dsn: "{{ .env.connector.starrocks.dsn }}"
```

#### Using default_catalog

For `default_catalog`, you can specify database directly in the DSN path (MySQL-style):
```
user:password@tcp(host:9030)/my_database?parseTime=true
```

#### Using external catalogs with DSN

For external catalogs (Iceberg, Hive, etc.), set `catalog` and `database` as separate properties (do not include database in DSN):
```yaml
type: connector
driver: starrocks

dsn: "user:password@tcp(host:9030)/?parseTime=true"
catalog: iceberg_catalog
database: my_database
```

If `catalog` is not specified, it defaults to `default_catalog`.

:::warning DSN Format

Only MySQL-style DSN format is supported. The `starrocks://` URL scheme is **not** supported. When using DSN, do not set `host`, `port`, `username`, `password` separately — these must be included in the DSN string.

:::

## Configuration Properties

| Property      | Description                                                           | Default              |
| ------------- | --------------------------------------------------------------------- | -------------------- |
| `host`        | StarRocks FE (Frontend) server hostname                               | Required (if no DSN) |
| `port`        | MySQL protocol port of StarRocks FE                                   | `9030`               |
| `username`    | Username for authentication                                           | `root`               |
| `password`    | Password for authentication                                           | -                    |
| `catalog`     | StarRocks catalog name (for external catalogs like Iceberg, Hive)     | `default_catalog`    |
| `database`    | StarRocks database name                                               | -                    |
| `ssl`         | Enable SSL/TLS encryption                                             | `false`              |
| `dsn`         | MySQL-format connection string (alternative to individual parameters) | -                    |
| `log_queries` | Enable logging of all SQL queries (useful for debugging)              | `false`              |

## External Catalogs

StarRocks supports external catalogs for querying data in Hive, Iceberg, Delta Lake, and other external data sources. To use an external catalog:

1. Set the `catalog` property to your external catalog name (e.g., `iceberg_catalog`)
2. Set the `database` property to the database within that catalog

```yaml
type: connector
driver: starrocks

host: starrocks-fe.example.com
port: 9030
username: analyst
password: "{{ .env.connector.starrocks.password }}"
catalog: iceberg_catalog
database: my_database
```

## Naming Mapping

StarRocks uses a three-level hierarchy: Catalog > Database > Table. In Rill's API:

| Rill Parameter   | StarRocks Concept | Example                              |
| ---------------- | ----------------- | ------------------------------------ |
| `database`       | Catalog           | `default_catalog`, `iceberg_catalog` |
| `databaseSchema` | Database          | `my_database`                        |
| `table`          | Table             | `my_table`                           |

## Creating Metrics Views

When creating metrics views against StarRocks tables, use the `table` property with `database_schema` to reference your data:

```yaml
type: metrics_view
display_name: My Dashboard
table: my_table
database_schema: my_database
timeseries: timestamp

dimensions:
  - name: category
    column: category

measures:
  - name: total_count
    expression: COUNT(*)
```

## Troubleshooting

### Connection Issues

If you encounter connection issues:

1. Verify the FE node hostname and port (default: 9030)
2. Check that your user has appropriate permissions
3. Ensure network connectivity to the StarRocks FE node
4. For SSL connections, verify SSL is enabled on the StarRocks server

### Timezone Handling

All timestamp values are returned in UTC. The driver parses DATETIME values from StarRocks as UTC time.

## Known Limitations

- **Read-only connector**: StarRocks is a read-only OLAP connector. Model creation and execution is not supported.

:::info Need help connecting to StarRocks?

If you would like to connect Rill to an existing StarRocks instance, please don't hesitate to [contact us](/contact). We'd love to help!

:::
