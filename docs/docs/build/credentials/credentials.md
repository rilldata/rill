---
title: Configure connector credentials
sidebar_label: Configure Local Credentials
sidebar_position: 00
---

Rill requires credentials to connect to remote data sources such as private buckets (S3, GCS, Azure), data warehouses (Snowflake, BigQuery), OLAP engines (ClickHouse, Apache Druid) or other DuckDB sources (MotherDuck). Please refer to the appropriate [connector](../../reference/connectors/connectors.md) and [OLAP engine](../../reference/olap-engines/olap-engines.md) page for instructions to configure credentials accordingly.

At a high level, configuring credentials and credentials management in Rill can be broken down into three categories:
- Setting credentials for Rill Developer
- [Setting credentials for a Rill Cloud project](/deploy/deploy-credentials)
- [Pushing and pulling credentials to / from Rill Cloud](/manage/project-management/variables-and-credentials)

## Setting credentials for Rill Developer

When reading from a source (or using a different OLAP engine), Rill will attempt to use existing credentials that have been configured on your machine.
1. Credentials that have been configured in your local environment via the CLI (for [AWS](../../reference/connectors/s3.md#local-credentials) / [Azure](../../reference/connectors/azure.md#local-credentials) / [Google Cloud](../../reference/connectors/gcs#rill-developer-local-credentials))
2. Credentials that have been passed in directly through the connection string or DSN (typically for databases - see [Source YAML](../../reference/project-files/sources.md) and [Connector YAML](../../reference/project-files/connectors.md) for more details)
3. Credentials that have been passed in as a [variable](../../deploy/templating.md) when starting Rill Developer via `rill start --env key=value`
4. Credentials that have been specified in your *`<RILL_PROJECT_HOME>/.env`* file, see  [credential naming schema](#credentials-naming-schema) for more information.

For more details, please refer to the corresponding [connector](../../reference/connectors/connectors.md) or [OLAP engine](../../reference/olap-engines/olap-engines.md) page.

:::note Ensuring security of credentials in use

If you plan to deploy a project (to Rill Cloud), it is not recommended to pass in credentials directly through the local connection string or DSN as your credentials will then be checked directly into your Git repository (and thus accessible by others). To ensure better security, credentials should be passed in as a variable / configured locally or specified in the project's local `.env` file (which is part of `.gitignore` and thus won't be included).

:::


## Variables

Project variables work exactly the same way as credentials and can be defined when starting rill via `--env key=value` or set in the .env file in the project directory.

```bash
variable=xyz
```

This variable will then be usable and referenceable for [templating](../../deploy/templating.md) purposes in the local instance of your project. 

:::info Fun Fact

Connector credentials are essentially a form of project variable, prefixed using the `connector.<connector_name>.<property>` syntax. For example, `connector.druid.dsn` and `connector.clickhouse.dsn` are both hardcoded project variables (that happen to correspond to the [Druid](/reference/olap-engines/druid.md) and [ClickHouse](/reference/olap-engines/clickhouse.md) OLAP engines respectively).

:::

:::tip Avoid committing sensitive information to Git

It's never a good idea to commit sensitive information to Git and goes against security best practices. Similar to credentials, if there are sensitive variables that you don't want to commit publicly to your `rill.yaml` configuration file (and thus potentially accessible by others), it's recommended to set them in your `.env` file directly and/or use `rill env set` via the CLI (and then optionally push / pull them as necessary).

:::

## Deploying to Rill Cloud 

Please see our [deploy credentials page](/deploy/deploy-credentials#configure-environmental-variables-and-credentials-for-rill-cloud) to configure your credentials on Rill Cloud. If you have configured your credentials via the `.env` file this will be deployed with your project. If not, follow the steps to deploy then configure your credentials via the CLI running `rill env configure`.



## Pulling Credentials and Variables from a Deployed Project on Rill Cloud

If you are making changes to an already deployed instance from Rill Cloud, it is possible to **pull** the credentials and variables from the Rill Cloud to your local instance of Rill Developer. If you've made any changes to the credentials, don't forget to run `rill env push` to push the variable changes to the project, or manually change these in the project's setting page.

### rill env pull

For projects that have been deployed to Rill Cloud, an added benefit of our Rill Developer-Cloud architecture is that credentials that have been configured can be pulled locally for easier reuse (instead of having to manually reconfigure these credentials in Rill Developer). To do this, you can run `rill env pull` from your project's root directory to retrieve the latest credentials (after cloning the project's git repository to your local environment).

![Pulling credentials from Rill Cloud](/img/build/credentials/rill-env-pull.png)

:::info Overriding local credentials

Please note when you run `rill env pull`, Rill will *automatically override any existing credentials or variables* that have been configured in your project's `.env` file if there is a match in the key name. This may result in unexpected behavior if you are using different credentials locally.

:::


### rill env push

As a project admin, you can either use `rill env configure` after deploying a project or `rill env push` to specify a particular set of credentials that your Rill Cloud project will use. If choosing the latter, you can update your *`<RILL_PROJECT_HOME>/.env`* file with the appropriate variables and credentials that are required. Alternatively, if this file has already been updated, you can run `rill env push` from your project's root directory.
- Rill Cloud will use the specified credentials and variables in this `.env` file for the deployed project.
- Other users will also be able to use `rill env pull` to retrieve these defined credentials for local use (with Rill Developer).

:::warning Overriding Cloud credentials

If a credential and/or variable has already been configured in Rill Cloud, Rill will warn you about overriding if you attempt to push a new value in your `.env` file. This is because overriding credentials can impact your deployed project and/or other users (if they pull these credentials locally).
![Pushing credentials that already exist to Rill Cloud](/img/build/credentials/rill-env-push.png)

:::


### Credentials Naming Schema 

Connector credentials are essentially a form of project variable, prefixed using the `connector.<connector_name>.<property>` syntax. For example, `connector.druid.dsn` and `connector.clickhouse.dsn` are both hardcoded project variables (that happen to correspond to the [Druid](/reference/olap-engines/druid.md) and [ClickHouse](/reference/olap-engines/clickhouse.md) OLAP engines respectively). Please see below for each source and its required properties. If you have any questions or need specifics, [contact us](../../contact)! 


<div
    style={{
    width: '100%',
    margin: 'auto',
    padding: '20px',
    textAlign: 'center', 
    display: 'flex', 
    justifyContent: 'center',
    alignItems: 'center'
    }}
>
|           **Source Name**   |        Property             |      Example         |
| :-----------------------: | :-------------------------:  | :------------------- |
|       **GCS**                |`GOOGLE_APPLICATION_CREDENTIALS`| `connector.gcs.google_application_credentials` |
|                          |`GCS_BUCKET_NAME`| `connector.gcs.gcs_bucket_name` |
| **AWS S3**                  | `AWS_ACCESS_KEY_ID`         | `connector.s3.aws_access_key_id` |
|                          | `AWS_SECRET_ACCESS_KEY`     |`connector.s3.aws_secret_access_key` |
|       **Azure**              |`AZURE_STORAGE_ACCOUNT`|`connector.azure.azure_storage_account`|
|                          |`AZURE_STORAGE_KEY`|`connector.azure.azure_storage_key`|
|                          |`AZURE_CLIENT_ID`|`connector.azure.azure_client_id`|
|                          |`AZURE_CLIENT_SECRET`|`connector.azure.azure_client_secret`|
|                          |`AZURE_TENANT_ID`|`connector.azure.azure_tenant_id`|
| **Big Query**               | `GOOGLE_APPLICATION_CREDENTIALS` |`connector.bigquery.google_application_credentials` |
|     **Snowflake**            |`DSN`|`connector.snowflake.dsn`|
|     **ClickHouse**           |`HOST`|`connector.clickhouse.host `|
|                          |`PORT`|`connector.clickhouse.port `|
|                          |`USERNAME`|`connector.clickhouse.username `|
|                          |`PASSWORD`|`connector.clickhouse.password `|
|                          |`SSL`|`connector.clickhouse.ssl `|
|                          |`DATABASE`|`connector.clickhouse.database `|
...

</div>