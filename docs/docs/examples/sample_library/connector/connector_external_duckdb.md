---
title: External DuckDB Connector Example YAML
tags:
- connector
- code
- complete_file
- duckdb
docs: https://docs.rilldata.com/connect/data-source/duckdb
hash: 5b329ea483b865c7fb2ca4baca25503d6f6352483c766120dca61473569d560a
---

```yaml
type: connector

driver: duckdb
managed: true

init_sql:
  ATTACH '/path/to/your/duckdb.db' AS external_duckdb;
  INSTALL httpfs;
  LOAD httpfs;
```
