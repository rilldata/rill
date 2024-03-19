---
title: Configure connector credentials
sidebar_label: Configure Credentials
sidebar_position: 00
---

Rill requires credentials to connect to remote data sources such as private buckets (S3, GCS, Azure), data warehouses (Snowflake, BigQuery), OLAP engines (ClickHouse, Apache Druid) or other DuckDB sources (MotherDuck). Please refer to the appropriate [connector](../../reference/connectors/connectors.md) and [OLAP engine](../../reference/olap-engines/olap-engines.md) page for instructions to configure credentials accordingly.

At a high level, configuring credentials and credentials management in Rill can be broken down into three categories:
- Setting credentials for Rill Developer
- Setting credentials for a Rill Cloud project
- Pushing and pulling credentials to / from Rill Cloud

## Setting credentials for Rill Developer

When reading from a source (or using a different OLAP engine), Rill will attempt to use existing credentials that have been configured on your machine.
1. Credentials that have been configured in your local environment via the CLI (for [AWS](../../reference/connectors/s3.md#local-credentials) / [Azure](../../reference/connectors/azure.md#local-credentials) / [Google Cloud](../../reference/connectors/gcs.md#local-credentials))
2. Credentials that have been passed in directly through the connection string or DSN (typically for databases - see [Source YAML](../../reference/project-files/sources.md) for more details)
3. Credentials that have been passed in as a [variable](../../deploy/templating.md) when starting Rill Developer via `rill start --var key=value`
4. Credentials that have been specified in your *`<RILL_PROJECT_HOME>/.env`* file

For more details, please refer to the corresponding [connector](../../reference/connectors/connectors.md) or [OLAP engine](../../reference/olap-engines/olap-engines.md) page.

:::note Ensuring security of credentials in use

If you plan to deploy a project (to Rill Cloud), it is not recommended to pass in credentials directly through the local connection string or DSN as your credentials will then be checked directly into your Git repository (and thus accessible by others). To ensure better security, credentials should be passed in as a variable / configured locally or specified in the project's local `.env` file (which is part of `.gitignore` and thus won't be included).

:::


## Setting credentials for a Rill Cloud project

When deploying a project, Rill Cloud will need appropriate credentials passed to it that can be used to read from a source (or to be used to establish a connection with the OLAP engine). This can be done either by:
1. Using [rill env configure](../../reference/cli/env/configure.md) to pass in connector level credentials (make sure to run this CLI command from the root directory of your Rill project)
2. Specifying all corresponding source level and OLAP engine credentials in the project's `.env` file and "pushing it" to Rill Cloud ([see below](#pushing-and-pulling-credentials-to--from-rill-cloud))

For more details, please refer to the corresponding [connector](../../reference/connectors/connectors.md) or [OLAP engine](../../reference/olap-engines/olap-engines.md) page.

:::tip Best practice

For projects that you deploy to production, it is recommended to use service accounts when possible instead of individual user credentials.


:::

## Pushing and pulling credentials to / from Rill Cloud

If you have a project deployed to Rill Cloud, Rill provides the ability to **sync** the credentials between your local instance of Rill Developer and Rill cloud. This provides the ability to quickly reuse existing credentials, if configured, instead of having to manually input credentials each time. This can be accomplished by leveraging the `rill env push` and `rill env pull` CLI commands respectively.

### rill env push

As a project admin, you can either use `rill env configure` after deploying a project or `rill env push` to specify a particular set of credentials that your Rill Cloud project will use. If choosing the latter, you can update your *`<RILL_PROJECT_HOME>/.env`* file with the appropriate variables and credentials that are required. Alternatively, if this file has already been updated, you can run `rill env push` from your project's root directory.
- Rill Cloud will use the specified credentials and variables in this `.env` file for the deployed project.
- Other users will also be able to use `rill env pull` to retrieve these defined credentials for local use (with Rill Developer).

:::warning Overriding Cloud credentials

If a credential and/or variable has already been configured in Rill Cloud, Rill will warn you about overriding if you attempt to push a new value in your `.env` file. This is because overriding credentials can impact your deployed project and/or other users (if they pull these credentials locally).
![Pushing credentials that already exist to Rill Cloud](/img/build/credentials/rill-env-push.png)

:::

### rill env pull

For projects that have been deployed to Rill Cloud, an added benefit of our Rill Developer-Cloud architecture is that credentials that have been configured can be pulled locally for easier reuse (instead of having to manually reconfigure these credentials in Rill Developer). To do this, you can run `rill env pull` from your project's root directory to retrieve the latest credentials (after cloning the project's git repository to your local environment).

![Pulling credentials from Rill Cloud](/img/build/credentials/rill-env-pull.png)

:::info Overriding local credentials

Please note when you run `rill env pull`, Rill will *automatically override any existing credentials or variables* that have been configured in your project's `.env` file if there is a match in the key name. This may result in unexpected behavior if you are using different credentials locally.

:::

### variables

Project variables work exactly the same way as credentials and can be pushed to Cloud / pulled locally using the same `rill env push` and `rill env pull` commands specifically. To do this, if you're in Rill Developer and want to set a variable through your `<RILL_PROJECT_HOME>/.env` file (and save):

```bash
variable=xyz
```

This variable should then be usable and referenceable for [templating](../../deploy/templating.md) purposes in the local instance of your project. Then, if you want to push these variable definitions to your deployed Cloud project, you can use `rill env push`. Similarly, if these variables had already been set in Rill Cloud for your project, you can use `rill env pull` to clone these variables locally (in your `.env` file).

:::info Fun Fact

Connector credentials are essentially a form of project variable, prefixed using the `connector.<connector_name>.<property>` syntax. For example, `connector.druid.dsn` and `connector.clickhouse.dsn` are both hardcoded project variables (that happen to correspond to the [Druid](/reference/olap-engines/druid.md) and [ClickHouse](/reference/olap-engines/clickhouse.md) OLAP engines respectively).

:::

:::tip Avoid committing sensitive information to Git

It's never a good idea to commit sensitive information to Git and goes against security best practices. Similar to credentials, if there are sensitive variables that you don't want to commit publicly to your `rill.yaml` configuration file (and thus potentially accessible by others), it's recommended to set them in your `.env` file directly and/or use `rill env set` via the CLI (and then optionally push / pull them as necessary).

:::


