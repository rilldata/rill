---
title: Environmental Variables and Credentials in Rill Cloud
sidebar_label: "Variables and Credentials"
sidebar_position: 50
---
<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

The credentials in a deployed Rill Cloud projects can be managed on the Settings page or via the CLI. If you have yet to deploy your credentials, please follow the steps in our [deploy credentials page](/developers/deploy/deploy-credentials). 

## Modifying Variables and Credentials via the Settings Page
Upon deployment via Rill Developer, if you have populated your .env file, the contents will be visible as seen below. If there are no environmental variables defined, please run `rill env push` from your local CLI and Rill will automatically push the credentials in your project's `.env` file to Rill Cloud. If you'd like to manually add the credentials, please see [our naming convention](#credentials-naming-schema) to get started. 

<img src = '/img/tutorials/admin/env-var-ui.png' class='rounded-gif' />
<br />


### Adding and Editing Environmental Variables / Importing a `.env` file
Once your environmental variables are added to Rill Cloud, they can be modified as needed.


<img src = '/img/manage/var-and-creds/add-variable.png' class='rounded-gif' />
<br />

:::tip Can't find the .env file?
By default, the hidden files will not be visible in the finder window. In order to view hidden files, you will need to enable "show hidden files".  
Keyboard shortcut: Command + Shift + .
:::

## Pushing and pulling credentials to / from Rill Cloud via the CLI

If you'd prefer to use the CLI to managed your credentials, this can be done by running the `rill env pull` to pull your deployed Rill Cloud project's variables locally, or `rill env push` to updated Rill Cloud project's variables.

:::tip Avoid committing sensitive information to Git

It's never a good idea to commit sensitive information to Git and goes against security best practices. Similar to credentials, if there are sensitive variables that you don't want to commit publicly to your `rill.yaml` configuration file (and thus potentially accessible by others), it's recommended to set them in your `.env` file directly and/or use `rill env set` via the CLI (and then optionally push / pull them as necessary).

:::

### `rill env push`

As a project admin, you can run `rill env push` from your local CLI that will update your Rill project with the contexts of your *`<RILL_PROJECT_HOME>/.env`* file.
- Rill Cloud will use the specified credentials and variables in this `.env` file for the deployed project.
- Other users will also be able to use `rill env pull` to retrieve these defined credentials for local use (with Rill Developer).

:::warning Overriding Cloud credentials

If a credential and/or variable has already been configured in Rill Cloud, Rill will warn you about overriding if you attempt to push a new value in your `.env` file. This is because overriding credentials can impact your deployed project and/or other users (if they pull these credentials locally).

:::

### `rill env pull`

For projects that have been deployed to Rill Cloud, an added benefit of our Rill Developer-Cloud architecture is that credentials that have been configured can be pulled locally for easier reuse (instead of having to manually reconfigure these credentials in Rill Developer). To do this, you can run `rill env pull` from your project's root directory to retrieve the latest credentials (after cloning the project's git repository to your local environment).

```bash
rill env pull
Updated .env file with cloud credentials from project "<Project_Name>".
```

:::info Overriding local credentials

Please note when you run `rill env pull`, Rill will *automatically override any existing credentials or variables* that have been configured in your project's `.env` file if there is a match in the key name. This may result in unexpected behavior if you are using different credentials locally.

:::

### Credentials Naming Schema

Connector credentials use a standardized naming convention. Generic credentials (shared across connectors like cloud providers) use standard names without a driver prefix, while driver-specific credentials use the `DRIVER_PROPERTY` format. Please see below for each source and its required properties. If you have any questions or need specifics, [contact us](/contact)!

:::note Legacy Naming Convention
Older projects may use the `connector.<connector_name>.<property>` syntax (e.g., `connector.druid.dsn`, `connector.clickhouse.dsn`). This format is still supported for backwards compatibility.
:::

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
| **Source Name** |             Property             | Environment Variable                |
| :-------------: | :------------------------------: | :---------------------------------- |
|     **GCS**     | `google_application_credentials` | `GOOGLE_APPLICATION_CREDENTIALS`    |
|                 |             `key_id`             | `GCP_ACCESS_KEY_ID`                 |
|                 |             `secret`             | `GCP_SECRET_ACCESS_KEY`             |
|   **AWS S3**    |       `aws_access_key_id`        | `AWS_ACCESS_KEY_ID`                 |
|                 |     `aws_secret_access_key`      | `AWS_SECRET_ACCESS_KEY`             |
|    **Azure**    |     `azure_storage_account`      | `AZURE_STORAGE_ACCOUNT`             |
|                 |       `azure_storage_key`        | `AZURE_STORAGE_KEY`                 |
|                 | `azure_storage_connection_string`| `AZURE_STORAGE_CONNECTION_STRING`   |
|                 |     `azure_storage_sas_token`    | `AZURE_STORAGE_SAS_TOKEN`           |
|  **Big Query**  | `google_application_credentials` | `GOOGLE_APPLICATION_CREDENTIALS`    |
|  **Snowflake**  |              `dsn`               | `SNOWFLAKE_DSN`                     |
|                 |           `password`             | `SNOWFLAKE_PASSWORD`                |
| **ClickHouse**  |              `host`              | `CLICKHOUSE_HOST`                   |
|                 |              `port`              | `CLICKHOUSE_PORT`                   |
|                 |            `username`            | `CLICKHOUSE_USERNAME`               |
|                 |            `password`            | `CLICKHOUSE_PASSWORD`               |
|                 |              `ssl`               | `CLICKHOUSE_SSL`                    |
|                 |            `database`            | `CLICKHOUSE_DATABASE`               |
|                 |              `dsn`               | `CLICKHOUSE_DSN`                    |

</div>