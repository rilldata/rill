---
title: "Ingesting Data into ClickHouse"
sidebar_label: "Ingesting Data Directly into ClickHouse"
sidebar_position: 50
hide_table_of_contents: false
tags:
  - OLAP:ClickHouse
  - Tutorial
---

## Importing your own Data into ClickHouse from ...

Currently, ClickHouse lacks some [direct ingestion](https://clickhouse.com/docs/en/migrations/snowflake) from certain providers. You can navigate to their website for a full list of data sources in which they support [direct ingestion](https://clickhouse.com/docs/en/integrations), via manual import or [ClickPipes](https://clickhouse.com/cloud/clickpipes).

### How does this affect Rill?

When switching from DuckDB, you may have noticed some changes to the capabilities of Rill. By default, we disable modeling when ClickHouse is enabled as the default OLAP engine. However, we can change this behavior by enabling the feature flag `clickhouseModeling`.

```yaml
features:
  clickhouseModeling: true
  ```

Once this is enabled, you'll be able to create model files and add sources in the UI and use these for SQL transformations, as you would with DuckDB. 

:::note
Currently not all the functionality is supported but our team is working on this to add more features! Please reach out to us on our community or via GitHub for any specific missing functionality that you looking for.
:::

## Ingestion directly from Snowflake to ClickHouse

In the below example, we are importing data from Snowflake to ClickHouse using S3 as an intermediate stage.
```yaml
type: model
materialize: true 

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

In order to use this method you will need to set your credentials in .env. If the .env does not already exist (it will be created by default if you have created a source), you can create a .env file in the rill directory by running `touch .env` and this should now be visible in Rill Developer.

```
connector.clickhouse.host="localhost"
connector.clickhouse.port=9000
connector.snowflake.dsn=""
connector.s3.aws_access_key_id=""
connector.s3.aws_secret_access_key=""
```
:::note
If you already set up ClickHouse via the .env file, you will just need to add your snowflake and s3 credentials.
:::


## Ingestion directly from BigQuery to ClickHouse

In the below example, we are importing data from Big Query directly to ClickHouse using GCS as an intermediate stage.


```yaml
type: model
materialize: true 

connector: bigquery
sql: |
    SELECT
      *
    FROM `<project_id>.<dataset_name>.<table>`

project_id: "<project_id>"

stage:
  connector: gcs
  path: 'gs://rill-bq-ch/temp/'


output:
  connector: clickhouse
```

You'll need to ensure that your provided `google_application_credentials` have all the required permissions on both [BigQuery](https://cloud.google.com/bigquery/docs/access-control) and [GCS](https://cloud.google.com/storage/docs/access-control/iam-roles). Ensure that your .env has the following:

```
connector.clickhouse.host="localhost"
connector.clickhouse.port=9000
google_application_credentials=""
```