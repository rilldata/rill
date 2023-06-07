---
title: Glob Patterns
description: Support for importing data sources using glob patterns
sidebar_label: Glob patterns
sidebar_position: 10
---

Rill supports ingesting data from a group of files using glob patterns in the URI of source files. For example, the following URI will ingest all parquet files in the `my-bucket` bucket that were created in January 2023:

`
gs://my-bucket/v=1/y=2023/m=1/*.parquet
`

By default, there are some limits on the amount of data that can be ingested using glob patterns. These limits can be configured in the `source.yaml` file. The following are the default limits:
- **Total bytes downloaded**: 10GB
- **Total file matches**: 1000
- **Total files listed**: 1 million

These limits can be increased or decreased as needed. For example, to increase the limit on the total bytes downloaded to 100GB, you would add the following line to the `source.yaml` file:

`
glob.max_total_size: 1073741824000
`

## Extract policies

Rill also supports extracting a subset of data matching a glob pattern. There are two types of extract policies:
1. **File based limits** : These limits restrict the number of files that are ingested.
2. **Row based limits** : These limits restrict the amount of data that is ingested from each file.

These policies can be applied together or individually.
  - Here are the possible combination and their semantics.
    - If both `rows` and `files` are specified, each file matching the `files` clause will be extracted according to the `rows` clause.
    - If only `rows` is specified, no limit on the number of files is applied. For example, getting a 1 GB `head` extract will download as many files as necessary.
    - If only `files` is specified, each file will be fully ingested.

Each policy can be configured by specifying two parameters:
1. **Size** : The number of files/size of data to be fetched.
2. **Strategy** : The strategy to fetch data. Currently, only **Head** (first n up to size) or **Tail** (last n up to size) is supported.

For example, you could extract the first 100MB of data from the first 10 files matching a glob pattern by using the following extract policy in the source.yaml file:

```
extract:
  files:
    strategy: head
    size: 10
  rows:
    strategy: head
    size: 100MB
```


### Performance considerations of extract policies

1. It is important to note that the system may fetch more data than what is specified in order to implement row based limits. 
    - For parquet files the data is typically fetched in row groups, and the entire row group is fetched. 
    - For zipped files, the files are fully downloaded.
2. Implementing a `tail`-based file strategy on a glob pattern matching a large number of files can take an unusually long time since the files are listed first, and then the tail is computed in-memory.

## Schema relaxation

Rill ingests data matching a glob pattern in batches and tries to automatically adjust the schema of ingested data if files matching the glob pattern have different schemas. 
This involves the following considerations:
1. Adding a new column if the previously ingested data has fewer columns than the new files.
2. Automatically converting the datatype of a column to match the datatype in all files.
3. Adding null values for columns that were previously present but are not present in the new data.

These features are enabled by default and can be disabled by setting `ingest.allow_field_addition`(for adding new columns) and `ingest.allow_field_relaxation`(for changing datatype and adding nulls for missing columns in new data) to **false** in source.yaml.

Example source.yaml:
```
type: "gcs"
uri: "gs://my-bucket/v=1/y=2023/m=*/d=0[1-7]/H=01/*.parquet" 

# add new column but don't change datatype and ingest null values for missing columns in subsequent data.
ingest.allow_field_addition: true
ingest.allow_field_relaxation: false

```