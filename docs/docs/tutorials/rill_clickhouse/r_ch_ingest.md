---
title: "Ingesting Data into ClickHouse"
sidebar_label: "Ingesting Data Directly into ClickHouse"
sidebar_position: 4
hide_table_of_contents: false
tags:
  - OLAP:ClickHouse
---

## Importing your own Data into ClickHouse from ...

Currently, Clickhouse lacks some [direct ingestion](https://clickhouse.com/docs/en/migrations/snowflake) from certain providers. You can navigate to their website for a full list of data sources in which they support [direct ingestion](https://clickhouse.com/docs/en/integrations), via manual import or [ClickPipes](https://clickhouse.com/cloud/clickpipes).

### How does this effect Rill?

When switching from DuckDB, you may have noticed some changes to the capabilities of Rill. By default, we disable modeling when Clickhouse is enabled as the default OLAP engine. However, we can change this behavior by enabling the feature flag `clickhouseModeling`.

```yaml
features:
  clickhouseModeling: true
  ```

Once this is enabled, you'll be able to create model files and add sources in the UI and use these for SQL transformations, as you would with DuckDB. 

:::note
Currently not all the functionality is supported but our team is working on this to add more features! Please reach out to us on our community or via GitHub for any specific missing functionality that you looking for.
:::

### Ingestion directly on ClickHouse or Rill?

Our team created some functionality within Rill for you to be able to import data directly from your warehouses to ClickHouse. 

In the below example, we are importing data from snowflake to ClickHouse using S3 as an intermediate stage.
```yaml
type: model
-- the source of data in Snowflake
connector: snowflake
sql: >
  select * from CUSTOMER limit 1001

-- the staging table in S3
stage:
  connector: s3
  path: s3://rill-developer.rilldata.io/snow

-- the output clickhouse connector
output:
  connector: clickhouse
  materialize: true
```

In order to use this method you will need to set your credentials in .env. If the .env does not already exist (it will be created default if you have created a source), you can create a .env file in the rill directoy by running `touch .env` and this should now be visible in Rill Developer.

```
connector.clickhouse.host="localhost"
connector.clickhouse.port=9000
connector.snowflake.dsn=""
connector.s3.aws_access_key_id=""
connector.s3.aws_secret_access_key=""
```
:::note
If you already set up clickhouse via the .env file, you will just need to add your snowflake and s3 credentials.
:::