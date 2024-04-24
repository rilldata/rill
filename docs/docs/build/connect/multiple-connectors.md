---
title: Setting up multiple connectors of the same type
description: Setting up multiple connectors of the same type in one project
sidebar_label: Multiple Connectors
sidebar_position: 20
---

Sometimes, you will need to set up multiple connectors of the same type in one project with different connection strings (DSN) or configurations for sources. A common example would be that you have a need for multiple [Snowflake](/reference/connectors/snowflake.md) sources that point to different databases and schemas. Therefore, when [deploying your project to Rill Cloud](/deploy/existing-project/#deploy-to-rill-cloud), you will want to specify multiple `snowflake.connector.dsn` connection strings, one corresponding to each unique "connection" you desire (with different connection parameters).

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
kind: source
connector: "snowflake-a"
dsn: "user:password@account_identifier/DB_A/SCHEMA_A?warehouse=COMPUTE_WH&role=ACCOUNTADMIN"
sql: "select * from table_A"
```

**sourceB.yaml**:
```yaml
kind: source
connector: "snowflake-b"
dsn: "user:password@account_identifier/DB_B/SCHEMA_B?warehouse=COMPUTE_WH&role=ACCOUNTADMIN"
sql: "select * from table_B"
```

## Setting credentials for each connector when deploying to Rill Cloud

Finally, when deploying the project to Rill Cloud, you will want to follow the same steps to [set the credentials](/build/credentials/#setting-credentials-for-a-rill-cloud-project) for each connector.

If using `rill env configure`, you should be prompted to input the correct `connector.<connector_name>.dsn` connection strings.

![Inputting credentials for each connector](/img/build/connect/multiple-connectors/rill-env-configure.png)

Similarly, you can also configure your project's `.env` file to contain the correct connection string for each connector DSN:

```shell
connector.snowflake-a.dsn="<input_connectionA_dsn>"
connector.snowflake-b.dsn="<input_connectionB_dsn>"
```

Then, you can use `rill env push` and `rill env pull` as necessary to [push and pull your credentials](/build/credentials/#pushing-and-pulling-credentials-to--from-rill-cloud) respectively for a deployed project on Rill Cloud.