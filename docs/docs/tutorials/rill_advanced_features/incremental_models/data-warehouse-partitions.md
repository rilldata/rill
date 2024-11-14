---
title: "Partitions with Data Warehouses"
description:  "Getting Started with Partitions"
sidebar_label: "Data Warehouse: Partitions and Incremental Models"
sidebar_position: 12
---

import ComingSoon from '@site/src/components/ComingSoon';

<ComingSoon />

<div class='contents_to_overlay'>



## Once we have the data ready, we can post this one.


Another advanced concept within Rill is using [Incremental Models](https://docs.rilldata.com/build/advancedmodels/incremental). To understand incremental models, we will also need to discuss [partitions](https://docs.rilldata.com/build/advancedmodels/partitions). 

:::tip requirements
You will need to setup the connection to your data warehouse, depending on the connection please refer to [our documentation](https://docs.rilldata.com/reference/connectors/). 

Your dataset will require a or equivalent to `updated_on` column to use.

:::

## Understanding partitions in Models

Here’s how it works at a high level:

- **Partitions Definition**: Each row from the result set becomes one "partition". The model processes each partition separately.
- **Execution Strategy**:
  - **First Partitions**: Runs without incremental processing.
  - **Subsequent Partitions**: Run incrementally, following the output connector's `incremental_strategy` (either append or merge for SQL connectors).

### Let's create a basic partitions model.


1. Create a YAML file: `SQL-incremental-tutorial.yaml`

2. Use `sql:` resolver to load files from your data warehouse

```yaml
sql: >
  SELECT *
  FROM some_table_in_bq_SF_with_updated_on_column
  {{ if incremental }} WHERE updated_on > CAST(FORMAT_TIMESTAMP('%Y-%m-%d', '{{ .state.max_day }}') AS DATE) {{ end }}
```

Note that ` {{if incremental}}` is needed here as we will use this to increment over your data! As stated in the beginning, you will need an `updated_on` column to calculate the increments. 

### Handling errors in partitions
If you see any errors in the UI regarding partitions, you may need to check the status. You can do this via the CLI running:
```bash
rill project partitions --<model_name> --local
```


### Refreshing Partitions 

Let's say a specific partition in your model had some formatting issues. After fixing the data, you would need to find the key for the partition and run `rill project partitions --<model_name> --local`.  Once found, you can run the following command that will only refresh the specific partition, instead of the whole model.

```bash
rill project refresh --model <model_name> --partition <partition_key>
```

## What is Incremental Modeling?
You can use incremental modeling to load only new data when refreshing a dataset. This becomes important when your data is large and it does not make sense to reload all the data when trying to ingest new data.


3. Add the SQL to calculate the state max day.
```yaml
incremental: true
state:
  sql: SELECT MAX(updated_on) as max_day FROM some_table_in_bq_SF_with_updated_on_column
```

This grabs the MAX value of updated_on from your table.

4. Finally, you will need to define the output and incremental stragety.

```yaml
output:
  connector: duckdb
  incremental_strategy: append
```

Please see below for the full YAML file on incremental modeling from a Data warehouse to DuckDB.
```yaml
materialize: true

connector: bigquery
#connector: snowflake

sql: >
  SELECT *
  FROM some_table_in_bq_SF_with_updated_on_column
  {{ if incremental }} WHERE updated_on > CAST(FORMAT_TIMESTAMP('%Y-%m-%d', '{{ .state.max_day }}') AS DATE) {{ end }}

incremental: true
state:
  sql: SELECT MAX(updated_on) as max_day FROM some_table_in_bq_SF_with_updated_on_column

output:
  connector: duckdb
  incremental_strategy: append
```



You now have a working incremental model that refreshed new data based on the `updated_on` key at 8AM UTC everyday. Along with writing to the default OLAP engine, DuckDB, we have also added some features to use staging tables for connectors that do not have direct read/write capabilities.

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />

</div>