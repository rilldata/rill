---
title: Incremental Model YAML
sidebar_label: Incremental Model YAML
sidebar_position: 31
---

## Properties

**`type`** - Refers to the resource type and must be `model` _(required)_.

**`sql`** - Sets the SQL query to extract data from source

**`state`** - explicitly defines a state in which your model to increment, *should be used with **incremental: true***
  - **`sql`** - Sets the SQL query for the state

**`splits`** - a special kind of state, defines the split of the underlying data, can either be `glob` or `sql`. *Cannot be used with **state:***
  - **`glob`** - define the glob pattern of your bucket
    - **`path`** - if you need to define the partition type, use path (keep glob empty)
    - **`partition`** - directory or hive 
  - **`sql`** - Sets the SQL query to define the split

**`splitswatermark`** - ???

**`splitsconcurrency`** - integer, increase concurrency to speed up processing. *Caution: increasing this to a large number will consume large amounts of resources*

**`timeout`** - defines the timeout (in seconds) to read source data, defaults to *3600 seconds, 60 minutes*

**`incremental`** - Refers to the model if it is incremental and must be `true` if using increments _(required)_.

**`refresh`** - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying source data _(optional)_.
  - **`cron`** - a cron schedule expression, which should be encapsulated in single quotes, e.g. `'* * * * *'` _(optional)_
  - **`every`** - a Go duration string, such as `24h` ([docs](https://pkg.go.dev/time#ParseDuration)) _(optional)_

**`inputproperties`** - ???

**`stage`** - defines the staging table. IE: In the direct data ingestion from Snowflake -> S3 -> ClickHouse, you can define `S3` as the staging table as direct read from Snowflake to ClickHouse is not supported.
  - **`connector`** - type of connection, *s3*
  - **`path`** - path for the temporary data *s3://bucket/temp-data*


**`output`** - if final output directory is a different data source, required for staging models
  - **`connector`** - type of connection
  - **`incremental_strategy`** - append or merge, can only be used with SQL outputs

**`materialize`** - materialize view into a table