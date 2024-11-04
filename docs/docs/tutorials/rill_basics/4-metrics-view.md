---
title: "4. Create the Metrics View"
sidebar_label: "4. Create the Metrics View"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
---

### What is a Metrics View? 
:::note
Starting from version 0.50, we have officially split the metrics view and dashboard and rolled out [visual metric editor](#via-the-ui) What this means is that we have a separate file for the metrics layer and a dashboard built on top of this. For more reasoning behind this change, please refer to our documentation.
:::
A metrics view is a layer in which you can create and define your measures and dimensions. Once you have defined your measures and dimensions, you can build the dashboard.


### Let's create a metrics view!

Now that the model is created, we can create a metrics view. There are two ways to do so:
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

## Via the Visual Metrics Editor
In the top right of the UI, select the viz button to navigate to the visual metrics editor. The below is an example of a completed visual metrics editor. We will make some modifcations to our current file to build something similar.

![img](/img/tutorials/102/new-viz-editor.png)

We can go ahead and change the following components as directed in the UI:

- `model`: commits___model
- `time column`: author_date
- `measures`: sum(added_lines)
- `dimensions`: author_name


![img](/img/tutorials/102/basic-viz-editor.png)

Once finished, the red border will disappear and your explore dashboard is ready to be created. If you need further information on each component see the next section, via the YAML.


## Via the YAML
Let's go over each component and what they are in order to better understand the metrics view and how to fix the dashboard.

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
