---
title: "Debugging Rill Projects"
description: "Tools and techniques for debugging Rill projects"
sidebar_label: "Debugging"
sidebar_position: 30
---

When building Rill projects, you'll encounter various debugging scenarios—from understanding project logs to tracing resource reconciliation. This section covers the tools and techniques available for troubleshooting your Rill projects.

- **[Understanding Project Logs](#understanding-project-logs)** - Learn the basics of reading and interpreting logs
- **[Troubleshooting Common Errors](#troubleshooting-common-errors)** - Resolve common error patterns
- **[Advanced Debugging Techniques](#advanced-debugging-techniques)** - Use debug flags, trace viewer, and cloud logs



## Understanding Project Logs

Whether you start Rill from the terminal or your favorite IDE, the terminal window will output the project logs. From reconciling items to partition ingestion and beyond, browsing the project logs is a great place to start when troubleshooting errors or slow-loading models.

### Log Format

Rill logs follow a structured JSON format. Here are some common log entries:

```bash
Reconciled resource             {"name": "commits__ (copy)_metrics_explore", "type": "Explore", "elapsed": "1ms"}
Executed model partition        {"model": "CH_incremental_commits_directory", "key": "55454ed4ad31cd3266988fe523103637", "data": {"path":"github-analytics/Clickhouse/2025/08","uri":"gs://rilldata-public/github-analytics/Clickhouse/2025/08"}, "elapsed": "283.188333ms"}
Executed model partition        {"model": "staging_to_CH", "key": "0030406e528b3799c8cbad6bfe609e83", "trace_id": "3073a89ac5cee9e7e3433ce0a34d291a", "span_id": "c3cb402d7b4af9b6", "data": {"day":"2022-12-20T00:00:00Z"}}
```

### Generic Logging

- **`name`** – Filename or YAML-defined name of the Rill object.  
- **`type`** – Resource type (e.g., `Connector`, `Model`, `MetricsView`, `Explore`, `API`, `Alert`, `Theme`, `Component`, `Canvas`).  
- **`elapsed`** – Time taken to reconcile, execute, or otherwise process the resource.  
- **`error`** – Error message generated during reconciliation or execution.  
- **`dependency_error`** – Boolean flag indicating that the resource failed due to another resource's error.  
- **`deleted`** – Boolean flag indicating that the resource was deleted.  
- **`path`** – Filesystem path to the resource YAML file.  
- **`logger_name`** – Name of the logger emitting the message.  
- **`message`** – Log message content.  

### Partitioning

- **`model`** – Name of the model associated with a partition operation.  
- **`partitions`** – Number of partitions resolved for a model.  
- **`key`** – Partition key (usually an MD5-like hash).  
- **`data`** – Partition-specific metadata or parameters.  

### Embedded ClickHouse

- **`addr`** – Host and port address for the embedded ClickHouse server. 
- **`sql`** – SQL statement executed by ClickHouse during query execution.   

### Debug

- **`sql`** – SQL statement executed by DuckDB during model evaluation or metrics computation.  
- **`args`** – SQL query parameters (if any).  
- **`trace_id`** – Unique trace identifier for the operation (used for distributed tracing).  
- **`span_id`** – Unique span identifier within the trace.



### Logging Examples

#### Simple: Project Creation

When you first initialize a Rill project, you'll see Rill reconcile a resource "duckdb" of type "Connector". This is expected as we explicitly create this file to initialize a connection to our embedded DuckDB.

```bash
Rill will create project files in "~/Desktop/GitHub/testing-folder/dsn". Do you want to continue? Yes
INFO    Serving Rill on: http://localhost:9009
INFO    Reconciling resource    {"name": "duckdb", "type": "Connector"}
INFO    Reconciled resource     {"name": "duckdb", "type": "Connector", "elapsed": "90ms"}
```

#### Simple: Connecting to a Data Source

When connecting to a data source via a connector, you'll see a "Connector" being reconciled. In the case of any errors, you'll see this in both the UI and the logs.

```bash
INFO    Reconciling resource    {"name": "gcs", "type": "Connector"}
INFO    Reconciled resource     {"name": "gcs", "type": "Connector", "elapsed": "39ms"}
WARN    Reconcile failed        {"name": "gcs", "type": "Connector", "elapsed": "1ms", "error": "failed to resolve templated property \"google_application_credentials\": template: :1:6: executing \"\" at <.env.connector.gcs.google_application_credentialsss>: map has no entry for key \"google_application_credentials\""}
```

Once connected, you'll likely create a model and see this also reconciling in the logs. Similar to the above, if there are any issues, you'll see it both in the logs and UI.

```bash
INFO    Reconciling resource    {"name": "commits__", "type": "Model"}
INFO    Reconciled resource     {"name": "commits__", "type": "Model", "elapsed": "944ms"}
# or
WARN    Reconcile failed        {"name": "commits__", "type": "Model", "elapsed": "682ms", "error": "blob (key \"github-analytics/Clickhouse/2025/06/commits_2025_0.parquet\") (code=Unknown): storage: object doesn't exist: googleapi: Error 404: No such object: rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_0.parquet, notFound", "errorVerbose": "blob (key \"github-analytics/Clickhouse/2025/06/commits_2025_0.parquet\") (code=Unknown):\n    gocloud.dev/blob.(*Bucket).Attributes\n        /Users/runner/go/pkg/mod/gocloud.dev@v0.36.0/blob/blob.go:913\n  - storage: object doesn't exist: googleapi: Error 404: No such object: rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_0.parquet, notFound"}
```

#### Intermediate: Creating Rill Objects

The next section of logs shows the creation of a metrics view and explore dashboard. You'll see some errors thrown in the metrics view that get resolved in Rill Developer.

```bash
INFO    Reconciling resource    {"name": "commits___metrics", "type": "MetricsView"}
WARN    Reconcile failed        {"name": "commits___metrics", "type": "MetricsView", "elapsed": "15ms", "error": "measure \"earliest_commit_date_measure\" is of type CODE_TIMESTAMP, but must be a numeric type\nmeasure \"latest_commit_date_measure\" is of type CODE_TIMESTAMP, but must be a numeric type"}
INFO    Reconciling resource    {"name": "commits___metrics", "type": "MetricsView"}
WARN    Reconcile failed        {"name": "commits___metrics", "type": "MetricsView", "elapsed": "21ms", "error": "measure \"earliest_commit_date_measure\" is of type CODE_TIMESTAMP, but must be a numeric type"}
INFO    Reconciling resource    {"name": "commits___metrics", "type": "MetricsView"}
INFO    Reconciled resource     {"name": "commits___metrics", "type": "MetricsView", "elapsed": "15ms"}
INFO    Reconciling resource    {"name": "commits___metrics_explore", "type": "Explore"}
INFO    Reconciled resource     {"name": "commits___metrics_explore", "type": "Explore", "elapsed": "1ms"}
```

#### Advanced: Dependency Errors

When a resource fails, dependent resources will also fail with a `dependency_error` flag. This helps you trace the root cause of cascading failures. In the example below, the `orders` model fails, which causes `orders_customers_model` to fail with a dependency error, which in turn causes `orders_customers_metrics` and `orders_customers_explore` to fail.

```bash
INFO    Reconciling resource    {"name": "duckdb", "type": "Connector"}
INFO    Reconciling resource    {"name": "gcs", "type": "Connector"}
INFO    Reconciling resource    {"name": "orders", "type": "Model"}
INFO    Reconciling resource    {"name": "customers", "type": "Model"}
INFO    Reconciled resource     {"name": "gcs", "type": "Connector", "elapsed": "1ms"}
INFO    Reconciled resource     {"name": "duckdb", "type": "Connector", "elapsed": "96ms"}
INFO    Reconciled resource     {"name": "customers", "type": "Model", "elapsed": "10.8s"}
WARN    Reconcile failed        {"name": "orders", "type": "Model", "elapsed": "17.5s", "error": "failed to create model: Cannot open file \"/path/to/project/tmp/default/duckdb/orders/data.db.wal\": No such file or directory"}
INFO    Reconciling resource    {"name": "orders_customers_model", "type": "Model"}
INFO    Reconciled resource     {"name": "orders_customers_model", "type": "Model", "error": "dependency error: resource \"orders\" (rill.runtime.v1.Model) has an error", "dependency_error": true}
INFO    Reconciling resource    {"name": "orders_customers_metrics", "type": "MetricsView"}
WARN    Reconcile failed        {"name": "orders_customers_metrics", "type": "MetricsView", "elapsed": "11ms", "error": "table \"orders_customers_model\" does not exist"}
INFO    Reconciling resource    {"name": "orders_customers_explore", "type": "Explore"}
INFO    Reconciled resource     {"name": "orders_customers_explore", "type": "Explore", "error": "dependency error: resource \"orders_customers_metrics\" (rill.runtime.v1.MetricsView) has an error", "dependency_error": true}
```

#### Advanced: Partitioned Models

The main takeaway for partitioned models is that you'll be able to see the number of partitions that Rill will start ingesting. This is especially important when creating [dev/prod](/developer/build/connectors/templating) environments and you're trying to avoid ingesting large amounts of data locally.

```bash
Resolved model partitions       {"model": "staging_to_CH", "partitions": 16}
INFO    Executed model partition        {"model": "staging_to_CH", "key": "0030406e528b3799c8cbad6bfe609e83", "data": {"day":"2022-12-20T00:00:00Z"}}
```

## Troubleshooting Common Errors

When debugging errors, start by checking the project logs and understanding the error messages. Here are common error patterns and how to resolve them:

### Model Errors

Model errors typically occur when there are issues with credentials, data processing, SQL syntax, or data type mismatches. Common error messages and their solutions:

- **`Failed to connect to ...`**: Issue with your connector. Check your credentials and [firewall settings](/developer/build/connectors/data-source#externally-hosted-services) if using externally hosted services
- **`Table with name ... does not exist!`**: Verify the table exists by running `rill query --sql "select * from {table_name} limit 1"` or checking your data source
- **`IO Error: No files found that match the pattern...`**: Check that your cloud storage folder path is correct and files exist
- **`some partitions have errors`**: Run `rill project refresh --model {model_name} --errored-partitions` to refresh errored partitions
- **`Out of Memory Error: ...`**: Contact [support](/contact) for assistance with memory issues

### Metrics View and Dashboard Errors

Metrics view and dashboard errors often stem from issues with the underlying models or configuration problems:

- **Model Dependencies:** Dashboards failing because their underlying models have errors. Check the [dependency errors](/developer/build/debugging#model-errors) section above
- **Missing Dimensions/Measures:** References to fields that don't exist in the underlying model. Verify that measures and dimensions in your metrics YAML match existing columns in your data
- **Type Mismatches:** Measures must be numeric types. Check that timestamp fields aren't being used as measures

### Checking Resource Status

To understand what's failing in your project:

1. **Check project logs** - Review the terminal output or use `rill project logs` for Rill Cloud projects
2. **Use the Trace Viewer** - Visualize resource reconciliation and trace execution paths
3. **Check resource status** - Use the `Status` tab in Rill Developer or [`rill project status`](/reference/cli/project/status) CLI command

:::tip Check upstream dependencies
The surfaced error might not be the root cause. A dashboard error could stem from an underlying model timeout. Always trace errors to their source by checking dependent resources.
:::

## Advanced Debugging Techniques

When standard logs aren't providing enough detail, Rill offers several advanced debugging options to help you diagnose issues more effectively.

### Using Debug and Verbose Flags

Rill Developer provides two flags for increasing log verbosity:

**`--verbose`**: Sets the log level to debug, showing more detailed information about what Rill is doing internally. This includes:
- More granular resource reconciliation steps
- Additional context about operations
- Extended error details

**`--debug`**: Collects additional debug information beyond just log verbosity. This flag enables:
- Enhanced debugging metadata
- More detailed trace information
- Additional diagnostic data useful for troubleshooting complex issues

```bash
# Increase log verbosity
rill start --verbose

# Collect additional debug info
rill start --debug

# Combine both for maximum detail
rill start --debug --verbose
```

:::tip When to use each flag
- Use `--verbose` when you need more detail about what Rill is doing but don't need deep debugging info
- Use `--debug` when troubleshooting complex issues that require additional diagnostic data
- Use both together when you need the most comprehensive debugging information
:::

### Log Format Options

By default, Rill outputs logs in a human-readable console format. For programmatic processing or filtering, you can output logs in JSON format:

```bash
rill start --log-format json
```

JSON format is useful when:
- Parsing logs with scripts or tools
- Filtering logs programmatically
- Integrating with log aggregation systems

### Viewing Rill Cloud Logs

For projects deployed to Rill Cloud, you can view logs directly from the CLI:

```bash
# View recent logs
rill project logs <project-name>

# Follow logs in real-time (like tail -f)
rill project logs <project-name> --follow

# Show only the last N lines
rill project logs <project-name> --tail 100

# Filter by log level
rill project logs <project-name> --level DEBUG
```

The `rill project logs` command provides the same structured log output you see in Rill Developer, making it easy to debug issues in production deployments.

### Checking Project Status

Use the `rill project status` command to get a quick overview of your project's health:

```bash
# Check status of a deployed project
rill project status <project-name>

# Check status of locally running project
rill project status --local
```

This command shows:
- Resource reconciliation status
- Error states for individual resources
- Dependency relationships
- Overall project health


### Trace Viewer

For complex debugging scenarios involving multiple resources and dependencies, use the [Trace Viewer](/developer/build/debugging/trace-viewer) to visualize resource reconciliation and trace execution paths across your project. The Trace Viewer helps you:

- Understand resource dependency chains
- Identify bottlenecks in reconciliation
- Visualize execution flows
- Debug cascading failures

To use the Trace Viewer, start Rill with the `--debug` flag:

```bash
rill start --debug
```

Then access the Trace Viewer through the Rill Developer UI to see a visual representation of your project's resource reconciliation.
