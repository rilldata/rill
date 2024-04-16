---
title: Druid
description: Power Rill dashboards using Druid
sidebar_label: Druid
sidebar_position: 2
---

## Overview

[Apache Druid](https://druid.apache.org/docs/latest/design/) is an open-source, high-performance OLAP engine designed for real-time analytics on large datasets. It excels in analytical workloads due to its columnar storage format, which enables fast data aggregation, querying, and filtering. Druid is particularly well-suited for use cases that require interactive exploration of large-scale data, real-time data ingestion, and fast query responses, making it a popular choice for applications in business intelligence, user behavior analytics, and financial analysis.

Rill supports connecting to an existing Druid cluster and using it as an OLAP engine to power Rill dashboards built against [external tables](build/olap/olap.md#external-olap-tables). This is particularly useful when working with extremely large datasets (hundreds of GBs or even TB+ in size).

## Supported versions

Rill supports connecting to Druid v26.0 or newer versions.

## Connection string (DSN)

Rill connects to Druid using the [HTTP API](https://druid.apache.org/docs/latest/api-reference/sql-api) and requires using a connection string of the following format: `http://<user>:<password>@<host>:<port>/druid/v2/sql`. If `user` or `password` contain special characters they should be URL encoded (ie `p@ssword` -> `p%40ssword`). This should be set in the `connector.druid.dsn` property in Rill.

As an example, this typically looks something like:

```bash

connector.druid.dsn="https://user:password@localhost:8888/druid/v2/sql"

```

:::info Need help connecting to Druid?

If you would like to connect Rill to an existing Druid instance, please don't hesitate to [contact us](../../contact.md). We'd love to help!

:::

## Setting the default OLAP connection

You'll also need to update the `olap_connector` property in your project's `rill.yaml` to change the default OLAP engine to Druid:

```yaml

olap_connector: druid

```

:::note

For more information about available properties in `rill.yaml`, see our [project YAML](../project-files/rill-yaml.md) documentation.

:::

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](multiple-olap.md) page.

:::

## Configuring Rill Developer

When using Rill for local development, there are two options to configure Rill to enable Druid as an OLAP engine:
- You can set `connector.druid.dsn` in your project's `.env` file or try pulling existing credentials locally using `rill env pull` if the project has already been deployed to Rill Cloud
- You can pass in `connector.druid.dsn` as a variable to `rill start` directly (e.g. `rill start --var connector.druid.dsn=...`)

:::tip Getting DSN errors in dashboards after setting `.env`?

If you are facing issues related to DSN connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.

:::

## Configuring Rill Cloud

When deploying a Druid-backed project to Rill Cloud, you have the following options to pass the appropriate connection string to Rill Cloud:
- Use the `rill env configure` command to set `connector.druid.dsn` after deploying the project
- If `connector.druid.dsn` has already been set in your project `.env`, you can push and update these variables directly in your cloud deployment by using the `rill env push` command

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

## Additional Notes

- At the moment, we do not support modeling with Druid. If this is something you're interested in, please [contact us](../../contact.md).
- For dashboards powered by Druid, [measure definitions](/build/dashboards/dashboards.md#measures) are required to follow standard [Druid SQL](https://druid.apache.org/docs/latest/querying/sql/) syntax.