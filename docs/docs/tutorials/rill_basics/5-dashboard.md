---
title: "5. Create the Explore Dashboard"
sidebar_label: "5. Create the Explore Dashboard"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
---

## Create the Explore dashboard
At this point our metrics view is ready! Let's rename the metrics to `commits___model_metrics.yaml` and select `Create Explore`.


This will create an explore-dashboards folder with a very simple YAML file. Let's go ahead a select preview to see what it looks like. You should see something similar to the below.


![simple](/img/tutorials/103/simple-dashboard.png)


We can definitely do better than that!

--- 
For a quick summary on the different components that we modified, and its respective parts in the dashboard UI.

<img src = '/img/tutorials/103/simple-dashboard.gif' class='rounded-gif' />
<br />


## Adding more Functionality

Let's add further dimensions and measure to the metrics layer and see the changes to the explore dashboard.

### Dimensions

From our dataset, we can add more dimensions to allow more filtering and exploration of the measures we will create.

	Add the following dimensions, with title and description.
		- author_name
		- author_timezone
		- filename

### Measures	

	We can definitely create better aggregations for some more meaningful data based on these commits.
		- sum(added_lines)
		- sum(deleted_lines)


You may need to reference the <a href='https://docs.rilldata.com/reference/project-files/explore-dashboards' target="_blank">metrics view YAML </a> reference guide to figure out the above. Your final output should look something like this! 

![finished](/img/tutorials/103/Completed-100-dashboard.png)


<details>
  <summary> Working Metrics View YAML</summary>
  ```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics_views

version: 1
type: metrics_view

table: commits___model

timeseries: author_date # Select an actual timestamp column (if any) from your table

dimensions:
  - column: author_name
    name: author_name
    label: The Author's Name
    description: The name of the author of the commit

  - column: author_timezone
    label: "The Author's TZ"
    description: "The Author's Timezone"

  - column: filename
    label: "The filename"
    description: "The name of the modified filename"

measures:
  - expression: SUM(added_lines)
    name: added_lines
    label: Sum of Added lines
    format_preset: humanize
    description: The aggregate sum of added_lines column.
    valid_percent_of_total: true

  - expression: "SUM(deleted_lines)"
    label: "Sum of deleted lines"
    description: "The aggregate sum of deleted_lines column."

```

</details>


### Completed visual metrics editor

If you decide to build out the metrics view via the UI, it should look something like below!

![img](/img/tutorials/103/visual-metric-editor.png)





import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />

