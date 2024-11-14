---
title: "Dashboard with more functionality"
description:  Further build on project
sidebar_label: "Advanced Dashboarding"
sidebar_position: 16
---

## Let's make a new dashboard

As we have learned in the previous course, we will need to set up the metrics view based on the new column names. 
Let's create a new metrics view via the UI. It should be named `advanced_metrics-view.yaml`. Let's copy the contents from our old dashboard and make some changes.

First, we will want to change the `table` value to the new model name `advaned_commits___model`

Add two new dimensions: `directory path` and `commit_msg`.

Add four new measures: `SUM(total_line_changes)`, `SUM(net_line_changes)`, `SUM(num_commits)` and lastly let's create a percentage Code Deletion measure.

On the `SUM(net_line_changes)` measure, add the following `name: net_line_changes`. While name is not required, this can be used by other components for reference, which will be discussed later.

### Creating a measure in the metric-view
Like the SQL Model, our dashboards also use the same OLAP engine and you can use aggregates or expressions to create new metrics. In our case, since we have the added_lines and delete_lines measures, we can create a percentage of lines deleted measure.

```
 SUM(deleted_lines) / (SUM(added_lines) + SUM(deleted_lines))
```
:::tip
When to create measures in the SQL Model layer vs the metric-view layer?
It depends.

Depending on the size of data, type of measure, and what you are caluclating, you can choose either. Sometimes it would be better if you are dealing with a lot of data to front load the calculation on the SQL level so your dashboards load faster. However, the way OLAP engines work (linke avg of avg article), you might get incorrect data by doing certain calculations in the SQL level. You'll have to test and see which works for you!
:::


<details>
  <summary> Example Working metrics view</summary>
```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics_views

version: 1
type: metrics_view

table: advanced_commits___model # Choose a table to underpin your metrics
timeseries: author_date # Choose a timestamp column (if any) from your table

dimensions: 
- column: directory_path
  label: "The directory"
  description: "The directory path"
  name: directory_path

- column: filename
  label: "The filename"
  description: "The name of the modified filename"
  name: filename

- column: author_name
  label: "The Author's Name"
  description: "The name of the author of the commit"
  name: author_name

- column: commit_msg
  label: "The commit message"
  description: "The commit description attached."
  name: commit_msg

measures:
- expression: "SUM(total_line_changes)"
  label: "Total number of Lines changed"
  description: "the total number of lines changes, addition and deletion"
  name: total_line_changes

- expression: "SUM(net_line_changes)"
  label: "Net number of Lines changed"
  description: "the total net number of lines changes"
  name: net_line_changes


- expression: "SUM(num_commits)"
  label: "Number of Commits"
  description: "The total number of commits"
  name: num_commits

- expression: "(SUM(deleted_lines)/(SUM(deleted_lines)+SUM(added_lines)))"
  label: "Code Deletion Percent %"
  description: "The percent of code deletion"
  format_preset: percentage
```
</details>

### Create the dashboard

Similarly to the Basics course, we can create an explore-dashboard on top of this metrics view by selecting `Create explore`. You're preview should look something like the below!

![img](/img/tutorials/204/advanced-dashboard.png)

Along with the dimensions and measures, you can define `theme:`, time zones, time ranges, and [security policies](https://docs.rilldata.com/manage/security). Feel free to test by uncommenting the parameters and seeing how it changes the explore dashboard.

```yaml
# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explores

type: explore

title: "advanced_metrics_view dashboard"
metrics_view: advanced_metrics_view

dimensions: '*'
measures: '*'

# theme: theme.yaml

#time_ranges: 
#  - PT15M // Simplified syntax to specify only the range
#  - PT1H
#  - PT6H
#  - P7D

#time_zones:
#  - America/New_York

#security:
#  access: "{{ .user.admin }} AND '{{ .user.domain }}' == 'rilldata.com'"
```

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />