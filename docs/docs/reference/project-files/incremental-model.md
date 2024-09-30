---
title: Incremental Model YAML
sidebar_label: Incremental Model YAML
sidebar_position: 31
---

## Properties

**`type`** - Refers to the resource type and must be `model` _(required)_.

**`sql`** - Sets the SQL query to extract data from source

**`state`** - 

**`splits`** - Defines the split of the underlying data, can either be `glob` or `sql`. 
  - **`glob`** - 
  - **`sql`** - 


**`incremental`** - Refers to the model if it is incremental and must be `true` if using increments _(required)_.

**`refresh`** - Specifies the refresh schedule that Rill should follow to re-ingest and update the underlying source data _(optional)_.
  - **`cron`** - a cron schedule expression, which should be encapsulated in single quotes, e.g. `'* * * * *'` _(optional)_
  - **`every`** - a Go duration string, such as `24h` ([docs](https://pkg.go.dev/time#ParseDuration)) _(optional)_

