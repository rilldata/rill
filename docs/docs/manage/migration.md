---
title: "Migrations"
description: For documenting required migrations
sidebar_label: "Migrations"
sidebar_position: 60
---

Please refer to the changes below and recommended migration steps.

## Migration

### v0.49 -> v0.50

Due to the [separation of dashboards to metrics layer and dashboards](/concepts/metrics-layer), you will need to review your current dashboards and make the following changes (note: Legacy dashboards will continue to function.):

**[Sample Legacy Dashboard Contents](https://docs.rilldata.com/reference/project-files/explore-dashboards):**

```yaml
title: #needs to be defined on metrics-view and dashboard
model: #defined on metrics-view
type: #defined on both, explore or metrics-view
timeseries: #defined on metrics-view

smallest_time_grain: #defined in metrics-view, 

default_...: #defined in dashboard
    dimensions:
    measures:
    comparison:
    ...


measures: #defined in metrics-view, 
    ...

dimensions: #defined in metrics-view, 
    ...

security: #defined on both metrics-view and dashboard but different capabilities
    ...

theme:  #defined in dashboard

available_time_zones: #defined in dashboard as time_zones:
available_time_ranges: #defined in dashboard as time_ranges:

first_day_of_week: #defined in metrics-view,
first_month_of_year: #defined in metrics-view,

```

**[Metrics_View YAML](/reference/project-files/metrics-view):**
```yaml
version: 1 #defines version 
type: metrics_view # metrics_view

title: The title of your metrics_view
display_name: The display_name
description: A description
model / table: refernce the model or table, 
database / database_schema: #if using a different OLAP engine, refers to database and schema (usually not required)

timeseries: your timeseries column

smallest_time_grain: #defines the smallest time grain 

first_day_of_week: #defines first day of week
first_month_of_year: #defines first month of year

dimensions: #your dimensions, can be copied from dashboard.yaml
    - name:
      label:
      column/expression:
      property:
      description:
      unnest:
      uri:

measures: #your measures, can be copied from dashboard.yaml
    - name:
      label:
      type:
      expressions:
      window:
      per:
      requires:
      description:
      format_preset / format_d3:
      valid_percent_of_total:

security: #your security policies can be copied from dashboard.yaml
```

**[Explore dashboard YAML](/reference/project-files/explore-dashboards):**
```yaml
type: explore

title: Title of your explore dashboard
description: a description
metrics_view: <your-metric-view-file-name>

dimensions: '*' #can use regex
measures: '*' #can use regex

theme: #your default theme

time_ranges: #was available_time_ranges
time_zones: #was available_time_zones

defaults: #define all the defaults within here
    dimensions:
    measures:
    time_range:
    comparison_mode:
    comparison_dimension:

security:
    access: #only access can be set on dashboard level, see metric view for detailed access policy
```

For any questions, please [contact the team](https://docs.rilldata.com/contact) via Slack, email, Discord or the in-app chat!