---
title: SQLite
description: Connect to data in SQLite
sidebar_label: SQLite
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[SQLite](https://www.sqlite.org/about.html) is a lightweight, self-contained SQL database engine renowned for its reliability, speed, and full-featured, serverless architecture. SQLite is primarily known as an in-process database and widely used in embedded systems, mobile applications, and various small-to-medium sized applications due to its simplicity, zero-configuration, and single-file database format. SQLite supports standard SQL syntax and includes features such as transactions and atomic commit and rollback, making it a practical choice for applications requiring a compact, efficient data management system. Rill supports connecting and reading from a SQLite database as a source through the [DuckDB SQLite extension](https://duckdb.org/docs/extensions/sqlite.html).

![Connecting to SQLite](/img/reference/connectors/sqlite/sqlite.png)

## Local credentials

In many cases, since SQLite is used as an in-process database, credentials are not required. Instead, Rill will need to know the path to the SQLite database file so that it can be read in accordingly. If creating a new SQLite source from the UI, you should pass in the appropriate path to the database file under **DB** and the name of the table under **Table**:

![SQLite UI creation](/img/reference/connectors/sqlite/sqlite_example.png)

On the other hand, if you are creating the source YAML file directly, the definition should look something like:

```yaml

type: "duckdb"
sql: "SELECT * FROM sqlite_scan('<path_to_sqlite_db>', '<table_name>');"

```

:::tip

If you plan to deploy the project to Rill Cloud, it is recommended that you move the SQLite database file to a `data` folder in your Rill project home directory. You can then use the relative path of the db file in your source definition (e.g. `data/test_sqlite.db`).

:::


## Cloud deployment

Once a project with a SQLite source has been deployed using `rill deploy`, Rill Cloud will need to be able to have access to and retrieve the underlying database file. In most cases, this means that the corresponding SQLite database file should be included within a directory in your Git repository, which will allow you to specify a relative path in your source definition (from the project root).

:::warning When Using An External SQLite Database

If the SQLite database file is external to your Rill project directory, you will still be able to use the fully qualified path to read this SQLite database _locally_ using Rill Developer. However, when deployed to Rill Cloud, this source will throw an **error**.

:::