---
title: "5. Create the Explore Dashboard"
sidebar_label: "5. Create the Explore Dashboard"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
---

At this point our metrics view is ready! Let's select `Create Explore Dashboard`. This will automatically populate the explore dashboard to select all the of created metrics and dimension in your metrics view. We can make changes to the view via the YAML or visual dashboard editor.

![img](/img/tutorials/103/Completed-100-dashboard.png)


## Making Changes
### Visual Explore Dashboard

![img](/img/tutorials/103/visual-dashboard-tutorial.png)

On the right panel, you are able to select measures, dimensions, time ranges, and various other components that control the view of your dashboard. In the main area, you will see a preview of what your dashboard will look like. You can also select the code view to make any needed changes and/or set more advanced settings as found in our [explore dashboard YAML reference.](https://docs.rilldata.com/reference/project-files/explore-dashboards)

### YAML
By default, the page will contain the basics parameters as seen below. You can add more advanced settings as you require for you use case.
```YAML
# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards

type: explore

title: "commits___model_metrics dashboard"
metrics_view: commits___model_metrics

dimensions: '*'
measures: '*'
```



### Explore Dasboard Components

For a quick summary on the different components of an explore dashboard, and its respective parts in the dashboard UI.

<img src = '/img/tutorials/103/simple-dashboard.gif' class='rounded-gif' />
<br />

--- 









import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />

