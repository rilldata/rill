---
title: "Rill Project Logs"
description: Alter dashboard look and feel
sidebar_label: "Rill Project Logs"
sidebar_position: 30
---

Whether you start Rill from the terminal or your favorite IDE, that window will output the project logs. From reconciling items to partitions ingestion and beyond, browsing the project logs is a great place to start when troubleshooting errors or slow loading models.


## Logging Examples

### Project Creation

When you first create a rill project, you'll see Rill reconcile a resource "duckdb" of type "connector". This is expected as we explicitly create this file to initialize a connection to our embedded DuckDB.

```bash
 Rill will create project files in "~/Desktop/GitHub/testing-folder/dsn". Do you want to continue? Yes
2025-08-05T15:45:21.932 INFO    Serving Rill on: http://localhost:9009
2025-08-05T15:45:31.491 INFO    Reconciling resource    {"name": "duckdb", "type": "Connector"}
2025-08-05T15:45:31.581 INFO    Reconciled resource     {"name": "duckdb", "type": "Connector", "elapsed": "90ms"}
```

### Connecting to a Data Source


```bash
2025-08-05T16:43:06.403 INFO    Reconciling resource    {"name": "commits__", "type": "Model"}
2025-08-05T16:43:07.348 INFO    Reconciled resource     {"name": "commits__", "type": "Model", "elapsed": "944ms"}
```

### Creating Rill Objects


### Refreshing Models


### Partitioned Models



## Debugging Complicated Issues

rill start --debug



rill start --verbose

