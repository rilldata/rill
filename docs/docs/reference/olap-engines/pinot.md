---
title: Pinot
description: Power Rill dashboards using Pinot
sidebar_label: Pinot
sidebar_position: 4
---

## Overview

[Apache Pinot](https://docs.pinot.apache.org/) is a real-time distributed OLAP datastore purpose-built for low-latency, high-throughput analytics, and perfect for user-facing analytical workloads.

Rill supports connecting to an existing Pinot cluster and using it as an OLAP engine to power Rill dashboards built against [external tables](build/olap/olap.md#external-olap-tables).

## Connection string (DSN)

Rill connects to Pinot using the [Pinot Golang Client](https://docs.pinot.apache.org/users/clients/golang) and requires a connection string of the following format: `http://<user>:<password>@<host>:<port>`. 
`host`and `port` should be of Pinot Controller server. If `user` or `password` contain special characters they should be URL encoded (ie `p@ssword` -> `p%40ssword`). This should be set in the `connector.pinot.dsn` property in Rill.

As an example, this typically looks something like:

```bash

connector.pinot.dsn="https://username:password@localhost:9000"

```

:::info Need help connecting to Pinot?

If you would like to connect Rill to an existing Pinot instance, please don't hesitate to [contact us](../../contact.md). We'd love to help!

:::

## Setting the default OLAP connection

You'll also need to update the `olap_connector` property in your project's `rill.yaml` to change the default OLAP engine to Pinot:

```yaml

olap_connector: pinot

```

:::note

For more information about available properties in `rill.yaml`, see our [project YAML](../project-files/rill-yaml.md) documentation.

:::

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](multiple-olap.md) page.

:::

## Configuring Rill Developer

When using Rill for local development, there are two options to configure Rill to enable Pinot as an OLAP engine:
- You can set `connector.pinot.dsn` in your project's `.env` file or try pulling existing credentials locally using `rill env pull` if the project has already been deployed to Rill Cloud
- You can pass in `connector.pinot.dsn` as a variable to `rill start` directly (e.g. `rill start --var connector.pinot.dsn=...`)

:::tip Getting DSN errors in dashboards after setting `.env`?

If you are facing issues related to DSN connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.

:::

## Configuring Rill Cloud

When deploying a Pinot-backed project to Rill Cloud, you have the following options to pass the appropriate connection string to Rill Cloud:
- Use the `rill env configure` command to set `connector.pinot.dsn` after deploying the project
- If `connector.pinot.dsn` has already been set in your project `.env`, you can push and update these variables directly in your cloud deployment by using the `rill env push` command

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

## Additional Notes

- At the moment, we do not support modeling with Pinot. If this is something you're interested in, please [contact us](../../contact.md).
- For dashboards powered by Pinot, [measure definitions](/build/dashboards/dashboards.md#measures) are required to follow [Pinot SQL](https://docs.pinot.apache.org/users/user-guide-query/querying-pinot) syntax.