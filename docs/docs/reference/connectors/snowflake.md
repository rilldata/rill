---
title: Snowflake 
description: Connect to data in Snowflake
sidebar_label: Snowflake
sidebar_position: 11
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

:::info Deprecation of password authentication

Snowflake has issued a [deprecation notice](https://www.snowflake.com/en/blog/blocking-single-factor-password-authentification/) for single-factor password authentication. Rill supports and recommends you use private key authentication to avoid any disruption of your service.

:::

## Overview

[Snowflake](https://docs.snowflake.com/en/user-guide-intro) is a cloud-based data platform designed to facilitate data warehousing, data lakes, data engineering, data science, data application development, and data sharing. It separates compute and storage, enabling users to scale up or down instantly without downtime, providing a cost-effective solution for data management. With its unique architecture and support for multi-cloud environments, including AWS, Azure, and Google Cloud Platform, Snowflake offers seamless data integration, secure data sharing across organizations, and real-time access to data insights, making it a common choice to power many busienss intelligence applications or use cases. Rill supports natively connecting to and reading from Snowflake as a source using the [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake).


<img src = '/img/reference/connectors/snowflake/snowflake.png' class='centered' />
<br />


## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), Rill will use the credentials passed via the Snowflake connection string in one of several ways:
1. As defined in the [source YAML configuration](../../reference/project-files/sources.md#properties) directly via the `dsn` property
2. As defined in the optional _Snowflake Connection String_ field from within the UI source creation workflow (this is equivalent to setting the `dsn` property in the underlying source YAML file)
3. As defined from the CLI when running `rill start --env connector.snowflake.dsn=...`

:::warning Beware of committing credentials to Git

Outside of local development, it is generally not recommended to specify / save the credentials directly in the `dsn` of your source YAML file as this information can potentially be committed to Git!

:::

Rill uses the following syntax when defining a connection string using private key:

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>
```
See the full documentation to setup [private key authentication](#using-keypair-authentication)

<img src = '/img/reference/connectors/snowflake/snowflake_conn_strings.png' class='rounded-gif' />
<br />

:::info Finding the Snowflake account identifier

To determine your [Snowflake account identifier](https://docs.snowflake.com/en/user-guide/admin-account-identifier), one easy way would be to check your Snowflake account URL and the account identifier to use in your connection string should be everything before `.snowflakecomputing.com`!

:::

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](/ingest-sources/credentials/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials that you have set locally for this source.

:::

## Cloud deployment

When deploying a project to Rill Cloud (i.e. `rill deploy`), Rill requires credentials to be passed via the Snowflake connection string as a source configuration `dsn` field or by passing / updating the credentials directly used by Rill Cloud by running:

```
rill env configure
```

:::info

Note that you must first `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

:::tip Did you know?

If you've configured credentials locally already (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/ingest-sources/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::

## Appendix

### Using keypair authentication

Rill supports using keypair authentication for enhanced authentication security to Snowflake as an alternative to basic authentication. Per the [Snowflake Go Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake#hdr-JWT_authentication) specifications, this will imply the following changes to the `dsn` being used (note the `authenticator` and `privateKey` key-value pairs):

:::info
Snowflake currently does not support encrypted keys for their Snowflake driver.
:::

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>
```

:::tip Best Practices

If using keypair authentication, you may want to consider rotating your public key to ensure compliance with security and governance best practices. If rotating your key, you will need to following the described steps below again.

:::

#### Generate a private key

Please refer to the [Snowflake documentation](https://docs.snowflake.com/en/user-guide/key-pair-auth) on how to configure a unencrypted private key to use in Rill.

#### Generate a Base64 URL-safe encoded version of your private key

Following similar steps, we will first need to generate a Base64 URL-safe encoded version of your private key:

```bash

cat rsa_key.p8 | base64 | tr '+/' '-_' | tr -d '\n'

```

:::info Check your OS version

Depending on the OS version, the command to generate a Base64 URL-safe encoded version of your key may slightly differ. Please check your OS reference manual for the correct syntax.

:::

:::tip Check if the encoded output ends with %

Before copying this output (for the next step), please make sure the resulting string does not end with a `%`. To double check, you can try writing the results to a text file and manually checking: `cat rsa_key.p8 | base64 | tr -d '\n' > private_key.txt`.

:::

#### Update Snowflake DSN with encoded private key in Rill

Taking the output of the previous step, you will want to update the DSN accordingly in your source definition:

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>
```

:::note

The Base64 URL-safe encoded private key should be added to your `privateKey` parameter.

:::