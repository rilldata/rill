---
title: Splits, a special state
description: C
sidebar_label: Split your Model
sidebar_position: 05
---

## What are Splits?

In Rill, splits are a special type of state in which you can explicitly split the model into parts. Depending on if your data is in cloud storage or a data warehouse, you can use the `glob` or `sql` parameters. 


## How to Define a Split
Under the `split:` parameter, you will define the pattern in which your data is stored.

### SQL
When defining your SQL, it is important to understand the data that you are querying and creating a split that makes sense. For example, possibly selecting a distinct customer_name per split, or possibly split the SQL by a chronological split, such as month.

```yaml
splits:
  sql: SELECT range AS num FROM range(0,10) #num is the split variable and can be referenced as {{split.num}}
  #sql: SELECT DISTINCT customer_name as cust_name from table #results in {{split.cust_name}}
  
  ```


### glob

When defining the glob pattern, you will need to consider whether you'd partition the data by folder or file.
In the first example, we are paritioning by each file with the suffix data.csv.
```yaml
splits:
  glob: gs://rendo-test/**/*data.csv
  ```

If you'd prefer to partition it by folder your can add the partition parameter and define it as `directory`.
```yaml
glob:
  path: gs://rendo-test/**/*data.csv
  partition: directory #hive
```


## Viewing Splits in Rill Developer


Once `splits:` is defined in your model, a new button will appear in the right hand panel, `View splits`.

![splits-ui](/img/build/incremental-models/splits-ui.png)

When selecting this, a new UI will appear with all of your splits and more information on each. Note that these can be sorted on all, pending, and errors.

![splits-ui](/img/build/incremental-models/splits-overview-ui.png)



### Refreshing Split Models

For split models that are not incremented, you only have the option to refresh the full data. 



### Refreshing Incremental Split models
If both `incremental` and `splits` are enabled, you have the ability to refresh a split individually.

![refresh-split](/img/build/incremental-models/splits-refresh-ui.png)