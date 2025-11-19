---
title: Incremental Partitioned Models
description: Create incremental partitioned models
sidebar_label: Incremental + Partitioned Models
sidebar_position: 30
---

Putting the two concepts together, it is possible to create an incremental partitioned model. Doing so will allow you to not only partition the model but also refresh only the partition that you need and incrementally ingest partitions.

:::note Need help?
If you need any assistance with setting up an incremental partitioned model, [reach out](/contact) to us for assistance!
:::

:::tip Looking for an example?

If you're looking for a working example, take a look at [my-rill-tutorial in our examples' repository](https://github.com/rilldata/rill-examples).

:::

As we already know how to set up these separately, let's see what changes in the UI when we enable both on a single model. In the following example, note that both incremental is enabled and partitions are defined by the Google Cloud Storage directory.

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

When this model loads, you will be able to both view the partitions and select a specific partition to refresh via the UI in Rill Developer. Unlike **partitioned-only** models, a new button is added in each of the partitions.

<img src='/img/build/advanced-models/incremental-partitions-developer.png' class='rounded-gif' />
<br />

Likewise, if you refresh using the **CLI**:

```bash
rill project refresh  --model CH_incremental_commits_directory --local --partition ba9f71625de8e042cabf3333576d502c
Refresh initiated. Check the project logs for status updates.
```

## How Incremental Partitioned Models Work

### Initial Ingestion:
When a model is first created, an initial ingestion will occur to bring in all the data, also known as a `Full Refresh`. All refreshes after this will be considered an `incremental refresh`. Note in the below image, the source table writes each section of data to a specific partition as mapped in the YAML file.

<img src='/img/build/advanced-models/initial-ingestion.png' class='rounded-gif' />
<br />

### Additional Partition:
If you add an additional partition to the source table, on the next manual or automatic refresh, Rill will detect the new partition and **only** add the additional partition to the model, as you can see in the diagram, the **blue** additional partition is added in its own partition in the partitioned model. If the other partitions have not been modified, these will not be touched.

<img src='/img/build/advanced-models/additional-partition.png' class='rounded-gif' />
<br />

### Modify Existing Partition:
If you modify any of the already existing partitions, **yellow**, Rill will re-ingest just the modified file during the scheduled refresh by checking the `last_modified_date` parameter.

<img src='/img/build/advanced-models/modified-partition.png' class='rounded-gif' />
<br />