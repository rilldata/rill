---
title: "Canvas Dashboards"
description: Using Template charts
sidebar_label: "Template Components"
sidebar_position: 10
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Using Rill Custom Components Templates

Rill provides a few templates for Rill-Authored charts. We will use the metrics_view, `advanced_metrics_view` to create a few custom components. For a more extensive list of examples, please see [our reference page](https://docs.rilldata.com/reference/project-files/components#Examples)!

### Components:

<Tabs>

<TabItem value="KPI" label="KPI" default>

```yaml
# Component YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/components
    
type: component

kpi:
  metrics_view: advanced_metrics_view
  time_range: P1W
  measure: net_line_changes #if name is defined
  comparison_range: P1W

  ```
</TabItem>

<TabItem value="Bar" label="Bar Charts - Rill Authored">

```yaml
# Component YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/components

type: component

data:
  metrics_sql: |
    select 
      measure_2,
      date_trunc('day', author_date) as date,     
    from advanced_metrics_view
    where author_date > '2024-07-14'

bar_chart:
  x: date
  y: measure_2
  ```
  </TabItem>
</Tabs>





import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />