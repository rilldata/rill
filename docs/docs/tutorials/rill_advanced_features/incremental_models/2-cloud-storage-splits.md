---
title: "Splits with Cloud Storage"
description:  "Getting Started with Splits"
sidebar_label: "Cloud Storage: Splits and Incremental Models"
sidebar_position: 12
---

Now that we understand what [Incremental Models](https://docs.rilldata.com/build/advancedmodels/incremental) and [splits](https://docs.rilldata.com/build/advancedmodels/splits) are, let's try to apply them to our project.

Since our ClickHouse data is hosted in GCS/S3, we will be using glob based splits, instead of the example sql select statement.

### Let's create a basic split model.
In the previous courses, we used a GCS connection to import ClickHouse's repository commit history. Let's go ahead and assume we are using the same folder structure.

```
#gs://rilldata-public/github-analytics/Clickhouse/YYYY/MM/filename_YYYY_MM.parquet

gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_.parquet
gs://rilldata-public/github-analytics/Clickhouse/*/*/modified_files_*.parquet
```
1. Create a YAML file: `splits-tutorial.yaml`

2. Use `glob:` resolver to load files from GCS
```yaml
type: model

splits:
  glob:
    connector: gcs
    path: gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet

```
3. Set the SQL statement to user the URI.
```yaml
sql: SELECT * FROM read_parquet('{{ .split.uri }}')
```

Once you save the file, Rill will start to ingest all the splits from GCS. This may take a few minutes. You can see the progress of the ingestion from the CLI.

### Viewing errors in splits

If you see any errors in the UI regarding your splits, you may need to check the status by selecting "view splits"

![img](/img/tutorials/302/splits-refresh-ui.png)


Or, you can check this via the CLI running:
```bash
rill project splits <model_name> --local
```

Once completed you should see the following:

![img](/img/tutorials/302/splits.png)

### Refreshing Splits 

Let's say a specific split in your model had some formatting issues. After fixing the data, you can either select `Refresh Split` in the UI or find the split ID by running `rill project splits --<model_name> --local`.  Once found, you can run the following command that will only refresh the specific split, instead of the whole model.

```bash
rill project refresh --model <model_name> --split <split_key>
```


## What is Incremental Modeling?
Once splits are set up, you can use incremental modeling to load only new data when refreshing a dataset. This becomes important when your data is large and it does not make sense to reload all the data when trying to ingest new data.

### Let's create an Incremental model for our commits and modified files sources.

0. Create a file CH_incremental_commits.yaml and CH_incremental_modified_files.yal

1. After copying the previous YAML contents, set `incremental` to true (For modified_files, make sure you change the file name!)

2. You can manually setup a `splits_watermark` but since our data is using the `glob` key, it is automatically set to the `updated_on` field. 

3. Let's set up a `refresh` based on `cron` that runs daily at 8AM UTC.
```
refresh:
    cron: "0 8 * * *"
```

Once Rill ingests the data, your UI should look something like this: 
![img](/img/tutorials/302/incremental.png)


Your YAML should look like the following:

```yaml
type: model

incremental: true
refresh:
    cron: "0 8 * * *"

splits:
  glob:
    connector: gcs
    path: gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet #modified_files_*.parquet

sql: SELECT * FROM read_parquet('{{ .split.uri }}')
```

You now have a working incremental model that refreshed new data based on the `updated_on` key at 8AM UTC everyday. Along with writing to the default OLAP engine, DuckDB, we have also added some features to use staging tables for connectors that do not have direct read/write capabilities.


Once this is created


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />