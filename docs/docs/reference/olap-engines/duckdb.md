---
title: DuckDB
description: Power Rill dashboards using DuckDB (default)
sidebar_label: DuckDB
sidebar_position: 1
---

## Overview

[DuckDB](https://duckdb.org/why_duckdb.html) is an in-memory, columnar, SQL database designed for analytical (OLAP) workloads, offering high-speed data processing and analysis. Its columnar storage model and vectorized query execution make it highly efficient for OLAP tasks, enabling fast aggregation, filtering, and joins on large datasets. DuckDB's ease of integration with data science tools and its ability to run directly within analytical environments like Python and R, without the need for a separate server, make it an attractive choice for OLAP applications seeking simplicity and performance.

By default, Rill includes DuckDB as an embedded OLAP engine that is used to ingest data from [sources](../connectors/connectors.md) and power your dashboards. Nothing more needs to be done if you wish to power your dashboards on Rill Developer or Rill Cloud. 

:::tip Interested in using DuckDB and another OLAP engine in the same project?

Well now you can! For more details, see our page on [Using Multiple OLAP Engines](multiple-olap.md).

:::