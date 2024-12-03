---
title: "Increment Model based on a state from Data Warehouses"
description:  "Getting Started with Partitions"
sidebar_label: "Data Warehouse: Increment Models"
sidebar_position: 12
---

Another advanced concept within Rill is using [Incremental Models](/build/incremental-models/#what-is-an-incremental-model) with a state defined. 

:::tip requirements
You will need to setup the connection to your data warehouse, depending on the connection please refer to [our documentation](https://docs.rilldata.com/reference/connectors/). 

In this example we use a DATE column as our defining state but depending on your data, you can use any defining column.

:::

## Understanding States in Models

Here’s how it works at a high level:

- **State Definition**: Based on a SQL query, defines a key that allows you to increment your model by
- **Execution Strategy**:
  - **Full Refresh**: Runs without incremental processing.
  - **Incremental Refresh**: Run incrementally based on the state defined, following the output connector's `incremental_strategy` (either append or merge for SQL connectors).

### Let's create a basic partitions model.

:::note Example
In this example, we are using a sample dataset that exists in Big Query: rilldata.ssb_100.date.
In this case our table is not getting updated, so instead we'll modify the SQL to show you how incremental works.
:::


1. Create a YAML file: `SQL_incremental_tutorial.yaml`

2. Use the following contents to create your own model.
```yaml
type: model
materialize: true

connector: "bigquery" #or "snowflake"

incremental: true
state:
  sql: SELECT MAX(DATE) as max_date FROM SQL_incremental_tutorial #should be the name of the current model

sql: |
  SELECT *,
         PARSE_DATE('%Y%m%d', CAST(D_DATEKEY AS STRING)) AS DATE
  FROM rilldata.ssb_100.date
  {{if incremental}} # when incremental refreshing this part of the SQL is used.
    WHERE PARSE_DATE('%Y%m%d', CAST(D_DATEKEY AS STRING)) = '{{.state.max_date}}' #normally would want to set this to where DATE > '{{.state.max_date}}' to only append new rows.
  {{else}} 
    LIMIT 10 #restricts the full refresh to only 10 rows, so when we run incremental, its easy to tell the difference. 
  {{end}}

output:
  connector: duckdb
  incremental_strategy: append #merge, requires unique_key
  #unique_key: [column_name] #if strategy is merge
```

3. In the UI, try refreshing both incrementally and fully to see the difference in the model that loads. 
- when selecting a full refresh, only 10 rows should be returned. 
- when selecting incremental refresh, it will **append** values to the inital 10 values in the full refresh. 

![img](/img/tutorials/302/data-warehouse-refresh.png)

:::note Partition vs. State
Unlike partitions, states do not paritition the dataset per refresh so you will not be able to via the UI or CLI, see if there is a specific partition that errored and manually refresh this. In the cases of data disrecpancies in a state incremented model, please run a full refresh. 
:::

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />
