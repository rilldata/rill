---
title: Metrics Cache Example YAML
tags:
- metrics
- code
- snippets
- clickhouse
docs: https://docs.rilldata.com/build/metrics-view/measures/windows
hash: 4d5e2718d3667f1bca96bfef0c3e9eb5731c13f12e38a7e341205e90bd32a691
---

```yaml
cache:
  enabled: true
  key_sql: "SELECT max(process_time) FROM demand_table"
  key_ttl: "24h"
```
