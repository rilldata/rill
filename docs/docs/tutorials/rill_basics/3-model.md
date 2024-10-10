---
title: "3. Create the SQL Model"
sidebar_label: "3. Create the SQL Model"
sidebar_position: 2
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
---

### What is a model?
A data model in Rill is a used to perform intermediate processing as well as any last mile ETL on the source data required. We recommend creating <a href="https://docs.rilldata.com/build/models/#one-big-table-and-dashboarding" target="_blank">"One Big Table"</a> for your dashboards.

### Let's create a model from our source data

Go ahead and select the `Create Model` button in the top right hand corner from the commits dataset.

<img src = '/img/tutorials/102/Add-Model.gif' class='rounded-gif' />
<br />

You'll see a new UI appear automatically and the contents are auto populated with some SQL. This is our Model page.

```SQL
select * from commits__
```

Let's try to make some changes to our SQL and see how the UI reacts.

```SQL
select * from commits order by author_date DESC
```
Notice that the preview table is automatically updated as we modify the SQL. This is due to our auto-save feature. In case of any errors that are encountered the UI will update accordingly and display the error.


<img src = '/img/tutorials/102/Model-SQL.gif' class='rounded-gif' />
<br />



:::tip
 
 Our Autosave feature can be enabled/disabled as needed via the rill.yaml file / project settings, or by simply selecting the toggle in the UI.

:::


### Let's merge the two tables!

Each dataset independently gives us some interesting information but in reality we want the data from both of these datasets.
- commits__ gives us information about the user who commited the changes.
- modified_files__ gives us information on the actual changes to the file and its directory.

We will grab all the columns from commits__ and only a few from modified_files__ as seen below. We will join the two datasets on the commit_hash column. As this is the SQL view that our dashboard will be based off of, we want to materialize it!

```SQL
-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models
-- @materialize: true

SELECT
    a.*,
    b.filename,
    b.added_lines,
    b.deleted_lines
FROM
    commits__ a
INNER JOIN
    modified_files__ b
ON
    a.commit_hash = b.commit_hash
```
> You can see all of the referenced source tables in the right panel. Hover over a table to see where it's referenced in the table.

### Concept: What is materialization?

By default, models created in DuckDB will be views. Both views and tables will be shown in under your DuckDB's connector table UI.


There are times where you’d rather materialize these as tables by adding the following to the SQL file

```yaml
-- @materialize: true
```
<details>
  <summary>Why materialize?</summary>
  
   You may experience some improved performance materializing SQL views for intermediate models in the case of complex SQL or large data.

    We generally recommend materializing finals models that power dashboards.

    However, you might experience some degradation of modeling experience [auto-save feature] for some specific situations including cross joins.

</details>


Ready to visualize our data?


import DocsRating from '@site/src/components/DocsRating';


---
<DocsRating />