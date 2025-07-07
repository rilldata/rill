---
title: "Let's get back to our project"
description:  Further build on project
sidebar_label: "Source File Changes"
sidebar_position: 00
---

Before we discuss the advanced features, we'll go over how to make changes in Rill Developer and push to Rill Cloud.

In our initial ingestion of the data, we brought in only a month's worth of data to ensure that we do not try to download **all the data** from our source. However, we would want our dashboards in Rill Cloud to display all the data. We can do this by defining the behavior of the source via the YAML. See [our reference documentation](/reference/project-files/sources) for more information.

```
gs://rilldata-public/github-analytics/Clickhouse/2025/03/modified_files_*.parquet
gs://rilldata-public/github-analytics/Clickhouse/2025/03/commits_*.parquet
```


:::tip Local Rill Developer

Rill Developer will always run as `dev` unless explicitly defined with starting Rill via `rill start --environment prod`.
:::

## Walkthough Changes

## PROD vs DEV

### commits.yaml

Your source file should look something like:
```yaml
# Source YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/sources

type: source

connector: "duckdb"
sql: "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/2025/03/commits_*.parquet')"
```

At the current setup, the source will only ingest data from March 2025. Instead, lets change that to our `dev` SQL and create a new line for the full data. This way, when we push the source to Rill Cloud, it won't be just for the month of March but historical, too!


```yaml
dev:
  sql: "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/2025/03/commits_*.parquet')"
  
sql:  "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet')"
```
### modified_files.yaml

Now let's do the same for `modified_files`.


```yaml
# Source YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/sources
  
type: source

connector: "duckdb"
dev:
  sql: "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/2025/03/modified*.parquet')"
  
sql:  "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/*/*/modified*.parquet')"
```


:::tip `{{if dev}} {{end}}`
Similar to separating the SQL file into two separate keys, you can also use `{{if dev}}` to define a rule for the source data.

```yaml
sql: "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet')
         {{if dev}} where author_date > '2025-01-01' {{end}}"```
:::

## Source Refresh

In order to keep the files from being static, you'll need to add a project refresh to the source! 
```yaml
refresh:
  every: 24h
  #cron: "0 8 * * *"
```
Now each day, this source will refresh on its own, and you'll have fresh data. Now let's take a look at the model.




## Let's create a basic partitioned model.
In the previous courses, we used a GCS connection to import ClickHouse's repository commit history. Let's go ahead and assume we are using the same folder structure.

```
#gs://rilldata-public/github-analytics/Clickhouse/YYYY/MM/filename_YYYY_MM.parquet

gs://rilldata-public/github-analytics/Clickhouse/**/commits_.parquet
gs://rilldata-public/github-analytics/Clickhouse/**/modified_files_*.parquet
```
1. Create a YAML file: `partitions-tutorial.yaml`

2. Use `glob:` resolver to load files from GCS
```yaml
type: model

partitions:
  glob:
    connector: gcs
    path: gs://rilldata-public/github-analytics/Clickhouse/2024/*/commits_*.parquet
```
3. Set the SQL statement to user the URI.
```yaml
sql: SELECT * FROM read_parquet('{{ .partition.uri }}')
```

Once you save the file, Rill will start to ingest all the partitions from GCS. This may take a few minutes. You can see the progress of the ingestion from the CLI.

```bash
2024-11-12T13:41:43.355 INFO    Executed model partition        {"model": "partitions_tutorial", "key": "3c4cdfc819f8a64ecaeecbc9ae9702af", "data": {"path":"github-analytics/Clickhouse/2024/01/commits_2024_01.parquet","uri":"gs://rilldata-public/github-analytics/Clickhouse/2024/01/commits_2024_01.parquet"}, "elapsed": "903.89675ms"}
2024-11-12T13:41:44.158 INFO    Executed model partition        {"model": "partitions_tutorial", "key": "ecd933fe9b5089f940e592d500b168a0", "data": {"path":"github-analytics/Clickhouse/2024/02/commits_2024_02.parquet","uri":"gs://rilldata-public/github-analytics/Clickhouse/2024/02/commits_2024_02.parquet"}, "elapsed": "802.034542ms"}
2024-11-12T13:41:44.945 INFO    Executed model partition        {"model": "partitions_tutorial", "key": "0a5023cdd0a340aa95f387bb20c1a942", "data": {"path":"github-analytics/Clickhouse/2024/03/commits_2024_03.parquet","uri":"gs://rilldata-public/github-analytics/Clickhouse/2024/03/commits_2024_03.parquet"}, "elapsed": "786.159292ms"}
```


Once completed you should see the following:

<img src = '/img/tutorials/advanced-models/ch-incremental.png' class='rounded-gif' />
<br />


### Viewing Partition Status in the UI

If you see any errors in the UI regarding your partitions, you may need to check the status by selecting "View partitions"

<img src = '/img/tutorials/advanced-models/partitions-refresh-ui.png' class='rounded-gif' />
<br />

Or, you can check this via the CLI running:
```bash
rill project partitions <model_name> --local

  KEY (50)                           DATA                                                                                                                                                              EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ----------------------------------------------------------------------------------------------------------------------------------------------------------------- ---------------------- --------- ------- 
  9a71c41f9c9b268e7ca3bedfe4c2774b   {"path":"github-analytics/Clickhouse/2014/01/commits_2014_01.parquet","uri":"gs://rilldata-public/github-analytics/Clickhouse/2014/01/commits_2014_01.parquet"}   2024-11-12T20:40:55Z   667ms    
  ...
```

### Refreshing Partitions 

When issues arise in partitions in your model, you will need to fix the underlying issue then refresh this specific partitions in Rill. In the UI, you can select the dropdown `Showing` and select errors.

<img src = '/img/tutorials/advanced-models/errored-partitions.png' class='rounded-gif' />
<br />
Or, if you prefer to refresh in the CLI, you can run the command to refresh all errored partitions.

```bash
rill project refresh --model partitions_tutorial --errored-partitions --project my-rill-tutorial --local
Error: can't refresh partitions on model "partitions_tutorial" because it is not incremental

```

As we reviewed before, this is only possible when the model is an incremental one! Did you catch the issue before running the command? 

## What is Incremental Modeling?
Once partitions are set up, you can use incremental modeling to load only new data when refreshing a dataset. This becomes important when your data is large, and it does not make sense to reload all the data when trying to ingest new data.

### Let's create an Incremental model for our commits and modified files sources.

0. Create a file CH_incremental_commits.yaml and CH_incremental_modified_files.yaml

1. After copying the previous YAML contents, set `incremental` to true 

2. You can manually setup a `partitions_watermark` but since our data is using the `glob` key, it is automatically set to the `updated_on` field. 

3. Let's set up a `refresh` based on `cron` that runs daily at 8AM UTC.
```
refresh:
    cron: "0 8 * * *"
```

Once Rill ingests the data, your UI should look something like this: 

<img src = '/img/tutorials/advanced-models/incremental.png' class='rounded-gif' />
<br />

Your YAML should look like the following:

```yaml
type: model

incremental: true
refresh:
    cron: "0 8 * * *"

partitions:
  glob:
    connector: gcs
    path: gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet #modified_filies_*.parquet

sql: SELECT * FROM read_parquet('{{ .partition.uri }}')
```

You now have a working incremental model that refreshed new data based on the `updated_on` key at 8AM UTC everyday. Along with writing to the default OLAP engine, DuckDB, we have also added some features to use staging tables for connectors that do not have direct read/write capabilities.


