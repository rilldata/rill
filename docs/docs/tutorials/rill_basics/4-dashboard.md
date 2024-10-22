---
title: "4. Metrics View and Dashboard"
sidebar_label: "4. Create the Metrics View and Dashboard"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
---

### What is a Metrics View? 
:::note
Starting from version 0.50, we have officially split the metrics-view and dashboard and rolled out [visual metric editor](#via-the-ui) What this means is that we have a separate file for the metrics layer and a dashboard built on top of this. For more reasoning behind this change, please refer to our documentation.
:::
A metrics view is a layer in which you can create and define your measures and dimensions. Once you have defined your measures and dimensions, you can build the dashboard.


### Let's create a metrics view!

Now that the model is created, we can create a metrics-view. There are two ways to do so:
1. Generate metrics with AI
2. Start Simple using the +Add, Metrics 

<details>
  <summary>How does Generate metrics with AI work?</summary>
  
    We send a set of YAML and project files to OpenAI to suggest the dimensions, measures, and various other key pairs for your dashboard. 
</details>

Let's go ahead and create a simple metrics layer via the UI and build onto it. 


![img](/img/tutorials/102/create-metrics-view-ui.png)


As you can see, the default dashboard YAML is as follows:

```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics_views

version: 1
type: metrics_view

table: example_table # Choose a table to underpin your metrics
timeseries: timestamp_column # Choose a timestamp column (if any) from your table

dimensions:
  - column: category
    label: "Category"
    description: "Description of the dimension"

measures:
  - expression: "SUM(revenue)"
    label: "Total Revenue"
    description: "Total revenue generated"
```

For now, you'll see a red box around the UI and the preview button grayed out. This indicates something is wrong with the YAML. 


## Fixing the Metrics View
From here, we have two options! 
Starting from version 0.50, we have introduced the [visual-metric-editor](/build/metrics-view/#using-the-visual-metrics-editor), in the top right corner, you can select whether you want to modify the YAML directly or use a UI tool.

## Via the UI
When navigating to the visual metric editor, you will see the following:

![img](/img/tutorials/102/new-viz-editor.png)

We can go ahead and change the following components as directed in the UI:

- model: commits___model
- time columns: author_date
- measures: sum(added_lines)
- dimensions: author_name


![img](/img/tutorials/102/basic-viz-editor.png)

Once finished, the red border will disappear and your explore dashboard is ready to be created. If you need further information on each component see the next section, via the YAML.


## Via the YAML
Let's go over each component and what they are in order to better understand the metrics-view and how to fix the dashboard.

### Type 

```yaml
version: 1
type: metrics_view
```
The type is a Rill required key pair as it indicates to Rill what type of file this is. Whether a `source`, `metrics_vew`, `connector`, etc. We can keep this as is.

---

### Underlying Table ###
```yaml
table: commits___model # Note that this has 3 "_"! 
```

The underlying table can be defined here, let's change it to `commits___model`.

---

### Time series ###
```yaml
timeseries: author_date # Select an actual timestamp column (if any) from your table
```
The time-series column is a date type column within your table. Let's set this to `author_date`.

---
### Dimensions ###
```yaml

dimensions:
  - column: author_name #the column name in the table
    label: "The Author's Name" #A label
    description: "The name of the author of the commit" #A description, displayed when hovered over dimension

```
Dimensions are used for exploring segments and filtering the dashboard. Each dimension hve a set of required key pairs. For the `column`, you will need to set it to a column from your model or table, assign it a label (this is how it's labeled in the dashboard) and provide a description. The description will be displayed when hovering over, and can provide more context.

Our first dimension wil be the author's name. Let's go ahead and make the changes above.

---
### Measures ###

```yaml
measures:
  - expression: "SUM(added_lines)"
    label: "Sum of Added lines"
    description: "The aggregate sum of added_lines column"
```

Measure are the numeric aggreagtes of columns from your data model. These function will use DuckDB SQL aggregation functions and expressions. Similar to dimensions, you will need to create an expression based on the column of your underlying model or table.

Our first measure will be: `SUM(added_lines)`.

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

