---
title: SQLite
description: Connect to data in SQLite
sidebar_label: SQLite
sidebar_position: 80
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[SQLite](https://www.sqlite.org/about.html) is a lightweight, self-contained SQL database engine renowned for its reliability, speed, and full-featured, serverless architecture. SQLite is primarily known as an in-process database and is widely used in embedded systems, mobile applications, and various small-to-medium-sized applications due to its simplicity, zero-configuration, and single-file database format. SQLite supports standard SQL syntax and includes features such as transactions and atomic commit and rollback, making it a practical choice for applications requiring a compact, efficient data management system. Rill supports connecting to and reading from a SQLite database as a source through the [DuckDB SQLite extension](https://duckdb.org/docs/extensions/sqlite.html).

## Connect to SQLite


In many cases, since SQLite is used as an in-process database, credentials are not required. Instead, Rill will need to know the path to the SQLite database file so that it can be read accordingly.

```yaml
type: connector
driver: sqlite 

dsn: "file:mydatabase.db" 
```

Alternatively, you can create the connector directly using the [connector YAML reference documentation](/reference/project-files/connectors#sqlite). 


:::tip

If you plan to deploy the project to Rill Cloud, it is recommended that you move the SQLite database file to a `data` folder in your Rill project home directory. You can then use the relative path of the db file in your source definition (e.g., `data/test_sqlite.db`).

:::

## Cloud deployment

Once a project with a SQLite source has been deployed using `rill deploy`, Rill Cloud will need to be able to access and retrieve the underlying database file. In most cases, this means that the corresponding SQLite database file should be included within a directory in your Git repository, which will allow you to specify a relative path in your source definition (from the project root).

:::warning When Using An External SQLite Database

If the SQLite database file is external to your Rill project directory, you will still be able to use the fully qualified path to read this SQLite database _locally_ using Rill Developer. However, when deployed to Rill Cloud, this source will throw an **error**.

:::