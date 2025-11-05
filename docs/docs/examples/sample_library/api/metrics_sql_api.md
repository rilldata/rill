---
title: Metrics SQL API Example YAML
tags:
- code
- complete_file
hash: 22ec94dab7cf3601e857492020ae2fe3c34f09525fe8c19bec4692bc02803c8f
---

```yaml
# API YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/apis
# Test your API endpoint at http://localhost:9009/v1/instances/default/api/<filename>

type: api

metrics_sql: |
  select
    date_trunc('day', __TIME_TIME) as day,
    revenue_received,
    cross_sell_conversions,
    ad_spend,
    cross_sell_impressions,
    cross_sell_clicks
    from order_embedded_dashboard
    where
    1=1
  {{ if not (hasKey .args "startDate") }} and  __TIME_TIME > time_range_start('P30D') {{ end }}
  {{ if (hasKey .args "startDate") }} and day >= '{{ .args.startDate }}' {{ end }}
  {{ if (hasKey .args "endDate") }} and day <= '{{ .args.endDate }}' {{ end }}
```
