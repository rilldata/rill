---
title: CLI Documentation
description: You can create and augment projects in Rill Developer using the CLI.
---

# The Rill CLI

:::tip

## Quick start a new project
You can create and augment projects in Rill Developer using the CLI. Every project starts by initializing the experience. Once initialized, you can ingest data into the project and start the application.

```
rill init
rill import-source /path/to/data_1.parquet
rill start
```

or try our example:
```
rill init-example
```
<!-- (Please note that the command `rill init-example` is temporarily unavailable on Windows.) -->

:::

## Help Menu
CLI comands help us initialize and augment projects. If you would like to see information on all the available CLI commands, you can use the ```--help``` option.  There are additional details on each command below.

```
rill --help
```

## Initialize your project
Initialize your project using the ```init``` command.  

```
rill init
```

## Project references
You can specify a project folder outside of the current folder by including the `--project` option.

```
rill init --project /path/to/a/new/project
rill import-source /path/to/data_1.parquet --project /path/to/a/new/project
rill start --project /path/to/a/new/project
```

## Start your project
Start the application to interact with your imported sources and revisit projects you have created.

```
rill start
```
  
The Rill Developer application will be available at [http://localhost:8080](http://localhost:8080).

## Import your data
Import datasets of interest into Rill Developer's [duckDB](https://duckdb.org/docs/sql/introduction) database to make them available. We currently support .parquet, .csv, and .tsv data ingestion.

```
rill import-source /path/to/data_1.parquet
rill import-source /path/to/data_2.csv
rill import-source /path/to/data_3.tsv
```

### Source names
By default the source name will be a sanitized version of the dataset file name. You can specify a name using the `name` command.
  
```
rill import-source /path/to/data_1.parquet --name my_source
```

### Source overwrite
By default source name conflicts will prompt a warning message asking if you want to overwrite the existing source data. You can force Rill Developer to overwrite any existing sources without this warning by using the `force` command.
  
```
rill import-source /path/to/data_1.parquet --name my_source
```

### File Delimiters
If you have a dataset that is delimited by a character other than a comma or tab, you can use the `--delimiter` option. DuckDB can also attempt to automatically detect the delimiter, so it is not strictly necessary.

```
rill import-source /path/to/data_4.txt --delimiter "|"
```

## Dropping a source
If you have added a source to Rill Developer that you want to drop, you can do so using the `drop-source` command.

```
rill drop-source my_source
```
---
## Existing duckDB databases

### Connecting
You can connect to an existing duckdb database by passing the `--db` option with a path to the db file.

Any updates made directly to the sources in the database will be reflected in Rill Developer.  Similarly, any changes made by Rill Developer will modify the database.

Make sure to have only one connection open to the database, otherwise there will be some unexpected issues.
```
rill init --db /path/to/duckdb/database.db
```

### Copying
You can also copy over the database so that there are no conflicts and overrides that are propigated to the source by passing the `--db` option with `--copy` to achieve this.

```
rill init --db /path/to/duckdb/database.db --copy
```