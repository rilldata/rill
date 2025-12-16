---
title: Control Connector Permissions
description: read only vs read/write permissions on live connectors
sidebar_label: Connector Mode
sidebar_position: 30
---

When creating a connection to an OLAP engine (like ClickHouse Cloud or MotherDuck), Rill provides a `mode` setting that controls whether the connection can only read data or also write/modify data. This is critical for preventing accidental data modification when connecting to existing production databases.

## Modes

The connector supports two modes:

*   **`read`** (Default): Restricts the connection to read-only operations. This is the safest mode for connecting to existing OLAP databases where you only want to query data for dashboards.
    *   In this mode, Rill disables model execution and table mutations.
    *   You cannot create new models or ingest data *into* this connector using Rill.
*   **`readwrite`**: Enables full read and write capabilities.
    *   Allows Rill to create, drop, and modify tables.
    *   Required if you want to use this connector as a destination for model materialization (e.g., ingesting data from S3 into ClickHouse via Rill).
    *   Automatically set to `readwrite` for Rill-managed embedded connectors.

## Configuration

You can set the mode in your connector YAML file:

```yaml
type: connector
driver: clickhouse
# ... other connection properties ...
mode: "read" # or "readwrite"
```

:::warning Data Safety
We strongly recommend leaving the mode as `read` (the default) when connecting to an external, pre-existing OLAP database to avoid any risk of accidental data loss or schema changes. Only use `readwrite` if you specifically intend for Rill to manage tables within that database.
:::
