---
title: Glob Patterns
description: Support for importing data sources using glob patterns
sidebar_label: Glob Patterns
sidebar_position: 30
---

Rill supports ingesting data from a group of files using glob patterns in the URI of source files. This allows you to specify a pattern that matches multiple files, making it easier to ingest data from a group of related files. Let's explore how glob patterns work in Rill.

## Ingesting Data with Glob Patterns

To ingest data using glob patterns, you include the pattern in the URI of the source files. Here's an example URI that will ingest all Parquet files in the `my-bucket` bucket that were created in January 2023:
`
gs://my-bucket/y=2023/m=01/*.parquet
`

By default, Rill can apply certain limits when using glob patterns to ingest data. The default limits are as follows:
- **Total size of all matching files**: 100 GB
- **Total file matches**: unlimited
- **Total files listed**: unlimited

These limits can be configured in the `.yaml` file for the source. To modify the default limits, you can update the `.yaml` file with following fields:
- `glob.max_total_size`: The maximum total size (in bytes) of all objects. 
- `glob.max_objects_matched`: The total file matches allowed.
- `glob.max_objects_listed`: The total files listed to match against the glob pattern. 

For example, to set the limit on the total bytes downloaded to 1 GB, you would add the following line to the `source.yaml` file:
```yaml
type: "duckdb"
sql: "select * from table"

glob.max_total_size: 1073741824
glob.max_objects_matched: 100000000
```

## Extract policies

Rill also supports extracting a subset of data matching a glob pattern. There are two types of extract policies:
1. **Row based limits**: These limits restrict the amount of data that is ingested from each file.
2. **File based limits**: These limits restrict the number of files that are ingested.


You can apply these policies individually or in combination to control the extraction process.
  - Here are the possible combination and their semantics.
    - If only `rows` is specified, no limit on the number of files is applied. For example, getting a 1 GB `head` extract will download as many files as necessary.
    - If only `files` is specified, each file up to limit will be fully ingested.
    - If both `rows` and `files` are specified, each file matching the `files` clause will be extracted according to the `rows` clause.

Each policy can be configured by specifying two parameters:
1. **`size`**: The size of data in bytes (for rows) or number of files (for files) to fetch..
2. **`strategy`**: The strategy to fetch data. Currently, only `head` (first n up to size) or `tail` (last n up to size) is supported.

### Example 1: Extract first 100 MB
To extract first 100 MB data from a source use the following extract policy in the .yaml file for source:
```yaml
extract:
  rows:
    strategy: head
    size: 100MB
```

### Example 2: Extract the first 100 MB from each of the first 10 files
To extract the first 100 MB of data from each of the first 10 files from a source use the following extract policy in the .yaml file for source:
```yaml
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
2. **Automatic datatype conversion**: Rill automatically converts the datatype of a column to match the datatype in all files. This ensures consistency across the ingested data. The datatype is only relaxed to wider types like `INTEGER` to `BIGINT` or `DOUBLE` to `VARCHAR` and would never be a breaking change like `VARCHAR` to `INTEGER` or `DOUBLE` to `INTEGER`.
3. **Handling missing columns**: If columns were present in the previously ingested data but are not present in the new data, Rill adds null values for those columns.

Schema relaxation is enabled by default. Set `ingest.allow_schema_relaxation: false` in your source's YAML file to disable it.

For example:
```yaml
type: source
connector: "gcs"
uri: "gs://my-bucket/y=2023/m=*/d=0[1-7]/H=01/*.parquet" 

ingest.allow_schema_relaxation: false
```