---
title: "Canvas Dashboards using Vega Lite"
description: Creating Canvas Dashboards in Rill
sidebar_label: "Custom Components"
sidebar_position: 12
---
Along with Rill template charts, you can create your own custom components by using Vega Lite. 

The three components that are required for a custom component using Vega Lite are:
1. `type: component` - defined on all component type YAML objects in Rill
2. `data`: - whether a `metrics_sql` or `sql`, your dashboard needs data
3. `vega_lite`: - where you will define the Vega Lite components and chart information


## Let's create a few custom components for our Canvas dashboard

Let's start by creating a chart for our Canvas dashboard.

![project-view](/img/tutorials/301/add-custom-dashboard.png)

### Creating a bar graph that calculates the top 5 contributors to the Repository

Let's use the `advanced_commits__model` table to obtain our data.

:::note Coming from Rill and ClickHouse course?
The advanced_commits_model file is created with the following contents, assuming you have already imported the data from [the first page](/tutorials/rill_basics/import).

<details>
  <summary>Click me for SQL</summary>
```sql
-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models
-- @materialize: true

WITH commit_file_stats AS (
    SELECT
        a.*,
        b.filename,
        b.added_lines,
        b.deleted_lines,
        REGEXP_EXTRACT(b.new_path, '(.*/)', 1) AS directory_path, 
    FROM
        commits__ a
    inner JOIN
        modified_files__ b
    ON
        a.commit_hash = b.commit_hash
)
SELECT
    author_date,
    cast(author_date as date) as date,
    author_name,
    directory_path,
    filename,
    STRING_AGG(DISTINCT commit_msg, ', ') AS commit_msg,

    COUNT(DISTINCT commit_hash) AS num_commits,
    SUM(added_lines) - SUM(deleted_lines) AS net_line_changes, 
    SUM(added_lines) + SUM(deleted_lines) AS total_line_changes, 

    -- (SUM(deleted_lines) / (SUM(added_lines) + SUM(deleted_lines))) as CodeDeletePercent, 
    sum(added_lines) as added_lines,
    sum(deleted_lines) as deleted_lines, 

FROM
    commit_file_stats
WHERE
    directory_path IS NOT NULL
GROUP BY 
    --directory_path, filename, author_name, author_date
    ALL
ORDER BY
    author_date DESC 
```
</details>
:::


**Preparing the Data**

In order to calculate the top contributors, we will need to grab the `author_name` and `net_line_changes` column. Then group by the `author_name` column and order by `net_line_changes` and finally filter the timeseries column, `author_date` and grab the top 5 users.

<details>
  <summary>Click me for SQL</summary>
```sql
data:
  sql: |
    select     
      author_name,
      sum(net_line_changes) as net_lines
    from advanced_commits___model
    where author_date > '2024-07-21 00:00:00'
    {{ if .args.author }} AND author_name = '{{ .args.author }}' {{ end }}

    group by author_name
    order by net_lines desc
    limit 5
```
</details>


**Preparing the Visuals**

Now that this is complete, we can create the vega_lite component. Depending on how you'd like to style your component, you can adjust the `width`, `height`, etc. However, a few components need to be set as follows:

- `"data": { "name": "table" }`,
- define the encoding to contain `x`, and `y` with the `field` corresponding to the column from the SQL statement. 
- type is set based on the [data type](https://vega.github.io/vega-lite/docs/type.html)


<details>
  <summary>Click me for component</summary>
```yaml
vega_lite: |
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "data": { "name": "table" },
    "mark": "bar",
    "width": "container",
    "height": 500,
    "encoding": {
      "y": {
        "field": "author_name",
        "type": "nominal",
        "axis": { "title": "",
                  "orient": "left" }
      },
      "x": {
        "field": "net_lines",
        "type": "quantitative",
        "axis": { "title": "# of commits" }
      }
    }
  }
```
</details>

If you are unfamiliar with creating vega-lite specs, you can use natural language to create components. Please see the following how-to for more information! [Using Natural Language to create Components](https://docs.rilldata.com/tutorials/other/custom-charts)

**Custom Component Complete**

With everything completed, you should have a YAML file with the following contents and a custom component that looks something like the following. 
![img](/img/tutorials/301/top-contributors.png)

```yaml
# Component YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/components
    
type: component

data:
  sql: |
    select     
      author_name,
      sum(net_line_changes) as net_lines
    from advanced_commits___model
    where author_date > '2024-07-21 00:00:00 Z'
    group by author_name
    order by net_lines desc
    limit 5
vega_lite: |
  {
    "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
    "data": { "name": "table" },
    "mark": "bar",
    "width": "container",
    "height": 500,
    "encoding": {
      "y": {
        "field": "author_name",
        "type": "nominal",
        "axis": { "title": "",
                  "orient": "left" }
      },
      "x": {
        "field": "net_lines",
        "type": "quantitative",
        "axis": { "title": "# of commits" }
      }
    }
  }
```




import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />