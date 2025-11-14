---
title: Column Level Filters in Metrics Views
tags:
- security
- code
- snippets
docs: https://docs.rilldata.com/build/metrics-view/security
hash: d59e390d79fe4d5d06a13ff309bd6566701c6a4476c64e333adaf5421be703bc
---

```YAML
security:
  exclude:
    - if: "'{{ .user.domain }}' != 'example.com'"
      names:
        - ssn
        - id
```
