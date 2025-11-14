---
title: Row Level Filters
tags:
- security
- code
- snippets
docs: https://docs.rilldata.com/build/metrics-view/security
hash: d5741e75d4bc8755811d09f18efd1d904699172cf49f198146a8720c9f916724
---

```YAML
security:
  row_filter: region = '{{ .user.region }}'
```
