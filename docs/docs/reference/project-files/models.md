---
title: Model SQL
sidebar_label: Model SQL
sidebar_position: 20
---

Data transformations in Rill Developer are powered by DuckDB and their dialect of SQL. Under the hood, data models are created as views in DuckDB.

Please visit the [DuckDB documentation](https://duckdb.org/docs/sql/introduction) for insight into how to write your models.

In your Rill project directory, create a `<model_name>.sql` file in the `models` directory containing a DuckDB SQL `SELECT` statement. Rill will automatically detect and parse the model next time you run `rill start`.
