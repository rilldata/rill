---
title: Create Incremental Models
description: C
sidebar_label: Create Incremental Models
sidebar_position: 00
---

In Rill, < insert descriptio why incremental modeling is important.>

## What is an Incremental Model

Unlike [regular models} (link to model page) that are created via SQL select queries, incremental models are define in a YAML file and are used when 
- the ingestion data is large,
- ...




## Creating an Incremental Model

 In order to enable incremental model, you will need to set the following: `incremental: true`.


```yaml
type: model

sql: SELECT now() AS inserted_on
incremental: true
```


### Types of Incremental Models

1. Incremental Model with explicit State
2. Incremental Model with Splits
    2a. SQL splits
    2b. glob splits

