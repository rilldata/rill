---
title: "Custom Dashboards"
description: Using Template charts
sidebar_label: "Template Charts"
sidebar_position: 10
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Using Rill Custom Chart templates

Rill provides a few templates for custom charts. We will use the dashboard, `dashboard_1` to create a few custom dashboards.

<Tabs>
<TabItem value="KPI" label="KPI Charts" default>

```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts
    
type: component

kpi:
  metric_view: dashboard_1
  time_range: P1W
  measure: net_line_changes #if name parameter is defined on measure
  comparison_range: P1W
  ```
</TabItem>
<TabItem value="Bar" label="Bar Charts">

```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

type: component

data:
  metrics_sql: |
    select 
      measure_2,
      date_trunc('day', author_date) as date 
    from dashboard_1
    where author_date > '2024-07-14 00:00:00 Z'

bar_chart:
  x: date
  y: measure_2
  

  ```
  </TabItem>


  <TabItem value="Line" label="Line Charts">

  ```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

type: component

data:
    metrics_sql: |
      select 
        measure_0,
        date_trunc('day', author_date) as date 
      from dashboard_1
      where author_date > '2024-07-14 00:00:00 Z'

line_chart:
  x: date
  y: measure_0
  
    ```
  </TabItem>
</Tabs>


Additional Templates available soon : 

- Pie Chart
- Table
- Pivot Table
- Area Chart
- Scatter Plot
- Choropleth
- Layer Map


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />