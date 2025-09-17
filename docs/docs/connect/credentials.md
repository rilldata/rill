---
title: Configure Local Credentials
sidebar_label: Configure Local Credentials
sidebar_position: 15
---

Rill requires credentials to connect to remote data sources such as private buckets (S3, GCS, Azure), data warehouses (Snowflake, BigQuery), OLAP engines (ClickHouse, Apache Druid), or other DuckDB sources (MotherDuck). Please refer to the appropriate [connector](/connect) and [OLAP engine](/connect/olap) page for instructions to configure credentials accordingly.

At a high level, configuring credentials and credential management in Rill can be broken down into three categories:
- Setting credentials for Rill Developer
- [Setting credentials for a Rill Cloud project](/deploy/deploy-credentials)
- [Pushing and pulling credentials to / from Rill Cloud](/manage/project-management/variables-and-credentials)

## Setting credentials for Rill Developer

When reading from a data source (or using a different OLAP engine), Rill will attempt to use existing credentials that have been configured on your machine.
1. Credentials that have been configured in your local environment via the CLI (for [AWS](/connect/data-source/s3#local-credentials) / [Azure](/connect/data-source/azure#rill-developer-local-credentials) / [Google Cloud](/connect/data-source/gcs#rill-developer-local-credentials))
2. Credentials that have been passed in directly through the connection string or DSN (typically for databases - see [Source YAML](/reference/project-files/sources) and [Connector YAML](/reference/project-files/connectors) for more details)
3. Credentials that have been passed in as a [variable](/connect/templating) when starting Rill Developer via `rill start --env key=value`
4. Credentials that have been specified in your *`<RILL_PROJECT_HOME>/.env`* file, see [credential naming schema](#credentials-naming-schema) for more information.

For more details, please refer to the corresponding [connector](/connect) or [OLAP engine](/connect/olap) page.

:::note Ensuring security of credentials in use

If you plan to deploy a project (to Rill Cloud), it is not recommended to pass in credentials directly through the local connection string or DSN as your credentials will then be checked in directly to your Git repository (and thus accessible by others). To ensure better security, credentials should be passed in as a variable / configured locally or specified in the project's local `.env` file (which is part of `.gitignore` and thus won't be included).

:::


## Variables

Project variables work exactly the same way as credentials and can be defined when starting rill via `--env key=value`, set in the .env file in the project directory, or defined in the rill.yaml.

```bash
#.env
variable: xyz
```

```yaml
#rill.yaml
env:
  variable: xyz
```

This variable will then be usable and referenceable for [templating](/connect/templating) purposes in the local instance of your project. 

:::info Fun Fact

Connector credentials are essentially a form of project variable, prefixed using the `connector.<connector_name>.<property>` syntax. For example, `connector.druid.dsn` and `connector.clickhouse.dsn` are both hard coded project variables (that happen to correspond to the [Druid](/connect/olap/druid) and [ClickHouse](/connect/olap/clickhouse) OLAP engines respectively).

:::

:::tip Avoid committing sensitive information to Git

It's never a good idea to commit sensitive information to Git and it goes against security best practices. Similar to credentials, if there are sensitive variables that you don't want to commit publicly to your `rill.yaml` configuration file (and thus potentially accessible by others), it's recommended to set them in your `.env` file directly and/or use `rill env set` via the CLI (and then optionally push / pull them as necessary).

:::

## Deploying to Rill Cloud 

If you have configured your credentials via the `.env` file this will be deployed with your project. If not, follow the steps to deploy then configure your credentials via the CLI running [`rill env configure`](/deploy/deploy-credentials#configure-environmental-variables-and-credentials-for-rill-cloud).

 

## Pulling Credentials and Variables from a Deployed Project on Rill Cloud

If you are making changes to an already deployed instance from Rill Cloud, it is possible to **pull** the credentials and variables from the Rill Cloud to your local instance of Rill Developer. If you've made any changes to the credentials, don't forget to run `rill env push` to push the variable changes to the project, or manually change these in the project's settings page.

### rill env pull

For projects that have been deployed to Rill Cloud, an added benefit of our Rill Developer-Cloud architecture is that credentials that have been configured can be pulled locally for easier reuse (instead of having to manually reconfigure these credentials in Rill Developer). To do this, you can run `rill env pull` from your project's root directory to retrieve the latest credentials (after cloning the project's git repository to your local environment).

![img](/img/build/credentials/rill-env-pull.png)

:::info Overriding local credentials

Please note when you run `rill env pull`, Rill will *automatically override any existing credentials or variables* that have been configured in your project's `.env` file if there is a match in the key name. This may result in unexpected behavior if you are using different credentials locally.

:::


### rill env push

As a project admin, you can either use `rill env configure` after deploying a project or `rill env push` to specify a particular set of credentials that your Rill Cloud project will use. If choosing the latter, you can update your *`<RILL_PROJECT_HOME>/.env`* file with the appropriate variables and credentials that are required. Alternatively, if this file has already been updated, you can run `rill env push` from your project's root directory.
- Rill Cloud will use the specified credentials and variables in this `.env` file for the deployed project.
- Other users will also be able to use `rill env pull` to retrieve these defined credentials for local use (with Rill Developer).

:::warning Overriding Cloud credentials

If a credential and/or variable has already been configured in Rill Cloud, Rill will warn you about overriding if you attempt to push a new value in your `.env` file. This is because overriding credentials can impact your deployed project and/or other users (if they pull these credentials locally).
![img](/img/build/credentials/rill-env-push.png)


:::

### Cloning an Existing Project from Rill Cloud

If you cloned the project using `rill project clone <project-name>` and are an admin of that project, the credentials will be pulled automatically. Note that there are some limitations with monorepos where credentials may not be pulled correctly. In those cases, credentials are also pulled when running `rill start`, assuming you have already authenticated via the CLI with `rill login`.


### Credentials Naming Schema 

Connector credentials are essentially a form of project variable, prefixed using the `connector.<connector_name>.<property>` syntax. For example, `connector.druid.dsn` and `connector.clickhouse.dsn` are both hard coded project variables (that happen to correspond to the [Druid](/connect/olap/druid) and [ClickHouse](/connect/olap/clickhouse) OLAP engines respectively). Please see below for each source and its required properties. If you have any questions or need specifics, [contact us](/contact)!
