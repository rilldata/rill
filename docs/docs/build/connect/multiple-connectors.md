---
title: Setting up multiple connectors of the same type
description: Setting up multiple connectors of the same type in one project
sidebar_label: Multiple Connectors
sidebar_position: 20
---

Sometimes, you will need to set up multiple connectors of the same type in one project with different connection strings (DSN) or configurations for sources. A common example would be that you have a need for multiple [Snowflake](/reference/connectors/snowflake.md) sources that point to different databases and schemas. Therefore, when [deploying your project to Rill Cloud](/deploy/deploy-dashboard/#deploying-a-project-from-rill-developer), you will want to specify multiple `snowflake.connector.dsn` connection strings, one corresponding to each unique "connection" you desire (with different connection parameters).

## Defining multiple connectors in `rill.yaml`

By default, Rill will infer the connection type when creating a source. However, in the case that multiple unique connectors of the same type (with different connection strings) are needed within the context of the same Rill project, we will need to first explicitly define each connector within the project's [rill.yaml](/reference/project-files/rill-yaml) file.

This can be done by specifying the connector type and name under the `connectors` property. For example, in the following `rill.yaml` file, we are defining an unique `snowflake-a` and `snowflake-b` connector in this project (both of `snowflake` type). 

```yaml
connectors:
- type: snowflake
  name: snowflake-a
- type: snowflake
  name: snowflake-b
```

:::info Naming your connectors

Any connectors you define explicitly in your `rill.yaml` file can be named however you want (with the appropriate type). This name, _however_, will then need to be used with the `connector` property in the corresponding [source.yaml](/reference/project-files/sources) definition.

:::

## Updating your `source.yaml` to use the defined connector

For <u>each</u> source that's using one of these connectors, make sure to update the [source.yaml](/reference/project-files/sources) definition accordingly so that the `connector` property specified the correct connector by name (previous section). 

For example, let's say we had `sourceA` and `sourceB` defined that point to different databases and schemas (in Snowflake).

**sourceA.yaml**:
```yaml
type: source
connector: "snowflake-a"
dsn: "user:password@account_identifier/DB_A/SCHEMA_A?warehouse=COMPUTE_WH&role=ACCOUNTADMIN"
sql: "select * from table_A"
```

**sourceB.yaml**:
```yaml
type: source
connector: "snowflake-b"
dsn: "user:password@account_identifier/DB_B/SCHEMA_B?warehouse=COMPUTE_WH&role=ACCOUNTADMIN"
sql: "select * from table_B"
```

## Setting credentials for each connector when deploying to Rill Cloud

Credentials that are defined in a project's `.env` file and defined in a `connector_name.yaml` will automatically be deployed with the project. 
If you need to make changes to the DSN after deployment, you can [set the credentials via the Rill Cloud UI](/deploy/deploy-credentials#configure-environmental-variables-and-credentials-for-rill-cloud) for each connector or by running  `rill env configure`. You will be prompted to input the correct `connector.<connector_name>.dsn` connection strings.

![Inputting credentials for each connector](/img/build/connect/multiple-connectors/rill-env-configure.png)

Or, you can also configure your project's `.env` file manually to contain the correct connection string for each connector DSN, and run `rill env push` to [push and pull your credentials](/build/credentials/#pulling-credentials-and-variables-from-a-deployed-project-on-rill-cloud) to Rill Cloud.

```shell
connector.snowflake-a.dsn="<input_connectionA_dsn>"
connector.snowflake-b.dsn="<input_connectionB_dsn>"
```
