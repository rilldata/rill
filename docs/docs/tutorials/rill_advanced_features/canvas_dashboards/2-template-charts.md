---
title: "Canvas Dashboards"
description: Using Template charts
sidebar_label: "Template Components"
sidebar_position: 10
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Using Rill Custom Components Templates

Rill provides a few templates for Rill-Authored charts. We will use the dashboard, `dashboard_1` to create a few custom dashboards. For a more extensive list of examples, please see [our reference page](https://docs.rilldata.com/reference/project-files/components#Examples)!

### Charts:

<Tabs>

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
  <TabItem value="Pie" label="Pie Charts">

  ```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

type: component


    ```
  </TabItem>

      <TabItem value="Area" label="Area Charts">

  ```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

type: component


    ```
  </TabItem>


      <TabItem value="Scatter" label="Scatter Plots">

  ```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

type: component


    ```
  </TabItem>


</Tabs>

### Others:

<Tabs>
<TabItem value="KPI" label="KPI" default>

```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts
    
type: component

kpi:
  metrics_view: dashboard_1
  time_range: P1W
  measure: net_line_changes #if name parameter is defined on measure
  comparison_range: P1W
  ```
</TabItem>
<TabItem value="Layer" label="Layer Map">

  ```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

type: component


    ```
  </TabItem>
<TabItem value="Choropleth" label="Choropleth Charts">

  ```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

type: component


    ```
  </TabItem>
      <TabItem value="Table" label="Table">

  ```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts
    
type: component

table:
  measures:
    - net_line_changes
  metrics_view: "dashboard_1"
  time_range: "P3M"
  #comparison_range: "P3M"

  row_dimensions:
    - author_name
  #col_dimensions:
  #  - filename 
    ```
  </TabItem>
    <TabItem value="Pivot Table" label="Pivot Table">

  ```yaml
# Chart YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/charts

type: component


    ```
  </TabItem>
</Tabs>



import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />