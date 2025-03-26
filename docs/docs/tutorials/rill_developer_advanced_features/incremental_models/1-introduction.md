---
title: "Partitions and Increments"
description:  "Whats the differences"
sidebar_label: "Partitions and Incremental Models"
sidebar_position: 1
---


In order to help with data ingestion into Rill, we will introduce the concepts of [partitions](https://docs.rilldata.com/build/incremental-models/#what-are-partitions) and [incremental models](https://docs.rilldata.com/build/incremental-models/#what-is-an-incremental-model) Before diving into our ClickHouse project, let's understand what each of these are used for and do.



:::tip Review the Reference! 
While we will go over the main points to get started, there are more customizations possiblities so we recommened to review the [reference guide](https://docs.rilldata.com/reference/project-files/advanced-models) and docs along with following the tutorial.

:::
## [Incremental Model](https://docs.rilldata.com/build/incremental-models/#what-is-an-incremental-model)

An incremental model is defined using the following key pair.

```yaml
incremental: true
```

Once this is enabled, Rill will configure the model YAML as an incrementing model. 
In some of the examples, we will use both a time based incremental and glob based increments. 

## [Partitioned Model](https://docs.rilldata.com/build/incremental-models/#what-are-partitions)


Partitions in models are enabled by defining the partition parameter as seen below:

```yaml
partitions:
    sql/glob: some partition definition
```

Depending on your data, this can be defined as a `SQL:` statement or a `glob:` pattern. Once configured, Rill will try to partition your existing data into smaller subcategories which allows you to refresh specific partitions only instead of reingesting the whole dataset. (only when incremental is enabled)

By running the following command, you can see all the available partitions.
```bash
rill project partitions <model_name>
```

Let's look at a few simple examples before diving into our ClickHouse project.

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />