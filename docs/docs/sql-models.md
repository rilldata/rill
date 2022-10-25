---
title: Model SQL transformations
description: Data is modeled into One Big Table using DuckDB.
---

## SQL transformations with DuckDB
Data transformations in Rill Developer are powered by DuckDB and their dialect of SQL. Please visit their [documentation](https://duckdb.org/docs/sql/introduction) for insight into how to write your queries.

## One Big Table
It is powerful to be able to translate many ad hoc questions into a data framework that can answer a class of questions at scale. For example, high-level insights that are relevant to a company (how much revenue did we make last week?) are more actionable for an employee if it is relevant for their role (how did my campaign increase revenue last week?).

To experience the full potential of Rill, model your data sources into "One Big Table" – a granular resource that contains as much information as possible and can be rolled up in a meaningful way. This flexible OBT can be combined with a generalizable metrics definition to enable ad hoc slice-and-dice discovery through Rill's interactive dashboard.