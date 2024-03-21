---
title: ClickHouse
description: Power Rill dashboards using ClickHouse
sidebar_label: ClickHouse
sidebar_position: 3
---

## Overview

[ClickHouse](https://clickhouse.com/docs/en/intro) is an open-source, column-oriented OLAP database management system known for its ability to perform real-time analytical queries on large-scale datasets. Its architecture is optimized for high performance, leveraging columnar storage and advanced compression techniques to speed up data reads and significantly reduce storage costs. ClickHouse's efficiency in query execution, scalability, and ability to handle even petabytes of data makes it an excellent choice for real-time analytic use cases. 

Rill supports connecting to an existing ClickHouse instance and using it as an OLAP engine to power Rill dashboards built against [external tables](build/olap/olap.md#external-olap-tables). This is particularly useful when working with extremely large datasets (hundreds of GBs or even TB+ in size).


![Rill on ClickHouse](/img/reference/olap-engines/clickhouse/clickhouse.gif)

## Connection string (DSN)

Rill is able to connect to ClickHouse using the [ClickHouse Go Driver](https://clickhouse.com/docs/en/integrations/go). An appropriate connection string (DSN) will need to be set through the `connection.clickhouse.dsn` property in Rill.

A very simple example might look like the following:

```bash

connector.clickhouse.dsn="clickhouse://<hostname>:<port>?username=<username>&password=<password>"

```

:::info Check your port

In most situations, the default port is 9440 for TLS and 9000 when not using TLS. However, it is worth double checking the port that your ClickHouse instance is configured to use when setting up your connection string.

:::

:::note DSN properties

For more information about available DSN properties and setting an appropriate connection string, please refer to ClickHouse's [documentation](https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#dsn).

:::

### Connecting to ClickHouse Cloud

If you are connecting to an existing [ClickHouse Cloud](https://clickhouse.com/cloud) instance, you can retrieve the connection string from the admin navigation panel.

![ClickHouse Cloud connection string](/img/reference/olap-engines/clickhouse/clickhouse-cloud.png)

Because ClickHouse Cloud requires a secure connection over [https](https://github.com/ClickHouse/clickhouse-go?tab=readme-ov-file#http-support-experimental), you can pass in a https URL with `secure=true` and `skip_verify=true` as additional URL parameters:

```bash

connector.clickhouse.dsn="https://<hostname>:<port>?username=<username>&password=<password>&secure=true&skip_verify=true"

```

:::info Need help connecting to ClickHouse?

If you would like to connect Rill to an existing ClickHouse instance, please don't hesitate to [contact us](../../contact.md). We'd love to help!

:::

## Setting the default OLAP connection

You'll also need to update the `olap_connector` property in your project's `rill.yaml` to change the default OLAP engine to ClickHouse:

```yaml

olap_connector: clickhouse

```

:::note

For more information about available properties in `rill.yaml`, see our [project YAML](../project-files/rill-yaml.md) documentation.

:::

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](multiple-olap.md) page.

:::

## Configuring Rill Developer

When using Rill for local development, there are two options to configure Rill to enable ClickHouse as an OLAP engine:
- You can set `connector.clickhouse.dsn` in your project's `.env` file or try pulling existing credentials locally using `rill env pull` if the project has already been deployed to Rill Cloud
- You can pass in `connector.clickhouse.dsn` as a variable to `rill start` directly (e.g. `rill start --env connector.clickhouse.dsn=...`)

:::tip Getting DSN errors in dashboards after setting `.env`?

If you are facing issues related to DSN connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.

:::

## Configuring Rill Cloud

When deploying a ClickHouse-backed project to Rill Cloud, you have the following options to pass the appropriate connection string to Rill Cloud:
- Use the `rill env configure` command to set `connector.clickhouse.dsn` after deploying the project
- If `connector.clickhouse.dsn` has already been set in your project `.env`, you can push and update these variables directly in your cloud deployment by using the `rill env push` command

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

## Additional Notes

- At the moment, we do not officially support modeling with ClickHouse. If this is something you're interested in, please [contact us](../../contact.md).
- For dashboards powered by ClickHouse, [measure definitions](/build/dashboards/dashboards.md#measures) are required to follow standard [ClickHouse SQL](https://clickhouse.com/docs/en/sql-reference) syntax.