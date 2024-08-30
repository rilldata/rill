---
title: "Metrics View (Dashboard)"
sidebar_label: "4. Create the Dashboard"
sidebar_position: 3
hide_table_of_contents: false
---

### What is a Metrics View? 

A metrics view is a dashboard where you can visualize and drill down into your data.

### Let's create a metrics view!

Now that the model is created, we can create a metrics-view. There are two ways to do so:
1. Generate dashboard with AI
2. Start Simple using the +Add, Dashboard 

<details>
  <summary>How does Generate dashboard with AI work?</summary>
  
    We send a set of YAML and project files to GPT-4o to suggest the dimensions, measures, and various other key pairs for your dashboard. 
</details>

Let's go ahead and create a simple dashboard via the UI and build onto it. 

<img src = '/img/tutorials/102/Add-Dashboard.gif' class='rounded-gif' />
<br />


As you can see, the default dashboard YAML is as follows:

```yaml
# Dashboard YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards

type: metrics_view #the type is required for all YAML files in Rill to define the type

title: "Dashboard Title"
table: example_table # Choose a table [or model] to underpin your dashboard / 
timeseries: timestamp_column # Select an actual timestamp column (if any) from your table

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


## Fixing the Dashboard
Let's go over each component and what they are in order to better understand the metrics-view and how to fix the dashboard.

### Type ###

```yaml
type: metrics_view
```
The type is a Rill required key pair as it indicates to Rill what type of file this is. Whether a `source`, `metrics_vew`, `connector`, etc. We can keep this as is.

---

### Title and underlying table ###
```yaml
title: "My Tutorial dashboard"

#table: example_table # OR
model: commit___model #_ _ _, there are three underbars there!
```
The title will be displayed in the metrics-view UI and should be defined as required. <br />
Depending on your use-case, `table` or `model` will be used to underpin your dashboard.<br />

Let's go ahead and change the title to `My Tutorial dashboard` and comment out the table and add `model: commit___model`.


---

### Time series ###
```yaml
timeseries: author_date # Select an actual timestamp column (if any) from your table
```
The timeseries column is a Date type column within your table. While technically this is not required, many of the features require a timeseries column and your resulting dashboard will lack key functionality. Let's set this to `author_date`.

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

At this point our dashboard should be previewable! 
Let's go ahead a select preview so we can take a look. You should see something similar to the below.


![simple](/img/tutorials/103/simple-dashboard.png)


We can definitely do better than that!

--- 
For a quick summary on the different components that we modified and it's respective parts in the dashboard UI.



<img src = '/img/tutorials/103/simple-dashboard.gif' class='rounded-gif' />
<br />

<details>
  <summary>Dashboard not working?</summary>
  
    Go ahead and copy the YAML below into your dashboard.
  ```yaml
# Dashboard YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards

type: metrics_view

title: "My Tutorial Project"
#table: example_table # Choose a table to underpin your dashboard
model: commits___model

timeseries: author_date # Select an actual timestamp column (if any) from your table

dimensions:
  - column: author_name
    label: "The Author's Name"
    description: "The name of the author of the commit"

measures:
  - expression: "SUM(added_lines)"
    label: "Sum of Added lines"
    description: "The aggregate sum of added_lines column."
```
</details>

## Adding more functionalty

Let's add further dimensions and measures. For

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


You may need to reference the <a href='https://docs.rilldata.com/reference/project-files/dashboards' target="_blank">dashboard YAML </a> reference guide to figure out the above. Your final output should look something like this! 

![finished](/img/tutorials/103/Completed-100-dashboard.png)


<details>
  <summary> Working Dashboard YAML</summary>
  ```yaml
# Dashboard YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards

type: metrics_view

title: "My Tutorial Project"
#table: example_table # Choose a table to underpin your dashboard
model: commits___model

timeseries: author_date # Select an actual timestamp column (if any) from your table

dimensions:
  - column: author_name
    label: "The Author's Name"
    description: "The name of the author of the commit"

  - column: author_timezone
    label: "The Author's TZ"
    description: "The Author's Timezone"

  - column: filename
    label: "The filename"
    description: "The name of the modified filename"
 
measures:
  - expression: "SUM(added_lines)"
    label: "Sum of Added lines"
    description: "The aggregate sum of added_lines column."

  - expression: "SUM(deleted_lines)"
    label: "Sum of deleted lines"
    description: "The aggregate sum of deleted_lines column."
```

</details>


import ComingSoon from '@site/src/components/ComingSoon';


## Modifying the metrics view via the UI

<ComingSoon />

<div class='contents_to_overlay'>
**Main components of the metrics view Editor**

Now that we've discussed the YAML and have a general understanding of how it works, I want to show you how to use the metric view editor! This is our UI based editor that allows for a point and click method to modify the metric-view.



</div>



import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />

