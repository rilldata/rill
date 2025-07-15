---
title: MotherDuck
description: Power Rill dashboards using MotherDuck
sidebar_label: MotherDuck
sidebar_position: 5
---



## Overview

[MotherDuck](https://motherduck.com/) is a cloud-based OLAP engine built on top of [DuckDB](https://duckdb.org/), providing the familiar DuckDB SQL experience with the scalability and collaboration features of the cloud. MotherDuck enables you to run analytical queries on large datasets, share data and queries, and leverage DuckDB’s performance with cloud storage and compute.

Rill supports connecting to MotherDuck as an OLAP engine to power dashboards built against [external tables](../../concepts/OLAP#external-olap-tables) as well as import the data from MotherDuck into Rill's embed DuckDB as a [source](/reference/connectors/motherduck). This is especially useful for teams who want the flexibility of DuckDB with the convenience of a managed, collaborative cloud backend.


## Configuring Rill Developer with MotherDuck

To use MotherDuck with Rill Developer, you will need a MotherDuck account and an access token. You can obtain a token from your [MotherDuck account dashboard](https://app.motherduck.com/settings/tokens).

Create a new file under the `connectors/` folder with the following contents:

```yaml
#connectors/motherduck.yaml
type: connector
driver: duckdb

path: "md:my_db"

init_sql: |
  INSTALL 'motherduck';
  LOAD 'motherduck';
  SET motherduck_token= {{ .env.motherduck_token }} 
```

1. You can create/edit the `.env` file manually in the project directory, if it doesn't already exist, and add [`connector.motherduck.token`](#motherduck-token).
2. If this project has already been deployed to Rill Cloud, you can try pulling existing credentials locally using `rill env pull`.
3. You can pass in `connector.motherduck.token` as a variable to `rill start` directly (e.g. `rill start --env connector.motherduck.token=...`).

:::tip Getting token/DSN errors in dashboards after setting `.env`?
If you are facing issues related to token connection errors in your dashboards even after setting the connection string via the project's `.env` file, try restarting Rill using the `rill start --reset` command.
:::

## MotherDuck Token

You can generate a personal access token from your [MotherDuck account settings](https://app.motherduck.com/settings/tokens). This token is required for authentication and should be kept secure.

Add the following to your `.env` file:

```bash
connector.motherduck.token="<your-motherduck-token>"
```

For more information about available DSN properties, see the [MotherDuck documentation](https://motherduck.com/docs/key-tasks/authenticating-and-connecting-to-motherduck/authenticating-to-motherduck/#authentication-using-an-access-token).

## Configuring Rill Cloud

When deploying a MotherDuck-backed project to Rill Cloud, you have the following options to pass the appropriate token to Rill Cloud:
1. If you have created the environmental variable file, `.env`, this will be deployed with the project.
2. For existing projects, you can use `rill env push` to update the token, if required. 
3. If you manually passed the connector when running `rill start`, you will need to use the `rill env configure` command to set `connector.motherduck.token` onto Rill Cloud, as well.


:::info
Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`, or pass the --project project_name flag. 
:::

## Setting the default OLAP connection
Creating a connection to a OLAP engine will automatically add the `olap_connector` property in your project's [rill.yaml](../project-files/rill-yaml.md) and change the default OLAP engine to MotherDuck. Once this is changed, you'll notice that some of the UI features are removed as we currently do not support modeling and direct source ingestion in MotherDuck.

```yaml
olap_connector: motherduck
```

:::info Interested in using multiple OLAP engines in the same project?
Please see our [Using Multiple OLAP Engines](multiple-olap.md) page.
:::

## Reading from multiple schemas

Rill supports reading from multiple schemas in MotherDuck from within the same project in Rill Developer and all accessible tables (given the permission set of the underlying user) should automatically be listed in the lower left-hand tab, which can then be used to [create dashboards](/build/dashboards/).

## Additional Notes

- At the moment, we do not officially support modeling with MotherDuck, but it is possible via YAML configurations.
- For dashboards powered by MotherDuck, [measure definitions](/build/metrics-view/metrics-view.md#measures) are required to follow standard [DuckDB SQL](https://duckdb.org/docs/sql/introduction) syntax.
- For more information, see the [MotherDuck documentation](https://docs.motherduck.com/).
