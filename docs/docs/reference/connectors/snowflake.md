---
title: Snowflake 
description: Connect to data in Snowflake
sidebar_label: Snowflake
sidebar_position: 11
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Snowflake](https://docs.snowflake.com/en/user-guide-intro) is a cloud-based data platform designed to facilitate data warehousing, data lakes, data engineering, data science, data application development, and data sharing. It separates compute and storage, enabling users to scale up or down instantly without downtime, providing a cost-effective solution for data management. With its unique architecture and support for multi-cloud environments, including AWS, Azure, and Google Cloud Platform, Snowflake offers seamless data integration, secure data sharing across organizations, and real-time access to data insights, making it a common choice to power many busienss intelligence applications or use cases. Rill supports natively connecting to and reading from Snowflake as a source using the [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake).

![Connecting to Snowflake](/img/reference/connectors/snowflake/snowflake.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), Rill will use the credentials passed via the Snowflake connection string in one of several ways:
1. As defined in the [source YAML configuration](../../reference/project-files/sources.md#properties) directly via the `dsn` property
2. As defined in the optional _Snowflake Connection String_ field from within the UI source creation workflow (this is equivalent to setting the `dsn` property in the underlying source YAML file)
3. As defined from the CLI when running `rill start --var connector.snowflake.dsn=...`

:::warning Beware of committing credentials to Git

Outside of local development, it is generally not recommended to specify / save the credentials directly in the `dsn` of your source YAML file as this information can potentially be committed to Git!

:::

Rill uses the following [syntax](https://pkg.go.dev/github.com/snowflakedb/gosnowflake#hdr-Connection_String) when defining the Snowflake connection string:

```sql
<username>:<password>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>
```

![Retrieving Snowflake connection parameters](/img/reference/connectors/snowflake/snowflake_conn_strings.png)

:::info Finding the Snowflake account identifier

To determine your [Snowflake account identifier](https://docs.snowflake.com/en/user-guide/admin-account-identifier), one easy way would be to check your Snowflake account URL and the account identifier to use in your connection string should be everything before `.snowflakecomputing.com`!

:::

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](/build/credentials/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials that you have set locally for this source.

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

If you've configured credentials locally already (in your `<RILL_HOME>/.env` file), you can use `rill env push` to [push these credentials](/build/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::

## Appendix

### Using keypair authentication

Rill supports using keypair authentication for enhanced authentication security to Snowflake as an alternative to basic authentication. Per the [Snowflake Go Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake#hdr-JWT_authentication) specifications, this will imply the following changes to the `dsn` being used (note the `authenticator` and `privateKey` key-value pairs):

```sql
<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>&authenticator=SNOWFLAKE_JWT&privateKey=<privateKey_base64_url_encoded>
```

:::tip Best Practices

If using keypair authentication, you may want to consider rotating your public key to ensure compliance with security and governance best practices. If rotating your key, you will need to following the described steps below again.

:::

#### Generate a private key

You will first want to generate a 2048-bit PKCS#8 encoded RSA private key:

```bash

openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -pkeyopt rsa_keygen_pubexp:65537 | openssl pkcs8 -topk8 -outform der -nocrypt > rsa_key.p8

```

:::info

This will create a private key called `rsa_key.p8` in your current working directory.

:::

#### Generate a public key

You will next want to extract a 2048-bit PKI encoded RSA public key from the private key:

```bash

openssl pkey -pubout -inform der -outform der -in rsa_key.p8 -out rsa_key.pub

```

:::info

This will create a public key called `rsa_key.pub` in your current working directory.

:::


#### Store the private and public keys securely

It is recommended to copy the public and private key files to a local directory for storage (and record the path to these files). Please note that the private key is stored using the PKCS#8 (Public Key Crypotgraphy Standards) format. 

:::note Securing your keys

To ensure best practices, please make sure to secure these key files when they is not being used and to protect these files from unauthorized access by using the appropriate file permission mechanisms provided by your operating system.

:::

#### Generate a Base64 URL-safe encoded version of your public key

Before assigning your public key to your user in Snowflake, we will need to generate a Base64 URL-safe encoded version of the public key using the following command:

```bash

cat rsa_key.pub | base64 | tr -d '\n'

```

:::info Check your OS version

Depending on the OS version, the command to generate a Base64 URL-safe encoded version of your key may slightly differ. Please check your OS reference manual for the correct syntax.

:::

:::tip Check if the encoded output ends with %

Before copying this output (for the next step), please make sure the resulting string does not end with a `%`. To double check, you can try writing the results to a text file and manually checking: `cat rsa_key.pub | base64 | tr -d '\n' > public_key.txt`.

:::

#### Assign the encoded public key to a Snowflake user

Taking the output from the previous step, you can follow the steps described in Snowflake's documentation [**here**](https://docs.snowflake.com/user-guide/key-pair-auth#assign-the-public-key-to-a-snowflake-user) to assign the public key to an appropriate Snowflake user. 

#### Verify the user's public key fingerprint

Follow Snowflake's documentation [**here**](https://docs.snowflake.com/user-guide/key-pair-auth#verify-the-user-s-public-key-fingerprint) to verify the public key fingerprint and ensure the public key has been configured properly for the user.

:::info

You can use the same commands provided by Snowflake with minimal changes. You will only need to update the name of the user and public key (if different from `rsa_key.pub`)

:::

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