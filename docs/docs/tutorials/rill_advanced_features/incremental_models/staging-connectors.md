---
title: "Staging Connectors"
description:  "Getting started with more advanved topics"
sidebar_label: "Staging Connectors"
sidebar_position: 13
---

There are some connections that are not natively supported such as Snowflake to Clickhouse. In order to successfully ingest data from these types of sources, there are times where a staging table is required. 


:::tip requirements
In order to successfully follow this course, you will need to create an account on Snowflake, AWS, and ClickHouse Cloud. 

Snowflake: We will be using the sample ORDERS dataset in SNOWFLAKE_SAMPLE_DATA, TPCH_SF1.
AWS: In order to use the staging table with S3, you need to have an access key setup with read/write access to S3.
ClickHouse: You will write the output dataset from Snowflake to ClickHouse cloud.
:::


## Getting the Connections ready

Please refer to our docuemntation on how to prepare the [s3](https://docs.rilldata.com/reference/connectors/s3) and [snowflake](https://docs.rilldata.com/reference/connectors/snowflake) connections.

Once these are setup, we can create the staging model file. Let's create one called `staging_to_CH.yaml`


### Creating the YAML components
First, let's define the model. We could add the refresh cron job here but since the data is static in Snowflake, there would be no reason to refresh the data. In the case that it was an updating dataset, you would need to add the incremental and refresh pairs.
```yaml
type: model 

incremental: true
refresh:
  cron: 0 0 * * *
```
Next, we can define the SQL splits based on a time frame. Since that data in the ORDERS dataset in old, we can make the range from some data in the 1990s. Feel free to navigate to your Snowflake console and run some SQL commands to better understand what the data is: 
```sql
select max(o_orderdate) from ORDERS; -- 1998-08-02
select min(o_orderdate) from ORDERS; -- 1992-01-01
```
Next, we use the range of dates created for our splits in our actual SQL query that will read data from Snowflake
```yaml
splits_concurrency: 3

splits:
    connector: duckdb
    sql: SELECT range as day FROM range(TIMESTAMPTZ '1995-01-01', TIMESTAMPTZ '1995-01-31', INTERVAL 1 DAY)

connector: snowflake
sql: SELECT * FROM SNOWFLAKE_SAMPLE_DATA.TPCH_SF1.ORDERS WHERE date_trunc('day', O_ORDERDATE) = '{{ .split.day }}'
```

Since Snowflake cannot write directly to ClickHouse and vice-versa, we use a S3 staging connector that has capabilities to write/read from ClickHouse and Snowflake.
```yaml
stage:
  connector: s3
  path: s3://rilldata-public-s3/temp-data
```
Lastly, we define connector to write the final table to.
```yaml
output:
  connector: clickhouse
  ```

  Your final output should look like:

![img](/img/tutorials/302/staging.png)

:::note
Our team is continuously working to add additional features to staging connectors. If you are looking for a specific combination, please reach out and let us know!
:::



import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />