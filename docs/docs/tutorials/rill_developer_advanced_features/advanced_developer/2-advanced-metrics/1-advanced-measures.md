---
title: "Advanced Measures and Dimensions"
description:  Further build on project
sidebar_display_name: "Advanced Measures and Dimensions"
sidebar_position: 2
tags:
  - Rill Developer
  - Advanced Features
---

Now that we have another dashboard based off of our new metrics view, let's return to `advanced_metrics.yaml` and further expand on the capabilities of a metrics view. As the metrics view is backed by DuckDB (assuming you are using the default configuration), you are able to use many of the features within the metrics view. 

Let's try to implement just a few of these. 
:::tip DuckDB Documentation
Take a look at the functions listed in [DuckDB's documentation](https://duckdb.org/docs/stable/sql/functions/aggregates.html) for more information on what is possible in Rill's metrics view. If you are using ClickHouse, see the [ClickHouse Documentation.](https://clickhouse.com/docs/sql-reference/functions)
:::

## Functions in Dimensions

Using the author email, (if you haven't already added this, refer to [SQL Modeling Continued](/tutorials/rill_developer_advanced_features/advanced_developer/advanced-modeling)), lets use `string_split` to use the domain as a dimension. This allows us know to slice and dice on the domain of the user who is committing changes to the ClickHouse Repository.

```yaml
  - column: author_email
    name: author_email
    description: User Email
    display_name: User Email
    
  - expression: string_split(author_email, '@')[2]
    name: author_domain
    description: User Domain
    display_name: User Domain
```

Let's also try to unnest the directory path of the files. This gives us the ability to see the changed specifically for a certain section of code, IE: the docs. In this example, we are use `regex_split_to_array` to split the string on the `/` character. Then using the `unnest` true, we are unnesting the values of the array into a flat single selection within the dimension. 

```yaml
  - expression: regexp_split_to_array(directory_path, '/')
    display_name: "The directory unnested"
    description: "The directory path"
    name: directory_path_unnested
    unnest: true
```

<img src = '/img/tutorials/rill-advanced/new_dimensions.png' class='rounded-gif' />
<br />

## Functions in Measures

Next, let's look at the measures and what else we can calculate from the data. For one, instead of recalculating measures, you can pass a `requires:` flag into measure that can re-use measures. On the previous page, `Code Deletion Percent` is calculated using the following:
```
 SUM(deleted_lines) / (SUM(added_lines) + SUM(deleted_lines))
```
### Referencing Measures
Instead, we can re-use the measure for sum(deleted_lines) and our new measure, `total_line_changes` to calculate the percentage.
```yaml
  - expression: (deleted_lines / total_line_changes)
    requires: [deleted_lines, total_line_changes]
    display_name: "Code Deletion Percent %"
    description: "The percent of code deletion"
    format_d3: ".2%"
    name: code_deletion_percent
```

### CASE / FILTER
Another feature that is very useful is CASE / FILTER within measures. If your measure is dependent on a column being true or a specific value, you can use either to ensure that you are calculating only the required information. In the following, we want to calculate the sum of total line changes specifically for the /docs/ directory.

```yaml
  - expression: SUM(total_line_changes) FILTER (WHERE directory_path like '%docs%')
    #expression: SUM(CASE WHEN directory_path like '%docs%' THEN total_line_changes END)
    name: total_doc_line_changes
    display_name: Total number of Lines Changed
    description: the total number of lines changes, addition and deletion
    valid_percent_of_total: true
```

### Fixed Measures
While this doesn't quite make sense for our repository monitoring use-case. Let's say that there was a quota for at least 100 lines of code change per day per author. That's a lot of lines to change! We can create a fixed measure to use to calculate percent to goal. In this case, we are getting the SUM of val, set to arbitrary 100, for the list of distinct authors for the time range that we are viewing. Then using this, we can calculate the percentage of total line change to guarantee. 

```yaml
  - name: guarantee_line_change
    display_name: Guaranatee number of Line Changes
    description: Totally Random Metric.
    expression: > 
              SELECT SUM(a.val) AS value 
                FROM (
                    SELECT unnest(
                        list(distinct {
                            key: author_name,
                            val: 100
                             })
                    ) a
                )
    format_preset: humanize
    valid_percent_of_total: false

  - name: percent_to_guarantee
    requires: [guarantee_line_change,total_line_changes]
    expression: total_line_changes/guarantee_line_change
    format_preset: percentage
    display_name: Percent Lines removed vs. Guarantee
```

<img src = '/img/tutorials/rill-advanced/new_measures.png' class='rounded-gif' />
<br />

<details>
  <summary> Working Metrics View</summary>
```yaml
# Metrics View YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics_views

version: 1
type: metrics_view

table: advanced_commits___model # Choose a table to underpin your metrics
timeseries: author_date # Choose a timestamp column (if any) from your table

dimensions:
  - column: author_email
    name: author_email
    description: User Email
    display_name: User Email
    
  - expression: string_split(author_email, '@')[2]
    name: author_domain
    description: User Domain
    display_name: User Domain
    
  - column: directory_path
    display_name: "The directory"
    description: "The directory path"
    name: directory_path

  - expression: regexp_split_to_array(directory_path, '/')
    display_name: "The directory unnested"
    description: "The directory path"
    name: directory_path_unnested
    unnest: true

  - column: filename
    display_name: "The filename"
    description: "The name of the modified filename"
    name: filename

  - column: author_name
    display_name: "The Author's Name"
    description: "The name of the author of the commit"
    name: author_name

  - column: commit_msg
    display_name: "The commit message"
    description: "The commit description attached."
    name: commit_msg

measures:
  - expression: sum(deleted_lines)
    name: deleted_lines
    display_name: Total number of Deleted Lines Changed
    description: the total number of lines changes, addition and deletion
    valid_percent_of_total: true

  - expression: sum(added_lines)
    name: added_lines
    display_name: Total number of Added Lines 
    description: the total number of lines changes, addition and deletion
    valid_percent_of_total: true

  - name: p99_quantile_added_lines
    expression: QUANTILE_CONT(added_lines, 0.99)
    format_d3: ".4f"
    description: P95 of Added Lines

  - expression: SUM(total_line_changes) FILTER (WHERE directory_path like '%docs%')
    #expression: SUM(CASE WHEN directory_path like '%docs%' THEN total_line_changes END)
    name: total_doc_line_changes
    display_name: Total number of Lines Changed [Docs]
    description: the total number of lines changes, addition and deletion
    valid_percent_of_total: true

  - expression: SUM(total_line_changes)
    name: total_line_changes
    display_name: Total number of Lines Changed 
    description: the total number of lines changes, addition and deletion containing the word "fix"
    valid_percent_of_total: true

  - expression: "SUM(net_line_changes)"
    display_name: "Net number of Lines changed"
    description: "the total net number of lines changes"
    name: net_line_changes

  - expression: "AVG(net_line_changes)"
    display_name: "AVG number of Lines changed"
    description: "the AVG net number of lines changes"
    name: avg_net_line_changes
    treat_nulls_as: 1

  - name: net_change_rolling
    display_name: 3 Day Rolling Avg Net Line Change
    expression: "AVG(net_line_changes)"
    requires: [net_line_changes]
    window:
      order: "author_date"
      frame: RANGE BETWEEN INTERVAL 3 DAY PRECEDING AND CURRENT ROW
    treat_nulls_as: 1


  - expression: "SUM(num_commits)"
    display_name: "Number of Commits"
    description: "The total number of commits"
    name: num_commits

  - expression: (deleted_lines / total_line_changes)
    requires: [deleted_lines, total_line_changes]
    display_name: "Code Deletion Percent %"
    description: "The percent of code deletion"
    format_d3: ".2%"
    name: code_deletion_percent

  - name: guarantee_line_change
    display_name: Guaranatee number of Line Changes
    description: Totally Random Metric.
    expression: > 
              SELECT SUM(a.val) AS value 
                FROM (
                    SELECT unnest(
                        list(distinct {
                            key: author_name,
                            val: 100
                             })
                    ) a
                )
    format_preset: humanize
    valid_percent_of_total: false

  - name: percent_to_guarantee
    requires: [guarantee_line_change,total_line_changes]
    expression: total_line_changes/guarantee_line_change
    format_preset: percentage
    display_name: Percent Lines removed vs. Guarantee

smallest_time_grain: day

```
</details>

----

There are many other functions and use-cases that you can apply in your metrics view. Take a look at our [documentation on advanced measures](/build/metrics-view/advanced-expressions/), and [DuckDB Function Documentation](https://duckdb.org/docs/stable/sql/functions/aggregates.html) or [ClickHouse Function Documentation](https://clickhouse.com/docs/sql-reference/functions) for more information! 