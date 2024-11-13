---
title: Advanced Model YAML
sidebar_label: Advanced Model YAML
sidebar_position: 35
hide_table_of_contents: true
---

In some cases, advanced models will be required when implementing advanced features such as incremental models. 

## Properties

**`type`** - refers to the resource type and must be 'model'

**`refresh`** - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying source data _(optional)_.
  - **`cron`** - a cron schedule expression, which should be encapsulated in single quotes, e.g. `'* * * * *'` _(optional)_
  - **`every`** - a Go duration string, such as `24h` ([docs](https://pkg.go.dev/time#ParseDuration)) _(optional)_

**`timeout`**
 — The maximum time to wait for model ingestion _(optional)_.

**`incremental`** - set to `true` or `false` whether incremental modeling is required _(optional)_

**`state`** - refers to the explicitly defined state of your model, cannot be used _(optional)_ 
  - **`sql/glob`** - refers to the location of the data depending if the data is cloud storage or a data warehouse.

**`partitions`** - refers to the special state that is defined by the a predefined partition. In the case of partitions, your data needs to already be in a supported format. 
  - **`sql/glob`** - refers to the location of the data depending if the data is cloud storage or a data warehouse.
  - **`path`** - 
  - **`partition`** - 

**`partitions_watermark`** - 

**`partitions_concurrency`** - 

**`stage`** - in the case of staging models, where an input source does not support direct write to the output and a staging table is required _(optional)_. 
  - **`connector`** - refers to the connector type for the staging table
  -**`path`** - path of the temporary staging table

**`output`** - in the case of staging models, where the output needs to be defined where the staging table will write the temporary data _(optional)_. 
  - **`connector`** - refers to the connector type for the staging table

**`materialize`** - 