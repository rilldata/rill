---
title: Code snippets for Rill modelling on ClickHouse
tags:
- model
- code
- snippets
- clickhouse
docs: https://docs.rilldata.com/build/models
hash: c741f0d4840a0c863fc8c253dc694ef18699b2f97c23d7b6e791d57d6870da8e
---

```yaml
dev:
  partitions:
    connector: s3
    glob: s3://rill-customer-bucket/2025/09/02/00/exchange-7f778d56f5-zplb*.parquet

partitions:
  connector: s3
  glob: s3://rill-customer-bucket/dataset/*/*/02/*/exchange*.parquet
```
