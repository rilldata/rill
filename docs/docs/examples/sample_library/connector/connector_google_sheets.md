---
title: GSheets Connector using DuckDB Example YAML
tags:
- connector
- code
- complete_file
- gcp
- gsheets
docs: https://docs.rilldata.com/connect/data-source/googlesheets
hash: 3863c4645eb6f8f15895c939edbdb3101537cf654d577361c8c7d5215fd1dfa6
---

```yaml
type: model

connector: "duckdb"

sql: "select * from read_csv_auto('https://docs.google.com/spreadsheets/d/SPREADSHEET_ID/export?format=csv&gid=SHEET_ID', normalize_names=True)"
```
