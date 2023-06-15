---
title: Glob Patterns
description: Support for importing data sources using glob patterns
sidebar_label: Glob patterns
sidebar_position: 10
---

Rill supports ingesting data from a group of files using glob patterns in the URI of source files. This allows you to specify a pattern that matches multiple files, making it easier to ingest data from a group of related files. Let's explore how glob patterns work in Rill.

## Ingesting Data with Glob Patterns

To ingest data using glob patterns, you include the pattern in the URI of the source files. Here's an example URI that will ingest all Parquet files in the `my-bucket` bucket that were created in January 2023:
`
gs://my-bucket/v=1/y=2023/m=1/*.parquet
`

By default, Rill applies certain limits when using glob patterns to ingest data. The default limits are as follows:
- **Total size of all matching files**: 10GB
- **Total file matches**: 1000
- **Total files listed**: 1 million

These limits can be configured in the `source.yaml` file. To modify the default limits, you can update the source.yaml file with following fields:
 - `glob.max_total_size` : The maximum total size (in bytes) of all objects. 
 - `glob.max_objects_matched`: The total file matches allowed.
 - `glob.max_objects_listed`: The total files listed to match against the glob pattern. 

For example, to increase the limit on the total bytes downloaded to 100GB, you would add the following line to the `source.yaml` file:
`
glob.max_total_size: 1073741824000
`

## Extract policies

Rill also supports extracting a subset of data matching a glob pattern. There are two types of extract policies:
1. **File based limits** : These limits restrict the number of files that are ingested.
2. **Row based limits** : These limits restrict the amount of data that is ingested from each file.

You can apply these policies individually or in combination to control the extraction process.
  - Here are the possible combination and their semantics.
    - If both `rows` and `files` are specified, each file matching the `files` clause will be extracted according to the `rows` clause.
    - If only `rows` is specified, no limit on the number of files is applied. For example, getting a 1 GB `head` extract will download as many files as necessary.
    - If only `files` is specified, each file upto limit will be fully ingested.

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

When implementing extract policies, there are some performance considerations to keep in mind:
1. It is important to note that the system may fetch more data than what is specified in order to implement row based limits. 
    - For parquet files the data is typically fetched in row groups, and the entire row group is fetched. 
    - For zipped files, the files are fully downloaded.
2. Implementing a `tail`-based file strategy on a glob pattern matching a large number of files can take an unusually long time since the files are listed first and then the tail is computed in-memory.

## Schema relaxation

When ingesting data that matches a glob pattern, Rill processes the data in batches and attempts to automatically adjust the schema if the files have different schemas. This allows for flexibility when dealing with files that might have variations in their structure. Here are the considerations involved in schema relaxation:

1. **Adding new columns**: If the previously ingested data has fewer columns than the new files, Rill adds the new columns to the schema.
2. **Automatic datatype conversion**: Rill automatically converts the datatype of a column to match the datatype in all files. This ensures consistency across the ingested data.
3. **Handling missing columns**: If columns were present in the previously ingested data but are not present in the new data, Rill adds null values for those columns.

By default, these schema relaxation features are enabled in Rill. However, if you prefer to disable them, you can set the following properties to false in the source.yaml file:

 - **ingest.allow_field_addition**: Disables the addition of new columns when ingesting data.
 - **ingest.allow_field_relaxation**: Disables the automatic datatype conversion and addition of null values for missing columns.

Here's an example **source.yaml** configuration:
```
type: "gcs"
uri: "gs://my-bucket/v=1/y=2023/m=*/d=0[1-7]/H=01/*.parquet" 

ingest.allow_field_addition: true
ingest.allow_field_relaxation: false

```

That covers the usage of glob patterns in Rill, including how to ingest data, apply extract policies, and handle schema relaxation. Enjoy working with your data using Rill's powerful features!