---
title: "Rill Project Logs"
description: Alter dashboard look and feel
sidebar_label: "Rill Project Logs"
sidebar_position: 30
---

Whether you start Rill from the terminal or your favorite IDE, that window will output the project logs. From reconciling items to partition ingestion and beyond, browsing the project logs is a great place to start when troubleshooting errors or slow loading models.


## Dissecting the Format of Common Logs

```bash
Reconciled resource             {"name": "commits__ (copy)_metrics_explore", "type": "Explore", "elapsed": "1ms"}
Executed model partition        {"model": "CH_incremental_commits_directory", "key": "55454ed4ad31cd3266988fe523103637", "data": {"path":"github-analytics/Clickhouse/2025/08","uri":"gs://rilldata-public/github-analytics/Clickhouse/2025/08"}, "elapsed": "283.188333ms"}
# debug
Executed model partition        {"model": "staging_to_CH", "key": "0030406e528b3799c8cbad6bfe609e83", "trace_id": "3073a89ac5cee9e7e3433ce0a34d291a", "span_id": "c3cb402d7b4af9b6", "data": {"day":"2022-12-20T00:00:00Z"}}
# verbose
grpc finished call      {"protocol": "grpc", "peer.address": "::1", "grpc.component": "server", "grpc.method_type": "unary", "grpc.method": "/rill.runtime.v1.RuntimeService/GetResource", "instance_id": "default", "args.instance_id": "default", "args.name.kind": "rill.runtime.v1.Theme", "args.name.name": "theme", "args.skip_security_checks": false, "grpc.code": "OK", "duration": "38.125µs"}
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

### Debug
- **`sql`** – SQL statement executed by DuckDB during model evaluation or metrics computation.  
- **`args`** – SQL query parameters (if any).  
- **`trace_id`** – Unique trace identifier for the operation (used for distributed tracing).  
- **`span_id`** – Unique span identifier within the trace.  

### Verbose
- **`protocol`** – Protocol used for gRPC calls (e.g., `"grpc"`).  
- **`peer.address`** – Remote address of the gRPC peer.  
- **`grpc.component`** – Component type for the gRPC call (usually `"server"`).  
- **`grpc.method_type`** – gRPC method type (e.g., `"unary"`).  
- **`grpc.method`** – Full gRPC method name being called.  
- **`grpc.code`** – Status code returned by the gRPC call.  
- **`duration`** – Duration of the gRPC call or query execution.  
- **`args.instance_id`** – Instance ID argument passed in a gRPC request.  
- **`args.glob`** – Glob pattern argument for file listing operations.  
- **`args.kind`** – Kind of Rill resource being listed or retrieved.  
- **`args.skip_security_checks`** – Boolean flag indicating if security checks were skipped.  
- **`args.path`** – Path argument for file retrieval operations.  
- **`args.name.kind`** – Kind of resource name provided in a `GetResource` call.  
- **`args.name.name`** – Name of resource provided in a `GetResource` call.  


## Logging Examples

### Project Creation

When you first initialize a Rill project, you'll see Rill reconcile a resource "duckdb" of type "connector". This is expected as we explicitly create this file to initialize a connection to our embedded DuckDB.

```bash
Rill will create project files in "~/Desktop/GitHub/testing-folder/dsn". Do you want to continue? Yes
2025-08-05T15:45:21.932 INFO    Serving Rill on: http://localhost:9009
2025-08-05T15:45:31.491 INFO    Reconciling resource    {"name": "duckdb", "type": "Connector"}
2025-08-05T15:45:31.581 INFO    Reconciled resource     {"name": "duckdb", "type": "Connector", "elapsed": "90ms"}
```

### Connecting to a Data Source

When connecting to a data source via a connector, you'll see a "Connector" being reconciled. In the case of any errors, you'll see this in both the UI and the logs.

```bash
2025-08-05T16:50:42.393 INFO    Reconciling resource    {"name": "gcs", "type": "Connector"}
2025-08-05T16:50:42.432 INFO    Reconciled resource     {"name": "gcs", "type": "Connector", "elapsed": "39ms"}
2025-08-05T16:55:47.725 WARN    Reconcile failed        {"name": "gcs", "type": "Connector", "elapsed": "1ms", "error": "failed to resolve templated property \"google_application_credentials\": template: :1:6: executing \"\" at <.env.connector.gcs.google_application_credentialsss>: map has no entry for key \"google_application_credentials\""}
```

Once connected, you'll likely create a model and see this also reconciling in the logs. Similar to the above, if there are any issues, you'll see it both in the logs and UI.
  
```bash
2025-08-05T16:43:06.403 INFO    Reconciling resource    {"name": "commits__", "type": "Model"}
2025-08-05T16:43:07.348 INFO    Reconciled resource     {"name": "commits__", "type": "Model", "elapsed": "944ms"}
#or
2025-08-05T16:58:17.137 WARN    Reconcile failed        {"name": "commits__", "type": "Model", "elapsed": "682ms", "error": "blob (key \"github-analytics/Clickhouse/2025/06/commits_2025_0.parquet\") (code=Unknown): storage: object doesn't exist: googleapi: Error 404: No such object: rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_0.parquet, notFound", "errorVerbose": "blob (key \"github-analytics/Clickhouse/2025/06/commits_2025_0.parquet\") (code=Unknown):\n    gocloud.dev/blob.(*Bucket).Attributes\n        /Users/runner/go/pkg/mod/gocloud.dev@v0.36.0/blob/blob.go:913\n  - storage: object doesn't exist: googleapi: Error 404: No such object: rilldata-public/github-analytics/Clickhouse/2025/06/commits_2025_0.parquet, notFound"}
```

### Creating Rill Objects

The next section of logs is creating a metrics view and explore dashboard. You'll see some errors are thrown in the metrics view and resolved in Rill Developer.

```bash
2025-08-05T17:00:08.191 INFO    Reconciling resource    {"name": "commits___metrics", "type": "MetricsView"}
2025-08-05T17:00:08.206 WARN    Reconcile failed        {"name": "commits___metrics", "type": "MetricsView", "elapsed": "15ms", "error": "measure \"earliest_commit_date_measure\" is of type CODE_TIMESTAMP, but must be a numeric type\nmeasure \"latest_commit_date_measure\" is of type CODE_TIMESTAMP, but must be a numeric type"}
2025-08-05T17:00:24.948 INFO    Reconciling resource    {"name": "commits___metrics", "type": "MetricsView"}
2025-08-05T17:00:24.969 WARN    Reconcile failed        {"name": "commits___metrics", "type": "MetricsView", "elapsed": "21ms", "error": "measure \"earliest_commit_date_measure\" is of type CODE_TIMESTAMP, but must be a numeric type"}
2025-08-05T17:00:33.899 INFO    Reconciling resource    {"name": "commits___metrics", "type": "MetricsView"}
2025-08-05T17:00:33.914 INFO    Reconciled resource     {"name": "commits___metrics", "type": "MetricsView", "elapsed": "15ms"}
2025-08-05T17:03:21.502 INFO    Reconciling resource    {"name": "commits___metrics_explore", "type": "Explore"}
2025-08-05T17:03:21.503 INFO    Reconciled resource     {"name": "commits___metrics_explore", "type": "Explore", "elapsed": "1ms"}
```

### Partitioned Models

The main takeaway for partitioned models is that you'll be able to see the number of partitions that Rill will start ingesting. This is especially important when creating [dev/prod](/deploy/templating) environments and you're trying to avoid ingesting large amounts of data locally.

```bash
Resolved model partitions       {"model": "staging_to_CH", "partitions": 16}
2025-08-05T17:19:09.269 INFO    Executed model partition        {"model": "staging_to_CH", "key": "0030406e528b3799c8cbad6bfe609e83", "data": {"day":"2022-12-20T00:00:00Z"}}
```


## Debugging Complicated Issues

Sometimes the default logs don't have enough information to figure out why something isn't working as planned. In these cases, you can set `--debug` or `--verbose` to get more information.

```bash
rill start --debug
rill start --verbose
```


## Rill Cloud Logs

Similar to the Rill Developer experience, you can view the logs in Rill Cloud to see the progress of your reconciliation.

```
rill project logs
```

To continually view the progression of your logs, use `-f` or `--follow`.

```
rill project logs -f
```
