---
title: Snowflake 
description: Connect to data in Snowflake
sidebar_label: Snowflake
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

Snowflake is a cloud-based data platform designed to facilitate data warehousing, data lakes, data engineering, data science, data application development, and data sharing. It separates compute and storage, enabling users to scale up or down instantly without downtime, providing a cost-effective solution for data management. With its unique architecture and support for multi-cloud environments, including AWS, Azure, and Google Cloud Platform, Snowflake offers seamless data integration, secure data sharing across organizations, and real-time access to data insights, making it a common choice to power many busienss intelligence applications or use cases. Rill supports natively connecting to and reading from Snowflake as a source using the [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake).

![Connecting to Snowflake](/img/reference/connectors/snowflake/snowflake.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), Rill will use the credentials passed via the Snowflake connection string in one of several ways:
1. As defined in the [source YAML configuration](../../reference/project-files/sources.md#properties) directly via the `dsn property`
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

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](../../build/credentials/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials that you have set locally for this source.

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

If you've configured credentials locally already (in your `<RILL_HOME>/.home` file), you can use `rill env push` to [push these credentials](../../build/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::
