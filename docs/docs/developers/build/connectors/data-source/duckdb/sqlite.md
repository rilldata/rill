---
title: SQLite
description: Connect to data in SQLite
sidebar_label: SQLite
sidebar_position: 160
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[SQLite](https://www.sqlite.org/about.html) is a lightweight, self-contained SQL database engine renowned for its reliability, speed, and full-featured, serverless architecture. SQLite is primarily known as an in-process database and is widely used in embedded systems, mobile applications, and various small-to-medium-sized applications due to its simplicity, zero-configuration, and single-file database format. SQLite supports standard SQL syntax and includes features such as transactions and atomic commit and rollback, making it a practical choice for applications requiring a compact, efficient data management system. Rill supports connecting to and reading from a SQLite database as a source through the [DuckDB SQLite extension](https://duckdb.org/docs/extensions/sqlite.html).

## Connect to SQLite

SQLite databases are read through DuckDB's [SQLite extension](https://duckdb.org/docs/extensions/sqlite.html) using the `sqlite_scan()` function. No separate connector is needed.

Create a model file (e.g., `models/my_sqlite_data.yaml`):

```yaml
type: model
connector: duckdb
materialize: true

sql: |
  SELECT *
  FROM sqlite_scan('data/mydatabase.db', 'my_table')
```

:::tip

If you plan to deploy the project to Rill Cloud, place the SQLite database file in a `data` folder in your Rill project directory and use the relative path (e.g., `data/mydatabase.db`).

:::

## Deploy to Rill Cloud

Once a project with a SQLite source has been deployed using `rill deploy`, Rill Cloud will need to be able to access and retrieve the underlying database file. In most cases, this means that the corresponding SQLite database file should be included within a directory in your Git repository, which will allow you to specify a relative path in your source definition (from the project root).

:::warning When Using An External SQLite Database

If the SQLite database file is external to your Rill project directory, you will still be able to use the fully qualified path to read this SQLite database _locally_ using Rill Developer. However, when deployed to Rill Cloud, this source will throw an **error**.

:::