---
title: "Let's get back to our project"
description:  Further build on project
sidebar_label: "Advanced SQL modeling"
sidebar_position: 16
---
## Advanced SQL modeling

In order to prepare for some further topics to be discussed, let's revisit our SQL model.

It is a simple join between two tables that was able to give us the user details and the commit details. But, in reality that doesn't give us much information about the repository. We can see some added and removed lines and filter based on the user and filename. Let's make a few more modifications.

Let's create a new model file `advanced_commits___mode.sql`.

:::tip
    We will add our original SQL code to a CTE (common table expression) with a few modifications then from there, create our expanded SQL query. 
:::

Without going into a whole SQL lecture in this course, we will make some modifications to the SQL in order to create some useful measures and dimensions. 
 Using the initially created table as a CTE, we will use some of <a href= 'https://duckdb.org/docs/sql/functions/regular_expressions' target="blank" >DuckDB's internal functions</a> to modify the original columns to something more useful.

- We want to remove the filename from directory path.
- Using added_lines and delete_lines, make some interesting measures.
- Using unique hash_commits, count the distint commit_msg to find unique commits.

```SQL
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
    commit_file_stats -- CTE table
WHERE
    directory_path IS NOT NULL -- removing any NULL values from the dimension
GROUP BY 
    --directory_path, filename, author_name, author_date
    ALL
ORDER BY
    author_date DESC -- ASC if wanted
    ```

The resulting SQL allows us to filter using the dimensions: `author_date`, `directory_path`, `filename` and `commit_msg`.

It gives us the following measures: `number of commits`, `net line changes`,  `total_line_changes`, `added_lines` and `deleted_lines`.

Using this model, we can create a new dashboard with more informative capabilities! On the dashboard itself, we can create some further useful measures.

:::tip
It can be confusing to be able to create measures in the SQL model layer as well as the dashboard's metric layer. This is intentional due to how the underlying OLAP engine works. For a more detailed reasoning, we have created a guide, Average of Averages, that explains further why this might not be the best to do on the SQL layer.
:::


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />