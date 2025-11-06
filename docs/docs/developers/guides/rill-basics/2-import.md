---
title: "2. Import the Source"
sidebar_label: "2. Import the Source"
sidebar_position: 1
hide_table_of_contents: false

tags:
  - Tutorial
  - OLAP:DuckDB
  - Rill Developer
  - Getting Started
---

Let's start at the beginning of all data pipelines: **the source**. 

### What is a Source?

In Rill, a source model represents your raw data. See our list of [connectors](/build/connectors) or select Add -> Add Data from Rill Developer to see the supported sources.

<img src = '/img/tutorials/rill-basics/add_data.png' class='rounded-gif' style={{maxWidth: '400px', height: 'auto'}} />
<br />

Depending on the source type, you will either need to explicitly provide credentials (for Snowflake, Athena, etc.), or Rill can dynamically retrieve them via the CLI (for AWS, GCP, Azure). In either case, the credentials need to be [stored in a `.env` file](/build/connectors/credentials) in order to be pushed to your deployed project. You might need to run `rill env configure` after deploying if your credentials are not pushed properly.

:::tip Where is the data being imported?
By default, Rill uses DuckDB as the underlying OLAP engine (see <a href='https://docs.rilldata.com/build/olap/' target="_blank">Connect OLAP engines</a>).
Please see our docs for the <a href="https://docs.rilldata.com/build/connectors/source/" target="_blank">supported list</a> of connectors.

We support various OLAP engines. If you have a different OLAP engine that you're interested in using, please let us know! Looking for a ClickHouse tutorial? Click [here!](/guides/rill-clickhouse/r_ch_launch)
:::

### Add a Data Source

Select the `+Add` dropdown and select `Data`. This will open a UI showing supported connectors.

<img src = '/img/tutorials/rill-basics/Adding-Data.gif' class='rounded-gif' />
<br />

For our tutorial, let's add two GCS storage sources from our public storage. 

In Rill, each dataset is added separately as a single source model. Once imported into Rill, you can then transform the data via SQL modeling (we'll cover that on the next page). Follow the steps in the UI and use the following URIs below.

```yaml 
gs://rilldata-public/github-analytics/Clickhouse/2025/03/modified_files_*.parquet
gs://rilldata-public/github-analytics/Clickhouse/2025/03/commits_*.parquet
```


:::tip Data
These are datasets derived from the commit history and modified files of our friends at ClickHouse's GitHub repository. In our example, we'll ingest a single month of data. However, Rill supports glob patterns, so you could modify the URL to `gs://rilldata-public/github-analytics/Clickhouse/**/modified_files_*.parquet` to ingest all years and months of data. However, that's a lot of data to ingest! 
:::

Once imported, you'll see the UI change with several things:
1. A `source_name.yaml` file created in the file explorer.
2. A DuckDB database created in the Connectors explorer.
3. Within the DuckDB database, under main, the source table with a preview when selected.
4. The right panel showing a summary of the data source and column values.

<img src = '/img/tutorials/rill-basics/Add-GCS.gif' class='rounded-gif' />
<br />

Now we're ready to create a `model`.

<details>
  <summary>Don't see what you're looking for?</summary>
  
    We are continually adding new sources and connectors in our releases. For a comprehensive list, you can refer to our <a href='https://docs.rilldata.com/build/connectors/source/'>connectors page</a>. Please don't hesitate to <a href='https://docs.rilldata.com/contact'>reach out</a> if there's a connector you'd like us to add!

    If this is your first time, you may need to refresh the browser for DuckDB to appear in the UI.
    
</details>

:::tip Too much data?
By default, all environments running locally are considered `dev` environments. This means that you can use environment variables to filter the input data, as Rill Developer is designed for testing purposes. For example, you can filter the repository data on the `author_date` column or simply use `limit ####`.
```