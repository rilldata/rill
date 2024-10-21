---
title: "2. Import the Source"
sidebar_label: "2. Import the Source"
sidebar_position: 1
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
---
import LoomVideo from '@site/src/components/LoomVideo'; // Adjust the path as needed

Let's start at the beginning of all data pipelines, **the source**.

### What is a Source?

In Rill, the source is your data. Whether this is from a data warehouse, cloud storage, or a RDBMS, Rill can read and import this data.

:::note Import Data?
By default, the underlying OLAP engine utilized is DuckDB (see <a href='https://docs.rilldata.com/build/olap/' target="_blank"> Connect OLAP engines</a>).
:::

Please see our docs for the 
<a href="https://docs.rilldata.com/build/connect/" target="_blank">supported list</a> of connectors.


### Adding a source is simple! 

Select the `+Add` dropdown and select `Data`, this will open a UI of supported connectors.


<img src = '/img/tutorials/102/Adding-Data.gif' class='rounded-gif' />
<br />

For our tutorial, let's add two GCS storage from our public storage. These are datasets derived from the commit history and modified files of our friends at ClickHouse's GitHub repository. You need to add each one separately. Please refer the GIF above.

```yaml 
gs://rilldata-public/github-analytics/Clickhouse/*/*/modified_files_*.parquet
gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet
```

Once imported, youll see the UI change with a few things..
1. The `source_name`.YAML file created in the file explorer
1. DuckDB database, created in the Connectors explorer
2. Within the DuckDB database, our imported data as a table with a preview.
3. The right panel giving a summary of the data

<img src = '/img/tutorials/102/Add-GCS.gif' class='rounded-gif' />
<br />


Now we're ready to create a `model`.

<details>
  <summary>Don't see what you're looking for?</summary>
  
    We are continually adding new sources and connectors in our releases. For a comprehensive list, you can refer to our <a href=''>connectors page</a>. Please don't hesitate to <a href='https://docs.rilldata.com/contact'>reach out</a> either if there's a connector you'd like us to add!

    If this it your first time, you may need to refresh the browser for DuckDB to appear in the UI.
    
</details>

:::note Too much data?
By default, all environments running locally are considered `dev` environments. This means that you can use environmental variables to filter the input data as Rill Developer is designed for testing purposes. For example, you can filter the repository data on the `author_date` column or simply use `limit ####`.
```
sql: "select * from read_parquet('gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet')
         {{if dev}} where author_date < TIMESTAMPTZ '2015-01-01 00:00:00 Z'  {{end}}"
```

You will need to start rill with `rill start --env dev`.
:::


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />