---
title: Snowflake
description: Connect to data in Snowflake
sidebar_label: Snowflake
sidebar_position: 75
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

:::info Deprecation of password authentication

Snowflake has issued a [deprecation notice](https://www.snowflake.com/en/blog/blocking-single-factor-password-authentification/) for single-factor password authentication. Rill supports and recommends using private key authentication to avoid any disruption of your service.

:::

## Overview

[Snowflake](https://docs.snowflake.com/en/user-guide-intro) is a cloud-based data platform designed to facilitate data warehousing, data lakes, data engineering, data science, data application development, and data sharing. It separates compute and storage, enabling users to scale up or down instantly without downtime, providing a cost-effective solution for data management. With its unique architecture and support for multi-cloud environments, including AWS, Azure, and Google Cloud Platform, Snowflake offers seamless data integration, secure data sharing across organizations, and real-time access to data insights, making it a common choice to power many business intelligence applications and use cases. You can connect to and read from Snowflake data warehouses using the [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake).

## Authentication Methods

To connect to Snowflake, you need to provide authentication credentials. Rill supports two methods:

1. **Use Keypair Authentication** (recommended - required for production)
2. **Use Password Authentication** (deprecated by Snowflake)

:::tip Authentication Methods
Snowflake has deprecated single-factor password authentication. We strongly recommend using keypair authentication to ensure uninterrupted service.
:::

## Using the Add Data UI

When you add a Snowflake data model through the Rill UI, the process follows two steps:

1. **Configure Authentication** - Set up your Snowflake connector with credentials via connection string
2. **Configure Data Model** - Define which database, schema, table, or query to execute

This two-step flow ensures your credentials are securely stored in the connector configuration, while your data model references remain clean and portable.

---

## Method 1: Keypair Authentication (Recommended)

Keypair authentication provides enhanced security and is the recommended method for connecting to Snowflake. This method works for both local development and Rill Cloud deployments.

### Connection String Format

Use the following syntax when defining a connection string using a private key:

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_url_safe>
```

See the [Appendix](#using-keypair-authentication) for detailed instructions on generating and formatting your private key.

<img src='/img/build/connectors/data-sources/snowflake_conn_strings.png' class='rounded-gif' />
<br />

:::info Finding the Snowflake account identifier

To determine your [Snowflake account identifier](https://docs.snowflake.com/en/user-guide/admin-account-identifier), one easy way is to check your Snowflake account URL. The account identifier to use in your connection string should be everything before `.snowflakecomputing.com`!

:::

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Snowflake** as the data source type
3. In the authentication step:
   - Enter your connection string with keypair authentication
4. In the data model configuration step:
   - Enter your SQL query
   - Configure other model settings as needed
5. Click **Create** to finalize

The UI will automatically create both the connector file and model file for you.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/my_snowflake.yaml`:

```yaml
type: connector
driver: snowflake

dsn: "{{ .env.connector.snowflake.dsn }}"
```

**Step 2: Create model configuration**

Create `models/my_snowflake_data.yaml`:

```yaml
type: model
connector: my_snowflake

sql: SELECT * FROM my_database.my_schema.my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.snowflake.dsn=myuser@myaccount/mydb/myschema?warehouse=mywh&role=myrole&authenticator=SNOWFLAKE_JWT&privateKey=<your_url_safe_private_key>
```

:::tip Did you know?
If this project has already been deployed to Rill Cloud and credentials have been set for this connector, you can use `rill env pull` to [pull these cloud credentials](/build/connectors/credentials#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.
:::

---

## Method 2: Password Authentication (Deprecated)

:::warning Deprecated by Snowflake
Snowflake has deprecated single-factor password authentication. This method may stop working in the future. We strongly recommend migrating to keypair authentication.
:::

### Connection String Format

```sql
<username>:<password>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>
```

### Manual Configuration

**Step 1: Create connector configuration**

Create `connectors/my_snowflake.yaml`:

```yaml
type: connector
driver: snowflake

dsn: "{{ .env.connector.snowflake.dsn }}"
```

**Step 2: Create model configuration**

Create `models/my_snowflake_data.yaml`:

```yaml
type: model
connector: my_snowflake

sql: SELECT * FROM my_database.my_schema.my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.snowflake.dsn=myuser:mypassword@myaccount/mydb/myschema?warehouse=mywh&role=myrole
```

---

## Using Snowflake Data in Models

Once your connector is configured, you can reference Snowflake tables and run queries in your model configurations.

### Basic Example

```yaml
type: model
connector: my_snowflake

sql: SELECT * FROM my_database.my_schema.my_table

refresh:
  cron: "0 */6 * * *"
```

### Custom SQL Query

```yaml
type: model
connector: my_snowflake

sql: |
  SELECT
    DATE_TRUNC('day', event_time) as event_date,
    event_type,
    COUNT(*) as event_count,
    SUM(revenue) as total_revenue
  FROM analytics.events.user_actions
  WHERE event_time >= DATEADD(day, -30, CURRENT_DATE())
  GROUP BY 1, 2

refresh:
  cron: "0 */6 * * *"
```

---

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

---

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide Snowflake credentials via the connection string as a source configuration `dsn` field used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#snowflake) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:
```
rill env push
```

:::tip Did you know?
If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.
:::

---

## Appendix

### Using Keypair Authentication

You can use keypair authentication for enhanced security when connecting to Snowflake as an alternative to password-based authentication, which Snowflake has deprecated. Per the [Snowflake Go Driver](https://github.com/snowflakedb/gosnowflake) specifications, this requires the following changes to the dsn:
- Remove the password
- Add `authenticator=SNOWFLAKE_JWT`
- Add `privateKey=<privateKey_url_safe>`

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_url_safe>
```

#### Generate a Private Key

Please refer to the [Snowflake documentation](https://docs.snowflake.com/en/user-guide/key-pair-auth) on how to configure an unencrypted private key to use in Rill.

The Snowflake Go Driver only supports **unencrypted PKCS#8 keys**. Make sure to include the `-nocrypt` flag, as encrypted keys are not supported. You can generate one using:

```bash
# Generate a 2048-bit unencrypted PKCS#8 private key
openssl genrsa 2048 | openssl pkcs8 -topk8 -inform PEM -out rsa_key.p8 -nocrypt
```

#### Convert the Private Key to a URL-Safe Format for the DSN

After generating the private key, you need to convert it into a URL-safe Base64 format for use in the Snowflake DSN. Run the following command:

```bash
# Convert URL safe format for DSN
cat rsa_key.p8 | grep -v "\----" | tr -d '\n' | tr '+/' '-_'
```

> Note: When copying the output, do not include the trailing % character that may appear in your terminal.

:::info Check your OS version

Depending on your OS version, above commands may differ slightly. Please check your OS reference manual for the correct syntax.

:::


:::tip Best Practices

If using keypair authentication, consider rotating your public key regularly to ensure compliance with security and governance best practices.

:::
