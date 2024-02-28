---
title: Snowflake 
description: Connect to data in Snowflake
sidebar_label: Snowflake
sidebar_position: 80
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## How to configure credentials in Rill

How you configure access to Snowflake depends on whether you are developing a project locally using `rill start` or are setting up a deployment using `rill deploy`.

### Configure credentials for local development

When developing a project locally, Rill will use the credentials passed via the Snowflake connection string as defined in the `dsn` property in the [source config](../../reference/project-files/sources.md#properties) or via `--var connector.snowflake.dsn=...` while running `rill start` from the CLI. Rill uses the following [syntax](https://pkg.go.dev/github.com/snowflakedb/gosnowflake#hdr-Connection_String) when defining the Snowflake connection string:
```sql
<username>:<password>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>
```

![Retrieving Snowflake connection parameters](/img/deploy/credentials/snowflake_conn_strings.png)

:::tip

To determine your [Snowflake account identifier](https://docs.snowflake.com/en/user-guide/admin-account-identifier), one easy way would be to check your Snowflake account URL and the account identifier to use in your connection string should be everything before `.snowflakecomputing.com`!

:::

### Configure credentials for deployments on Rill Cloud

Similar to the local development workflow, when deploying a project to Rill Cloud, credentials can be passed via Snowflake connection string as a source configuration `dsn` field or by passing / updating the credentials directly used by Rill Cloud by running:
```
rill env configure
```

:::info

Note that you must first `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::
