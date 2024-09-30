---
title: Create Incremental Models
description: C
sidebar_label: Create Incremental Models
sidebar_position: 00
---

In Rill, < insert description why incremental modeling is important.>

## What is an Incremental Model

Unlike [regular models](../models/models.md) that are created via SQL select queries, incremental models are define in a YAML file and are used when:
-
- ...
-

Whether your data exists in cloud storage or in a data warehouse, Rill will be able to increment and ingest depending on the settings you define.

:::note
Incremental Modeling in Rill is an ongoing development, while we do have support for the following, please reach out to us if you have any specific requirements.

Cloud Storage:
S3, GCS

Data Warehouse
BigQuery, Snowflake

Is this right? 
:::


## Creating an Incremental Model

 In order to enable incremental model, you will need to set the following: `incremental: true`.


```yaml
type: model

sql: SELECT now() AS inserted_on
incremental: true
```


### Types of Incremental Models
There are two main types of incremental models. 

1. Incremental Model with explicit `state:` defined
2. Incremental Model with Splits
    2a. SQL splits
    2b. glob splits

### Incremental Models with State defined

If your data is not [split](splits.md), you will need to define the incremental model with a predefined `state`.

```yaml
type: model
incremental: true

sql: SELECT {{ if incremental }} {{ .state.max_val }} + 1 {{ else }} 0 {{ end}} AS val, now() AS inserted_on
state:
  sql: SELECT MAX(val) as max_val FROM incremental_state
```

Once state is defined in an incremental model, its value can be used as a variable in your SQL statement. In the above example, the state is defined as the max value of `val` column from the table `incremental_state`. 

If the model is already an incremental model, the SQL is using the max value and adding 1 and saving as val. If it's the first run, IE, not incremental, it will set the val as 0.
