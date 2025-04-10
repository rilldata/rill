---
title: "Let's get back to our project"
description:  Further build on project
sidebar_label: "DEV and PROD Sources"
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

## Modifying the YAML.

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

## Adding automatic source refresh
In order to keep the files from being static, you'll need to add a project refresh to the source! 
```yaml
refresh:
  every: 24h
  #cron: "0 8 * * *"
```
Now each day, this source will refresh on its own, and you'll have fresh data. Now let's take a look at the model.
