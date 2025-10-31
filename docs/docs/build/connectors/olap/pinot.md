---
title: Pinot
description: Power Rill dashboards using Pinot
sidebar_label: Pinot
sidebar_position: 20
---

[Apache Pinot](https://docs.pinot.apache.org/) is a real-time distributed OLAP datastore purpose-built for low-latency, high-throughput analytics, and is perfect for user-facing analytical workloads.

Rill supports connecting to an existing Pinot cluster via a "live connector" and using it as an OLAP engine  built against [external tables](/build/connectors/olap#external-olap-tables) to power Rill dashboards. This is particularly useful when working with extremely large datasets (hundreds of GBs or even TB+ in size).


## Configuring Rill Developer with Pinot

When using Rill for local development, there are a few options to configure Rill to enable Pinot as an OLAP engine:

1. Connect to an OLAP engine via Add Data. This will automatically create the `pinot.yaml` file in your `connectors` directory and populate the `.env` file with `connector.pinot.password` or `connector.pinot.dsn` depending on which you select in the UI.

    For more information on supported parameters, see our [Pinot connector YAML reference docs](/reference/project-files/connectors#pinot).
    ```yaml
    type: connector
    driver: pinot

    dsn: "{{ .env.connector.pinot.dsn }}"
    ```

1. You can set `connector.pinot.dsn` in your project's `.env` file or try pulling existing credentials locally using `rill env pull` if the project has already been deployed to Rill Cloud.

:::tip Getting DSN errors in dashboards after setting `.env`?

If you are facing issues related to DSN connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.

:::

## Connection String (DSN)

Rill connects to Pinot using the [Pinot Golang Client](https://docs.pinot.apache.org/users/clients/golang) and requires a connection string of the following format: `http://<user>:<password>@<broker_host>:<port>?controller=<controller_host>:<port>`. If `user` or `password` contain special characters, they should be URL encoded (i.e., `p@ssword` -> `p%40ssword`). This should be set in the `connector.pinot.dsn` property in Rill.

<img src='/img/build/connectors/olap-engines/pinot/pinot-dsn.png' class='rounded-gif' style={{width: '75%', display: 'block', margin: '0 auto'}}/>

As an example, this typically looks like:

```bash
connector.pinot.dsn="http(s)://username:password@localhost:8000?controller=localhost:9000"
```

:::info Need help connecting to Pinot?

If you would like to connect Rill to an existing Pinot instance, please don't hesitate to [contact us](/contact). We'd love to help!

:::

## Setting the Default OLAP Connection

You'll also need to update the `olap_connector` property in your project's `rill.yaml` to change the default OLAP engine to Pinot:

```yaml
olap_connector: pinot
```

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](/build/connectors/olap/multiple-olap) page.

:::

## Configuring Rill Cloud

When deploying a Pinot-backed project to Rill Cloud, you have the following options to pass the appropriate connection string to Rill Cloud:
1. If you have followed the UI to create your Pinot connector, the password or DSN should already exist in the .env file. During the deployment process, this `.env` file is automatically pushed with the deployment.
2. Use the `rill env configure` command to set `connector.pinot.dsn` after deploying the project.
3. If `connector.pinot.dsn` has already been set in your project `.env`, you can push and update these variables directly in your cloud deployment by using the `rill env push` command.

## Support for Multi-Valued Dimensions

Multi-valued dimensions need to be defined in the dashboard YAML as expressions using the `arrayToMv` function. For example, if `RandomAirports` is a multi-valued column in a Pinot table, then the dimension definition will look like:

```yaml
- display_name: RandomAirports
  expression: arrayToMv(RandomAirports)
  name: RandomAirports
  description: "Random Airports"
```

Refer to the [Dashboard YAML](/reference/project-files/explore-dashboards) reference page for all dimension properties in detail.

:::note

Pinot does not support the unnest function, so don't set the `unnest` property to true in the dimension definition of the dashboard YAML.

:::

## Additional Notes

- At the moment, we do not support modeling with Pinot. If this is something you're interested in, please [contact us](/contact).
- For dashboards powered by Pinot, [measure definitions](/build/metrics-view/#measures) are required to follow [Pinot SQL](https://docs.pinot.apache.org/users/user-guide-query/querying-pinot) syntax.
