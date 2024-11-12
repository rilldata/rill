---
title: Partitions, a special state
description: C
sidebar_label: Partition your Model
sidebar_position: 05
---

## What are Partitions?

In Rill, partitions are a special type of state in which you can explicitly partition the model into parts. Depending on if your data is in cloud storage or a data warehouse, you can use the `glob` or `sql` parameters. 

You can manage partitions via the CLI using the `rill project partitions` command.
```bash
rill project partitions 
List partitions for a model

Usage:
  rill project partitions [<project>] <model> [flags]

Flags:
      --project string      Project Name
      --path string         Project directory (default ".")
      --model string        Model Name
      --pending             Only fetch pending partitions
      --errored             Only fetch errored partitions
      --local               Target locally running Rill
      --page-size uint32    Number of partitions to return 
```


## How to Define a Partition in a Model
Under the `partitions:` parameter, you will define the pattern in which your data is stored.

### SQL
When defining your SQL, it is important to understand the data that you are querying and creating a split that makes sense. For example, possibly selecting a distinct customer_name per partition, or possibly partition the SQL by a chronological partition, such as month.

```yaml
partitions:
  sql: SELECT range AS num FROM range(0,10) #num is the split variable and can be referenced as {{partition.num}}
  #sql: SELECT DISTINCT customer_name as cust_name from table #results in {{partition.cust_name}}
  ```


### glob

When defining the glob pattern, you will need to consider whether you'd partition the data by folder or file.
In the first example, we are paritioning by each file with the suffix data.csv.
```yaml
partitions:
  glob: gs://rendo-test/**/*data.csv
  ```

If you'd prefer to partition it by folder your can add the partition parameter and define it as `directory`.
```yaml
glob:
  path: gs://rendo-test/**/*data.csv
  partition: directory #hive
```


## Viewing Partitions in Rill Developer


Once `partitions:` is defined in your model, a new button will appear in the right hand panel, `View Partitions`. When selecting this, a new UI will appear with all of your partitions and more information on each. Note that these can be sorted on all, pending, and errors.

![img](/img/tutorials/302/partitions-refresh-ui.png)


# Incremental and Partitions 