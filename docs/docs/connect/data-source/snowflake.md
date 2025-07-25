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

[Snowflake](https://docs.snowflake.com/en/user-guide-intro) is a cloud-based data platform designed to facilitate data warehousing, data lakes, data engineering, data science, data application development, and data sharing. It separates compute and storage, enabling users to scale up or down instantly without downtime, providing a cost-effective solution for data management. With its unique architecture and support for multi-cloud environments, including AWS, Azure, and Google Cloud Platform, Snowflake offers seamless data integration, secure data sharing across organizations, and real-time access to data insights, making it a common choice to power many business intelligence applications and use cases. Rill supports natively connecting to and reading from Snowflake as a source using the [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake).

<img src='/img/reference/connectors/snowflake/snowflake.png' class='centered' />
<br />

## Local credentials

When using Rill Developer on your local machine (i.e., `rill start`), Rill will use the credentials passed via the Snowflake connection string in one of several ways:
1. As defined in the [Connector YAML configuration](/reference/project-files/connectors#snowflake) directly via the `dsn` property or distinct parameters
2. As defined in the optional _Snowflake Connection String_ field within the UI source creation workflow (this is equivalent to setting the `dsn` property in the underlying source YAML file)

:::warning Beware of committing credentials to Git

Outside of local development, it is generally not recommended to specify or save the credentials directly in the `dsn` of your source YAML file, as this information can potentially be committed to Git!

:::

Rill uses the following syntax when defining a connection string using a private key:

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>
```
See the full documentation to set up [private key authentication](#using-keypair-authentication).

<img src='/img/reference/connectors/snowflake/snowflake_conn_strings.png' class='rounded-gif' />
<br />

:::info Finding the Snowflake account identifier

To determine your [Snowflake account identifier](https://docs.snowflake.com/en/user-guide/admin-account-identifier), one easy way is to check your Snowflake account URL. The account identifier to use in your connection string should be everything before `.snowflakecomputing.com`!

:::

## Cloud deployment

When deploying a project to Rill Cloud (i.e., `rill deploy`), Rill requires credentials to be passed via the Snowflake connection string as a source configuration `dsn` field or by passing/updating the credentials used by Rill Cloud directly by running:

```
rill env configure
```


:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/connect/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::

## Appendix

### Using keypair authentication

Rill supports using keypair authentication for enhanced security when connecting to Snowflake, as an alternative to basic authentication. Per the [Snowflake Go Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake#hdr-JWT_authentication) specifications, this will require the following changes to the `dsn` being used (note the `authenticator` and `privateKey` key-value pairs):

:::info
Snowflake currently does not support encrypted keys for their Snowflake driver.
:::

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>
```

:::tip Best Practices

If using keypair authentication, consider rotating your public key regularly to ensure compliance with security and governance best practices. If you rotate your key, you will need to follow the steps below again.

:::

#### Generate a private key

Please refer to the [Snowflake documentation](https://docs.snowflake.com/en/user-guide/key-pair-auth) on how to configure an unencrypted private key to use in Rill.

#### Generate a Base64 URL-safe encoded version of your private key

Following similar steps, you will first need to generate a Base64 URL-safe encoded version of your private key:

```bash
cat rsa_key.p8 | base64 | tr '+/' '-_' | tr -d '\n'
```

:::info Check your OS version

Depending on your OS version, the command to generate a Base64 URL-safe encoded version of your key may differ slightly. Please check your OS reference manual for the correct syntax.

:::

:::tip Check if the encoded output ends with %

Before copying this output (for the next step), make sure the resulting string does not end with a `%`. To double-check, you can try writing the results to a text file and manually checking: `cat rsa_key.p8 | base64 | tr -d '\n' > private_key.txt`.

:::

#### Update Snowflake DSN with encoded private key in Rill

Take the output of the previous step and update the DSN accordingly in your source definition:

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>
```

:::note

The Base64 URL-safe encoded private key should be added to your `privateKey` parameter.

:::