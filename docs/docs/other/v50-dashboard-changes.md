---
title: "Dashboard split into two components"
description: For documenting required migrations
sidebar_label: "Changes to Dashboards"
sidebar_position: 20
---
As we continue to develop more features within Rill, it became clear that we needed to separate the dashboard into two components. 
1. Metrics view
2. Dashboard configuration

### Historically, in Rill...
<img src = '/img/concepts/metrics-view/old-dashboard.png' class='rounded-gif' />
<br />

Historically in Rill, the metrics layer and dashboard configuration were a single file. As seen above, the metrics would be defined **inside** a dashboard YAML file along with the dashboard components and dashboard customizations. We found that this was not the best approach as we continued development. In order to create a metrics layer in Rill as a first class resource and not a consequence of dashboards, we found it necessary to split the two resources into their own files. Thus, the metrics view was born.

## Splitting the Dashboard into two components, Metrics view and Dashboard Configuration
Splitting the metrics view into its own component allows us more freedom to continue building Rill and adding new additional features. Instead of querying a dashboard for data, we would be querying the metrics-layer. The dashboard will directly query the metrics view along with many new components that are currently being developed.

### New Metrics View as an independent object in Rill 

<img src = '/img/concepts/metrics-view/metrics-view-components.png' class='rounded-gif' />
<br />


### (Explore) Dashboard

With the split of metrics view, dashboard configurations experienced an overhaul. Instead of defining measure and dimensions, you will now reference the object into your dashboard. What this allows is creating customized dashboards for specific viewers and reusability of a single metrics view in multiple dashboards!

<img src = '/img/concepts/metrics-view/explore-dashboard.png' class='rounded-gif' />
<br />


## How to migrate your current Dashboards

### version 0.49 -> version 0.5X

Due to the [separation of dashboards to metrics layer and dashboards](/build/metrics-view), you will need to review your current dashboards and make the following changes (note: Legacy dashboards will continue to function.):

**[Sample Legacy Dashboard Contents](https://docs.rilldata.com/reference/project-files/explore-dashboards):**

```yaml
title: #defined on metrics view and dashboard
model: #defined on metrics view
type: #defined on both, explore or metrics view
timeseries: #defined on metrics view

smallest_time_grain: #defined in metrics view, 

default_dimensions:  #separate default
default_measures:    #values defined in
default_comparisons: #dashboard config
...



measures: #defined in metrics view, 
    ...

dimensions: #defined in metrics view, 
    ...

security: #defined on both metrics view and dashboard but different capabilities
    ...

theme:  #defined in dashboard

available_time_zones: #defined in dashboard as time_zones:
available_time_ranges: #defined in dashboard as time_ranges:

first_day_of_week: #defined in metrics view,
first_month_of_year: #defined in metrics view,

```
---
**[Metrics View YAML](/reference/project-files/metrics-views):**

Please check the reference for the required parameters for a metrics view.
```yaml
version: 1 #defines version 
type: metrics_view # metrics_view

title: The title of your metrics_view
display_name: The display_name
description: A description
model / table: reference to the model or table, 
database / database_schema: #if using a different OLAP engine, refers to database and schema (usually not required)

timeseries: your timeseries column

smallest_time_grain: #defines the smallest time grain 

first_day_of_week: #defines first day of week
first_month_of_year: #defines first month of year

dimensions: #your dimensions, can be copied from dashboard.yaml
    - name:
      display_name:
      column/expression:
      property:
      description:
      unnest:
      uri:

measures: #your measures, can be copied from dashboard.yaml
    - name:
      display_name:
      type:
      expression:
      window:
      per:
      requires:
      description:
      format_preset / format_d3:
      valid_percent_of_total:

security: #your security policies can be copied from dashboard.yaml
```

**[Explore dashboard YAML](/reference/project-files/explore-dashboards):**

Please check the reference for the required parameters for an explore dashboard.

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