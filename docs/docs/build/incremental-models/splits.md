---
title: Splits, a special state
description: C
sidebar_label: Splits
sidebar_position: 05
---

## What are Splits?

In Rill, splits < insert explanation>


## How to define a split


### SQL



### glob
by default its file?
```yaml
splits:
  glob: gs://rendo-test/*/*/*/*/*/*/rilldata-incremental-model.csv
  ```

 glob emits one split per directory
```yaml
glob:
  path: gs://rendo-test/**/*.csv
  partition: directory #hive
```

