---
title: Druid
description: Power Rill dashboards using Druid
sidebar_label: Druid
sidebar_position: 05
---

[Apache Druid](https://druid.apache.org/docs/latest/design/) is an open-source, high-performance OLAP engine designed for real-time analytics on large datasets. It excels in analytical workloads due to its columnar storage format, which enables fast data aggregation, querying, and filtering. Druid is particularly well-suited for use cases that require interactive exploration of large-scale data, real-time data ingestion, and fast query responses, making it a popular choice for applications in business intelligence, user behavior analytics, and financial analysis.

Rill supports connecting to an existing Druid cluster via a "live connector" and using it as an OLAP engine  built against [external tables](/build/connectors/olap#external-olap-tables) to power Rill dashboards. This is particularly useful when working with extremely large datasets (hundreds of GBs or even TB+ in size).


## Configuring Rill Developer with Druid

When using Rill for local development, there are a few options to configure Rill to enable Druid as an OLAP engine:
1. Connect to an OLAP engine via Add Data. This will automatically create the `druid.yaml` file in your `connectors` directory and populate the `.env` file with `connector.druid.password` or `connector.druid.dsn` depending on which you select in the UI.

For more information on supported parameters, see our [Druid connector YAML reference docs](/reference/project-files/connectors#druid).

```yaml 
type: connector

driver: druid
host: <HOSTNAME>
port: <PORT>
username: <USERNAME>
password: "{{ .env.connector.druid.password }}"
ssl: true 

# or 

dsn: "{{ .env.connector.druid.dsn }}"
```

2. You can manually set `connector.druid.dsn` in your project's `.env` file or try pulling existing credentials locally using `rill env pull` if the project has already been deployed to Rill Cloud.

:::tip Getting DSN errors in dashboards after setting `.env`?

If you are facing issues related to DSN connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.

:::

## Connection String (DSN)

<img src='/img/build/connectors/olap-engines/druid/druid-dsn.png' class='rounded-gif' style={{width: '75%', display: 'block', margin: '0 auto'}}/>
<br />

Rill connects to Druid using the [HTTP API](https://druid.apache.org/docs/latest/api-reference/sql-api) and requires a connection string of the following format: `http://<user>:<password>@<host>:<port>/druid/v2/sql`. If `user` or `password` contain special characters, they should be URL encoded (i.e., `p@ssword` -> `p%40ssword`). This should be set in the `connector.druid.dsn` property in Rill.

As an example, this typically looks like:

```bash
connector.druid.dsn="https://user:password@localhost:8888/druid/v2/sql"
```

:::info Need help connecting to Druid?

If you would like to connect Rill to an existing Druid instance, please don't hesitate to [contact us](/contact). We'd love to help!

:::

## Setting the Default OLAP Connection

When connecting to Druid via the UI, the default OLAP connection will be automatically added to your rill.yaml. This will change the way the UI behaves, such as adding new data sources, as this is not supported with a Druid-backed Rill project.

```yaml
olap_connector: druid
```

:::note

For more information about available properties in `rill.yaml`, see our [project YAML](/reference/project-files/rill-yaml) documentation.

:::

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](/build/connectors/olap/multiple-olap) page.

:::

## Configuring Rill Cloud

When deploying a Druid-backed project to Rill Cloud, you have the following options to pass the appropriate connection string to Rill Cloud:
1. If you have followed the UI to create your Druid connector, the password or DSN should already exist in the .env file. During the deployment process, this `.env` file is automatically pushed with the deployment.
2. Use the `rill env configure` command to set `connector.druid.dsn` after deploying the project.
3. If `connector.druid.dsn` has already been set in your project `.env`, you can push and update these variables directly in your cloud deployment by using the `rill env push` command.

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

## Supported Versions

Rill supports connecting to Druid v28.0 or newer versions.

## Additional Notes

- At the moment, we do not support modeling with Druid. If this is something you're interested in, please [contact us](/contact).
- For dashboards powered by Druid, [measure definitions](/build/metrics-view/#measures) are required to follow standard [Druid SQL](https://druid.apache.org/docs/latest/querying/sql/) syntax.