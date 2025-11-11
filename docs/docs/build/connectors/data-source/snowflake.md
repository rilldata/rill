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


## Connect to Snowflake

Create a connector with your credentials to connect to Snowflake. Here's an example connector configuration file you can copy into your `connectors` directory to get started:

```yaml
type: connector
driver: snowflake

dsn: "{{ .env.connector.snowflake.dsn }}" 
```

:::tip Using the Add Data Form
You can also use the Add Data form in Rill Developer, which will automatically create the `snowflake.yaml` file and populate the `.env` file with `connector.snowflake.*` parameters based on the parameters or connection string you provide.
:::



Use the following syntax when defining a connection string using a private key:

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_url_safe>
```
See the full documentation to set up [private key authentication](#using-keypair-authentication).

<img src='/img/build/connectors/data-sources/snowflake_conn_strings.png' class='rounded-gif' />
<br />

:::info Finding the Snowflake account identifier

To determine your [Snowflake account identifier](https://docs.snowflake.com/en/user-guide/admin-account-identifier), one easy way is to check your Snowflake account URL. The account identifier to use in your connection string should be everything before `.snowflakecomputing.com`!

:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

## Deploy to Rill Cloud

When deploying your project to Rill Cloud, you must provide Snowflake credentials via the connection string as a source configuration `dsn` field. If these credentials exist in your `.env` file, they'll be pushed with your project automatically.

To manually configure your environment variables, run:
```bash
rill env configure
```

:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::

## Appendix

### Using keypair authentication

You can use keypair authentication for enhanced security when connecting to Snowflake as an alternative to password-based authentication, which Snowflake has deprecated. Per the [Snowflake Go Driver](https://github.com/snowflakedb/gosnowflake) specifications, this requires the following changes to the dsn:
- Remove the password  
- Add `authenticator=SNOWFLAKE_JWT`  
- Add `privateKey=<privateKey_url_safe>` 

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_url_safe>
```

#### Generate a private key

Please refer to the [Snowflake documentation](https://docs.snowflake.com/en/user-guide/key-pair-auth) on how to configure an unencrypted private key to use in Rill.

The Snowflake Go Driver only supports **unencrypted PKCS#8 keys**. Make sure to include the `-nocrypt` flag, as encrypted keys are not supported. You can generate one using: 

```bash
# Generate a 2048-bit unencrypted PKCS#8 private key
openssl genrsa 2048 | openssl pkcs8 -topk8 -inform PEM -out rsa_key.p8 -nocrypt
```

#### Convert the private key to a URL-safe format for the DSN

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