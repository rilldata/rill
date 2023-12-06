---
title: Snowflake 
description: Connect to data in Snowflake
sidebar_label: Snowflake
sidebar_position: 60
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## How to configure credentials in Rill

How you configure access to Snowflake depends on whether you are developing a project locally using `rill start` or are setting up a deployment using `rill deploy`.

### Configure credentials for local development

When developing a project locally, Rill uses the credentials passed via a source config `dsn` (Snowflake connection string) field or via `--env connector.snowflake.dsn=...` while running `rill start`. 
Rill uses the following [format](https://pkg.go.dev/github.com/snowflakedb/gosnowflake#hdr-Connection_String) of Snowflake connection string:
```
my_user_name:my_password@ac123456/my_database/my_schema?warehouse=my_warehouse&role=my_user_role
```

### Configure credentials for deployments on Rill Cloud

Similarly to the local development, when deploying a project to Rill Cloud, credentials might be passed via Snowflake connection string as a source config `dsn` field.

Alternatively, you can pass/update the credentials used by Rill Cloud by running:
```
rill env configure
```
Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.
