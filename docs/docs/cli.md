---
sidebar_position: 10
title: CLI
---

# quick start new project
You can create and augment projects in Rill Developer using the CLI.  Every project starts by initializing the experience. Once initialized, you can ingest data into the project and start the UI.

```
rill init
rill import-source /path/to/data_1.parquet
rill start

```

# the Rill CLI
CLI comands help us initialize projects and programatically augment projects. If you would like to see information on all the available CLI commands, you can use the ```--help``` option.  There are additional details on each command below.

```
rill --help
```

## initialize your project
Initialize your project using the ```init``` command.  

```
rill init
```

### project references
Rill works best if you have `cd`ed into the project directory, since it assumes that you are in a project directory already. But you can also specify a new project folder by including the --project option.

```
rill init --project /path/to/a/new/project
rill import-source /path/to/data_1.parquet --project /path/to/a/new/project
rill start --project /path/to/a/new/project
```

### start your project
Start the User Interface to interact with your imported sources and revisit projects you have created.

```
rill start
```
  
The Rill Developer UI will be available at http://localhost:8080.

## import your data
Import datasets of interest into Rill Developer's [duckDB](https://duckdb.org/docs/sql/introduction) database to make them available. We currently support .parquet, .csv, and .tsv data ingestion.

```
rill import-source /path/to/data_1.parquet
rill import-source /path/to/data_2.csv
rill import-source /path/to/data_3.tsv
```

### source names
By default the source name will be a sanitized version of the dataset file name. You can specify a name using the --name option.
  
```
rill import-source /path/to/data_1.parquet --name my_source
```

### dropping a source
If you have added a source to Rill Developer that you want to drop, you can do so using the --drop-source option.

```
rill drop-source my_source
```

### csv delimiter
If you have a dataset that is delimited by a character other than a comma or tab, you can use the --delimiter option. DuckDB can also attempt to automatically detect the delimiter, so it is not strictly necessary.

```
rill import-source /path/to/data_4.txt --delimiter "|"
```

## existing duckDB databases

### connecting
You can connect to an existing duckdb database by passing ```--db``` with path to the db file.

Any updates made directly to the tables in the database will reflect in Rill Developer.  Similarly, any changes made by Rill Developer will modify the database.

Make sure to have only one connection open to the database, otherwise there will be some unexpected issues.
```
rill init --db /path/to/duckdb/file
```

### copying
You can also copy over the database so that there are no conflicts and overrides to the source. Pass ```--copy``` along with ```--db``` to achieve this.

```
rill init --db /path/to/duckdb/file --copy
```
