---
title: Configure Local Credentials
sidebar_label: Configure Local Credentials
sidebar_position: 15
---

Rill requires credentials to connect to remote data sources such as private buckets (S3, GCS, Azure), data warehouses (Snowflake, BigQuery), OLAP engines (ClickHouse, Apache Druid), or other DuckDB sources (MotherDuck). Please refer to the appropriate [connector](/build/connectors) and [OLAP engine](/build/connectors/olap) page for instructions to configure credentials accordingly.

At a high level, configuring credentials and credential management in Rill can be broken down into three categories:
- Setting credentials for Rill Developer
- [Setting credentials for a Rill Cloud project](/deploy/deploy-credentials)
- [Pushing and pulling credentials to / from Rill Cloud](/manage/project-management/variables-and-credentials)

## Setting credentials for Rill Developer


When reading from a data source (or using a different OLAP engine), Rill will attempt to use credentials in the following order of priority:

:::warning **Highly Recommended: Use .env for credentials**

While Rill **can** infer credentials from your local environment (AWS CLI, Azure CLI, Google Cloud CLI), **we HIGHLY recommend explicitly configuring credentials in your `.env` file** for better security, reliability, and portability. Environment-inferred credentials may vary across different setups and may not work consistently across different environments, team members, or when deploying to Rill Cloud.

:::

1. **Credentials referenced in connection strings or DSN within YAML files (RECOMMENDED)** - The UI creates YAML configurations that reference credentials from your `.env` file using templating (see [Connector YAML](/reference/project-files/connectors) for more details)
2. **Credentials passed in as variables** - When starting Rill Developer via `rill start --env key=value` (see [templating](/build/connectors/templating) for more details)
3. **Credentials configured via CLI** - For [AWS](/build/connectors/data-source/s3#local-aws-credentials-local-development-only) / [Azure](/build/connectors/data-source/azure#azure-cli-authentication-local-development-only) / [Google Cloud](/build/connectors/data-source/gcs#method-3-local-google-cloud-cli-credentials) - **NOT RECOMMENDED for production use**

For more details, please refer to the corresponding [connector](/build/connectors) or [OLAP engine](/build/connectors/olap) page.

:::note Ensuring security of credentials in use

If you plan to deploy a project (to Rill Cloud), it is not recommended to pass in credentials directly through the local connection string or DSN as your credentials will then be checked in directly to your Git repository (and thus accessible by others). To ensure better security, credentials should be passed in as a variable, configured locally, or specified in the project's local `.env` file (which is part of `.gitignore` and thus won't be included).

:::

## Variables

Project variables work exactly the same way as credentials and can be defined when starting rill via `--env key=value`, set in the `.env` file in the project directory, or defined in the rill.yaml.

### What is a `.env` file?

A `.env` file is a plain text file that stores environment variables and credentials for your Rill project. It's located in your project's root directory and follows a simple `key=value` format.

The `.env` file serves several important purposes:

- **Security**: Keeps sensitive credentials out of your codebase and Git repository (`.env` files are automatically ignored by `.gitignore`)
- **Consistency**: Provides a standardized way to manage credentials across different environments (local development, staging, production)
- **Integration**: Works seamlessly with Rill's templating system, allowing YAML files to reference credentials using `{{ env.VARIABLE_NAME }}` syntax

Example `.env` file:
```bash
# AWS S3 credentials
connector.s3.access_key_id=AKIAIOSFODNN7EXAMPLE
connector.s3.secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# Google Cloud credentials
connector.gcs.credentials_json={"type":"service_account","project_id":"my-project"}

# Database connection
connector.postgres.dsn=postgres://username:password@localhost:5432/mydb

# Custom variables
my_custom_variable=some_value
```
When creating any connector in Rill via the UI, these will be **automatically generated** in the `.env` file.

Additional variables can then be usable and referenceable for [templating](/build/connectors/templating) purposes in the local instance of your project. 

### Credentials Naming Schema 

Connector credentials are essentially a form of project variable, prefixed using the `connector.<connector_name>.<property>` syntax. For example, `connector.druid.dsn` and `connector.clickhouse.dsn` are both hard-coded project variables (that happen to correspond to the [Druid](/build/connectors/olap/druid) and [ClickHouse](/build/connectors/olap/clickhouse) OLAP engines respectively). Please see below for each source and its required properties. If you have any questions or need specifics, [contact us](/contact)!

:::tip Avoid committing sensitive information to Git

It's never a good idea to commit sensitive information to Git and it goes against security best practices. Similar to credentials, if there are sensitive variables that you don't want to commit publicly to your `rill.yaml` configuration file (and thus potentially accessible by others), it's recommended to set them in your `.env` file directly and/or use `rill env set` via the CLI (and then optionally push / pull them as necessary).

:::

## Deploying to Rill Cloud 

If you have configured your credentials via the `.env` file this will be deployed with your project. 

If not, follow the steps to deploy then configure your credentials via the CLI running [`rill env configure`](/deploy/deploy-credentials#configure-environmental-variables-and-credentials-for-rill-cloud).

## Cloning an Existing Project from Rill Cloud

If you cloned the project using `rill project clone <project-name>` and are an admin of that project, the credentials will be pulled automatically. Note that there are some limitations with monorepos where credentials may not be pulled correctly. In those cases, credentials are also pulled when running `rill start`, assuming you have already authenticated via the CLI with `rill login`.

For a detailed guide, see our [clone a project guide](/guides/clone-a-project).
 
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