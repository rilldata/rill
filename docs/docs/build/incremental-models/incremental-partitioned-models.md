---
title: Incremental Partitioned Models
description: C
sidebar_label: Incremental Partitioned Models
sidebar_position: 05
---


Putting the two concepts together, it is possible to create a incremental partitioned model. Doing so will allow you to not only partition the model but refresh only the partition that you need and incrementally ingest partitions 

:::tip

:::


As we already know how to set up these separately, let's see what changes in the UI when we enable both on a single model. In the following example, note that both incremental is enabled and partitions are defined by the google cloud storage directory.

```yaml
type: model

incremental: true
refresh:
    cron: "0 8 * * *"

partitions:
  glob:
    path: gs://rilldata-public/github-analytics/Clickhouse/2024/*/*
    partition: directory
  
sql: |
  SELECT * 
     FROM read_parquet('{{ .partition.uri }}/commits_*.parquet') 
    WHERE '{{ .partition.uri }}' IS NOT NULL
```

### Refreshing Partitions in Incremental Models

When this model loads, you will be able to both view the partitions and select a specific partition to refresh via the UI in Rill Developer. Unlike **only partitioned** models, a new button is added in each of the partitons. 

![img](/img/build/advanced-models/incremental-partitioned-model.png)

Likewise, if you refresh using the **CLI**.

```bash
rill project refresh  --model CH_incremental_commits_directory --local --partition ba9f71625de8e042cabf3333576d502c
Refresh initiated. Check the project logs for status updates.
```





## How Incremental Partitioned Models Work

### Initial Ingestion:
When a model is first created, an intial ingestion will occur to bring in all of the data. This is also what occurs when you run a `Full Refresh`.

<div style={{ textAlign: "center" }}>
<img src="/img/build/advanced-models/initial-ingestion.png" width="600" />
</div>

### Additional Partition:
If you add an additional partition to the source table, on the next refresh, Rill will detect the new partition and **only** add the additional partition to the model, as you can see in the diagram. If the other partitions have not been modified, these will not be touched. 
<div style={{ textAlign: "center" }}>

<img src="/img/build/advanced-models/addition-partition.png" width="600" />
</div>

### Modify Existing Partition:
If you modify any of the already existing partitions, Rill will reingest just the modified file during the scheduled refresh by checking the `last_modified_date` parameter.
<div style={{ textAlign: "center" }}>

<img src="/img/build/advanced-models/modified-partition.png" width="600" />
</div>

