---
title: "Ingesting Data into MotherDuck"
sidebar_label: "Ingesting Data Directly into MotherDuck"
sidebar_position: 40
hide_table_of_contents: false
tags:
  - OLAP:MotherDuck
  - Tutorial
---

## Importing your own Data into MotherDuck 



Since both Rill and MotherDuck are based on DuckDB, the read/write capabilities work without having to enable any feature flags.

Simple use the Rill UI to connect to a data source and define the output to MotherDuck.

<img src = '/img/build/connect/sources.png' class='rounded-gif' />
<br />


## Ingestion directly from GCS to MotherDuck


```yaml
# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/advanced-models

type: model
materialize: true

connector: motherduck
sql: |
  select * from read_parquet('gs://rilldata-public/auction_data.parquet')

output:
  connector: motherduck
```

## Ingestion directly from BigQuery to MotherDuck

In the below example, we are importing data from Big Query directly to MotherDuck using GCS as an intermediate stage.


```yaml
type: model
materialize: true 

connector: bigquery
sql: |
    SELECT
      *
    FROM `<project_id>.<dataset_name>.<table>`

project_id: "<project_id>"

output:
  connector: motherduck
```
