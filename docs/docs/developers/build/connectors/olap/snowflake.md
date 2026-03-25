---
title: Snowflake
description: Power Rill dashboards using Snowflake as an OLAP engine
sidebar_label: Snowflake
sidebar_position: 26
---

:::caution Beta

Snowflake as an OLAP engine is currently in beta. Functionality may change and some features may not yet be fully supported.

:::

[Snowflake](https://www.snowflake.com/) is a cloud data platform that provides data warehousing, data lake, and data sharing capabilities. Rill supports connecting to Snowflake as a read-only OLAP engine to power dashboards using your existing Snowflake tables.

:::info

Rill supports connecting to an existing Snowflake warehouse via a read-only OLAP connector and using it to power Rill dashboards with [external tables](/developers/build/connectors/olap#external-olap-tables).

:::

## Connect to Snowflake

After selecting "Add Data", select Snowflake and fill in your connection parameters. This will automatically create the `snowflake.yaml` file in your `connectors` directory.

### Connection Parameters

```yaml
type: connector
driver: snowflake

account: <ACCOUNT_IDENTIFIER>
username: <USERNAME>
password: "{{ .env.SNOWFLAKE_PASSWORD }}"
database: <DATABASE>
schema: <SCHEMA>
warehouse: <WAREHOUSE>
role: <ROLE>
```

### Setting the OLAP Connector

To use Snowflake as the OLAP engine for your project, update your `rill.yaml`:

```yaml
olap_connector: snowflake
```

Or set it via the CLI:

```bash
rill env set olap_connector snowflake
```

## Deploy to Rill Cloud

When deploying to Rill Cloud, ensure your Snowflake credentials are configured:

```bash
rill env set SNOWFLAKE_PASSWORD <password> --project <project-name>
```
