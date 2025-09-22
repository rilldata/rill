---
title: DuckDB
description: Power Rill dashboards using DuckDB (default)
sidebar_label: DuckDB
sidebar_position: 10
---

[DuckDB](https://duckdb.org/why_duckdb.html) is an in-memory, columnar SQL database designed for analytical (OLAP) workloads, offering high-speed data processing and analysis. Its columnar storage model and vectorized query execution make it highly efficient for OLAP tasks, enabling fast aggregation, filtering, and joins on large datasets. DuckDB's ease of integration with data science tools and its ability to run directly within analytical environments like Python and R, without the need for a separate server, make it an attractive choice for OLAP applications seeking simplicity and performance.

By default, Rill includes DuckDB as an embedded OLAP engine that is used to ingest data from [sources](/connect) and power your dashboards. Nothing more needs to be done if you wish to power your dashboards on Rill Developer or Rill Cloud. 

However, you may need to add additional extensions into the embedded DuckDB Engine. To do so, you'll need to define the [Connector YAML](/reference/project-files/connectors#duckdb) and use `init_sql` to install/load/set `extension_name`.

:::tip Optimizing performance on DuckDB

DuckDB is a very useful analytical engine but can start to hit performance and scale challenges as the size of ingested data grows significantly. As a general rule of thumb, we recommend keeping the size of data in DuckDB **under 50GB** along with some other [performance recommendations](/guides/performance). For larger volumes of data, Rill still promises great performance but additional backend optimizations will be needed. [Please contact us](/contact)!

:::

## Connecting to External DuckDB

Along with our embedded DuckDB, Rill provides the ability to connect to external DuckDB database files. This allows you to leverage existing DuckDB databases or create dedicated analytical databases outside of Rill's working directory.

### Configuration

To connect to an external DuckDB database, you'll need to configure a DuckDB connector in your project. Here's an example configuration:

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb
path: "/path/to/your/database.duckdb"
mode: "read"  # or "readwrite" for write access
pool_size: 5
cpu: 4
memory_limit_gb: 16
allow_host_access: true
init_sql: "INSTALL httpfs; LOAD httpfs;"
log_queries: true
```

### Key Configuration Options

**Core Properties:**
- `driver`: Must be "duckdb"
- `mode`: Connection mode - "read" (default) or "readwrite"
- `path`: Path to external DuckDB database file
- `attach`: Full ATTACH statement to attach a DuckDB database

**Resource Management:**
- `pool_size`: Number of concurrent connections and queries allowed
- `cpu`: Number of CPU cores available to the database
- `memory_limit_gb`: Amount of memory in GB available to the database
- `read_write_ratio`: Ratio of resources allocated to read vs write operations (0.0-1.0)
- `allow_host_access`: Whether access to local environment and file system is allowed

**SQL Configuration:**
- `init_sql`: SQL executed during database initialization
- `conn_init_sql`: SQL executed when a new connection is initialized

**Advanced Properties:**
- `log_queries`: Whether to log raw SQL queries executed through OLAP
- `secrets`: Comma-separated list of connector names to create temporary secrets for
- `database_name`: Name of the attached DuckDB database (auto-detected if not set)
- `schema_name`: Default schema used by the DuckDB database

### Important Considerations

:::warning File Location and Size

When connecting to external DuckDB files, consider the following:

- **File Location**: Ensure the DuckDB file is accessible from your Rill environment
- **File Size**: Large DuckDB files (>50GB) may impact performance
- **Permissions**: Ensure Rill has read/write access to the database file
- **Network Access**: For cloud deployments, ensure the file is accessible from the cloud environment

:::

<img src='/img/connect/connector/duckdb.png' class='rounded-gif' style={{width: '75%', display: 'block', margin: '0 auto'}}/>
<br />

Once connected, you can see all the tables in your external DuckDB database within Rill and create models, metrics views, and dashboards using this data.


## Using DuckDB Extensions

DuckDB supports a wide variety of extensions that can enhance its functionality. To use extensions with Rill's embedded DuckDB, configure them in your connector:

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb
init_sql: |
  INSTALL httpfs;
  LOAD httpfs;
  INSTALL spatial;
  LOAD spatial;
```

### Popular Extensions

- **httpfs**: Read and write files over HTTP/HTTPS
- **spatial**: Geospatial data processing
- **json**: Enhanced JSON functionality
- **parquet**: Optimized Parquet file support
- **sqlite_scanner**: Read SQLite databases
- **postgres_scanner**: Read PostgreSQL databases

For a complete list of available extensions, see the [DuckDB Extensions documentation](https://duckdb.org/docs/extensions/overview).

## Performance Optimization

### Memory Management

Configure memory settings based on your data size and available resources:

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb
memory_limit_gb: 32
cpu: 8
pool_size: 10
read_write_ratio: 0.8  # 80% for reads, 20% for writes
```

### Query Optimization

- Use appropriate data types for better compression
- Create indexes on frequently queried columns
- Use `VACUUM` to optimize storage after bulk inserts
- Consider partitioning large tables by date or category

### Best Practices

1. **Data Types**: Use appropriate DuckDB data types (e.g., `DATE`, `TIMESTAMP`, `VARCHAR`)
2. **Indexing**: Create indexes on columns used in WHERE clauses and JOINs
3. **Partitioning**: Partition large tables by time or categorical columns
4. **Compression**: DuckDB automatically compresses data, but consider data type choices
5. **Memory**: Monitor memory usage and adjust `memory_limit_gb` accordingly

## Troubleshooting

### Common Issues

**Connection Errors:**
- Verify the DuckDB file path is correct and accessible
- Check file permissions for read/write access
- Ensure the file is not corrupted

**Performance Issues:**
- Monitor memory usage and adjust `memory_limit_gb`
- Check if `pool_size` is appropriate for your workload
- Consider data partitioning for large datasets

**Extension Errors:**
- Verify extension names are correct
- Check if extensions are compatible with your DuckDB version
- Ensure `init_sql` syntax is valid

### Debugging

Enable query logging to debug performance issues:

```yaml
# connectors/duckdb.yaml
type: connector
driver: duckdb
log_queries: true
```

This will log all SQL queries executed through the OLAP interface, helping you identify slow queries and optimization opportunities.

## Multiple Engines 

While not recommended, it is possible in Rill to use multiple OLAP engines in a single project. For more information, see our page on [Using Multiple OLAP Engines](/connect/olap/multiple-olap).

## Additional Notes

- For dashboards powered by DuckDB, [measure definitions](/build/metrics-view/#measures) are required to follow standard [DuckDB SQL](https://duckdb.org/docs/sql/introduction) syntax.
- There is a known issue around creating a DuckDB source via the UI; you will need to create the YAML file manually.
- DuckDB supports most standard SQL functions and operators, making it easy to write complex analytical queries.
- For advanced analytics, consider using DuckDB's window functions, CTEs, and other advanced SQL features.