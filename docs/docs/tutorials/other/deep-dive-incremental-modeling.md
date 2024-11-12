---
title: "Deep Dive into Incremental Modeling"
sidebar_label: "Understanding Incremental Modeling"
sidebar_position: 10
hide_table_of_contents: false
---

In this article, we will, along with some code, go through a detailed step by step example of how to use incremental modeling.

:::note Requirements
In order to follow all the contets this guide, you will need access to a cloud based provider (s3, gcs) and datawarehouse (bq, snowflake) as well as be comfortable with running some basic python code (this will be provided).
:::

## Getting Started

Before starting the guide, it is a good idea to review the documentation on [incremental modeling](  and [partitions](. I will assume that you've already have gone through the trial and have some basic understanding of the concepts. 

Ensure that you have python installed on your system so that we can use some basic scripts to write files / run SQL queries to update the data manually.


## Cloud storage Incremental Modeling

In order to understand and control the new data ingestion, we will be using a python code that manually writes into GCS / S3 based on the following format `/YYYY/MM/DD/HH/MM/SS/rilldata-incremental-model.csv`. 

A few things to note:
1. On first run, the program will try to read your credentials. If it does not exist, it will request for you locate the JSON.
2. The code will request a number of rows to write to your bucket.
3. The number of rows will increase the column, `number` so that we can see the increasing values. 
4. Please modify line 142 and add you bucket name: `    bucket_name = ''  # replace with your bucket name`

Let's try to run the code and see what the first run looks like.

```bash
python3 deepdive.py

Enter the path to your GCS credentials JSON file: /Users/royendo/Downloads/rilldata.json
Max 'number' value found across all files: 0
Enter the number of new rows to add: 10
File uploaded to GCS at path: 2024/09/30/17/38/44/rilldata-incremental-model.csv
Successfully updated rilldata-incremental-model.csv with 10 new rows.
```

Now, lets open Rill and create an Incremental Model.

### Creating the Incremental Model.

```yaml
type: model
incremental: true

partitions:
  glob: glob: gs://rendo-test/**/rilldata-incremental-model.csv

sql: SELECT * FROM read_csv('{{ .split.uri }}', auto_detect=true, ignore_errors=1, header=true)
```

Let's take a minute to check out the split that was just created when reading in the initial data.
```bash
rill project partitions  deepdive --local 
  KEY                                DATA                                                                                                                                       EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ------------------------------------------------------------------------------------------------------------------------------------------ ---------------------- --------- ------- 
  591d2ae89c74e7c11516097fbf3c1723   {"path":"2024/09/30/17/38/44/rilldata-incremental-model.csv","uri":"gs://rendo-test/2024/09/30/17/38/44/rilldata-incremental-model.csv"}   2024-09-30T23:49:39Z   319ms     
  ```

Run the code again to insert 5 more rows. You won't see the new data yet populated in Rill even when refreshing the page. You'll have to run a refresh of the model. Let's do that.

### Refreshing the Incremental Model

```bash
rill project refresh --model deepdive --local                 
Refresh initiated. Check the project logs for status updates.
```

```bash
rill project partitions  deepdive --local                  
  KEY                                DATA                                                                                                                                       EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ------------------------------------------------------------------------------------------------------------------------------------------ ---------------------- --------- ------- 
  591d2ae89c74e7c11516097fbf3c1723   {"path":"2024/09/30/17/38/44/rilldata-incremental-model.csv","uri":"gs://rendo-test/2024/09/30/17/38/44/rilldata-incremental-model.csv"}   2024-09-30T23:49:39Z   319ms            
  442b7a42276b914d64eace9aab917a34   {"path":"2024/09/30/17/50/04/rilldata-incremental-model.csv","uri":"gs://rendo-test/2024/09/30/17/50/04/rilldata-incremental-model.csv"}   2024-09-30T23:53:05Z   165ms     
  ```

What about if we create several files and refresh them all together?

```bash
 rill project partitions  deepdive --local        
  KEY (5)                            DATA                                                                                                                                       EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ------------------------------------------------------------------------------------------------------------------------------------------ ---------------------- --------- ------- 
  591d2ae89c74e7c11516097fbf3c1723   {"path":"2024/09/30/17/38/44/rilldata-incremental-model.csv","uri":"gs://rendo-test/2024/09/30/17/38/44/rilldata-incremental-model.csv"}   2024-09-30T23:49:39Z   319ms            
  442b7a42276b914d64eace9aab917a34   {"path":"2024/09/30/17/50/04/rilldata-incremental-model.csv","uri":"gs://rendo-test/2024/09/30/17/50/04/rilldata-incremental-model.csv"}   2024-09-30T23:53:05Z   165ms            
  1fcda4bee740beb41f2237aa7711de95   {"path":"2024/09/30/17/54/02/rilldata-incremental-model.csv","uri":"gs://rendo-test/2024/09/30/17/54/02/rilldata-incremental-model.csv"}   2024-09-30T23:55:38Z   175ms            
  2596a58289ffa507ff17af7459d5be16   {"path":"2024/09/30/17/55/32/rilldata-incremental-model.csv","uri":"gs://rendo-test/2024/09/30/17/55/32/rilldata-incremental-model.csv"}   2024-09-30T23:55:38Z   169ms            
  2d33015a8e0b091ce44ef4c0961eda91   {"path":"2024/09/30/17/55/35/rilldata-incremental-model.csv","uri":"gs://rendo-test/2024/09/30/17/55/35/rilldata-incremental-model.csv"}   2024-09-30T23:55:39Z   160ms  
  ```

  Note that while they are in different folder paths in GCS, these are all executed on, or refreshed at the same time stamp, as you would expect. Rill is reading for the new files that did not exist before **or** files that have changed since last ingestion and is incrementally refreshing these. 

### Automatic Refreshes
After pushing to Rill Cloud, you will want to have an automatic refresh of the model, so you aren't running a refresh command manually as we just did. In this case, you can setup a refresh like below.

```yaml
refresh:
  cron: 0 0 * * *
  ```


### Per file or per directory?

Depending on your use case and how your data is set up, there are a few things to consider for how you set up your file hierarchy and incremental modeling. 

The current setup, creates a split per file. This doesn't have any adverse affects as we only have 1 file in each directory. But what if you had mulitple files in a directory that when modified, you want to ensure that **all** the files in the directory are refreshed and synced?

  ```yaml
glob:
  path: glob: gs://rendo-test/**/*.csv
  partition: directory
  ```




## Data Warehouse Incremental Modeling

