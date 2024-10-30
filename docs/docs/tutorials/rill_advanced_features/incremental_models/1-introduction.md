---
title: "Splits and Increments"
description:  "Whats the differences"
sidebar_label: "Splits and Incremental Models"
sidebar_position: 1
---

In order to help with data ingestion into Rill, we will introduce the concepts of [splits](https://docs.rilldata.com/build/advancedmodels/splits) and [incremental models](https://docs.rilldata.com/build/advancedmodels/incremental) Before diving into our ClickHouse project, let's understand what each of these are used for and do.

## Incremental Model

An incremental model is defined using the following key pair.

```yaml
incremental: true
```
Once this is enabled, this allows Rill to configure the model YAML as an incrementing model. 
In some following examples, we will use both a time based incremental and glob based increments. 

## Splits

Splits in models are enabled by defining the split parameter as seen below:

```yaml
splits:
    sql/glob: some splits definition
```

Depending on your data, this can be defined as a `SQL:` statement or a `glob:` pattern. Once configured, Rill will try to split your existing data into smaller subcategories which allows you to refresh specific sections only instead of reingesting the whole dataset. (only when incremental is enabled)

By running the following command, you can see all the available splits and run a refresh command on a specific key or keys.
```bash
rill project split <model_name>
```

```bash
rill project refresh --model <model_name> --split <split_key> 
```


Let's look at a few simple examples before diving into our ClickHouse project.

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />