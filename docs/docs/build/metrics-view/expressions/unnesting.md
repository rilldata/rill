---
title: "Unnest Dimensions"
description: Tips & Tricks for Metric Formatting
sidebar_label: "Unnest Dimensions"
sidebar_position: 06
---

### Unnest

 For multi-value fields, you can set the unnest property within a dimension. If true, this property allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match.

 ```yaml
  - label: "Example Column"
    column: multi_value_field
    description: "Unnested Column"
    unnest: true
```

Jon's example here