---
title: General Access Control
description: How to control access to a Rill resource
tags:
- security
- code
- snippets
docs: https://docs.rilldata.com/build/metrics-view/security
hash: c8fc12dbbe6913dc6df26241a93fff924f2fed6e72b483eb880e77db91a982e3
---

```YAML
security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
```
