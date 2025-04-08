---
title: "2. Import the Source"
sidebar_label: "2. Import the Source"
sidebar_position: 1
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
---

Let's start at the beginning of all data pipelines, **the source**.

### What is a Source?

In Rill, the source is your data. Whether you need to connect to the data warehouse via SQL or provide a folder in your cloud storage, Rill can read this data. Depending on the source type, you will need to either explicitly provide the credentials (Snowflake, Athena, etc.) or Rill can dynamically retrieve them via the CLI (AWS, GCP, Azure). Either way, the credentials will [be stored in a `.env` file](/build/credentials#deploying-to-rill-cloud), that is pushed with your Rill project.



:::tip Where is the data being imported?
By default, the underlying OLAP engine utilized is DuckDB (see <a href='https://docs.rilldata.com/build/olap/' target="_blank"> Connect OLAP engines</a>). 
Please see our docs for the 
<a href="https://docs.rilldata.com/build/connect/" target="_blank">supported list</a> of connectors.


We support various difference OLAP engines, if you have a different OLAP engine that you're interested in using, please let us know! 
:::


### Add a Data Source

Select the `+Add` dropdown and select `Data`, this will open a UI of supported connectors.


<img src = '/img/tutorials/102/Adding-Data.gif' class='rounded-gif' />
<br />

For our tutorial, let's add two GCS storage from our public storage. In Rill, each dataset is added separately and a single object. Once imported into Rill, you can then transform the data via model, but we'll get into that on the next page. Depending on the source type, you will need to provide a SQL statement, bucket location, URL, etc. While some of our sources derive the credentials via the CLI (S3, GCS, Azure), others you will need to manually input. As these datasets are public you will be able to ingest automatically.


```yaml 
gs://rilldata-public/github-analytics/Clickhouse/2025/03/modified_files_*.parquet
gs://rilldata-public/github-analytics/Clickhouse/2025/03/commits_*.parquet
```
:::note Data
These are datasets derived from the commit history and modified files of our friends at ClickHouse's GitHub repository. In our example, we'll ingest a single month of data however, Rill supports glob patterns in which you could modify the URL to `gs://rilldata-public/github-analytics/Clickhouse/**/modified_files_*.parquet` which would ingest all years and months of data. 
:::

Once imported, youll see the UI change with a few things.
1. The `source_name`.YAML file created in the file explorer.
2. DuckDB database, created in the Connectors explorer.
3. Within the DuckDB database, under main, the source table with a preview when selected.
4. The right panel giving a summary of the data source and the column values.

<img src = '/img/tutorials/102/Add-GCS.gif' class='rounded-gif' />
<br />


Now we're ready to create a `model`.

<details>
  <summary>Don't see what you're looking for?</summary>
  
    We are continually adding new sources and connectors in our releases. For a comprehensive list, you can refer to our <a href=''>connectors page</a>. Please don't hesitate to <a href='https://docs.rilldata.com/contact'>reach out</a> either if there's a connector you'd like us to add!

    If this it your first time, you may need to refresh the browser for DuckDB to appear in the UI.
    
</details>

:::tip Too much data?
By default, all environments running locally are considered `dev` environments. This means that you can use environmental variables to filter the input data as Rill Developer is designed for testing purposes. For example, you can filter the repository data on the `author_date` column or simply use `limit ####`.
```
sql: "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet')
         {{if dev}} where author_date > '2025-01-01' {{end}}"
```
:::
