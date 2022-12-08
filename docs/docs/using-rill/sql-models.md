---
title: Model with SQL
description: Transform your data into One Big Table using DuckDB SQL
sidebar_label: Model with SQL
sidebar_position: 20
---

Data models in Rill are just SQL `SELECT` statements that operate on imported source data. They allow you to join, transform, and clean data to prepare it for use in dashboards.

## SQL transformations with DuckDB

Data transformations in Rill Developer are powered by DuckDB and their dialect of SQL. Under the hood, data models are created as views in DuckDB. Please visit their [documentation](https://duckdb.org/docs/sql/introduction) for insight into how to write your queries.

## "One Big Table"

It is powerful to be able to translate many ad hoc questions into a data framework that can answer a class of questions at scale. For example, high-level insights that are relevant to a company (how much revenue did we make last week?) are more actionable for an employee if it is relevant for their role (how did my campaign increase revenue last week?).

To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable metrics definition to enable ad hoc slice-and-dice discovery through Rill's interactive dashboard.

## Adding a data model

### Using the UI

To add a new data model using the UI, click "+" by Models in the left hand navigation pane. You can now begin typing a SQL `SELECT` query for your model in the code editor – with keystroke-by-keystroke feedback!

When you add a data model using the UI, a code definition will automatically be created as a `.sql` file in the `models` folder in your Rill project.

### Using code

In your Rill project directory, create a `model_name.sql` file in the `models` directory containing a SQL `SELECT` statement. Rill will automatically detect and parse the model next time you run `rill start`.
