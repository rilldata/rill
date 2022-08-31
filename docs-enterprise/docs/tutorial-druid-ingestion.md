---
title: "Tutorial: Manual Batch Ingestion"
slug: "tutorial-druid-ingestion"
hidden: false
createdAt: "2021-08-11T22:56:54.932Z"
updatedAt: "2022-06-02T22:49:16.353Z"
---
# Tutorial: Load Sample Data into Druid

To get comfortable with Druid, we'll walk you through loading a sample data set. Normally you will specify a path to your data (for example, a BigQuery table), but in this example, you won't have to provide a path since it is build into this example. Note that since you are creating a dataset, you will need Editor privilege to access the Load Data tab referenced in this tutorial.

1. **Click on `Druid Console`**
    This is a button in the upper right of RCC. A new tab will be created for you that displays the Druid console. You'll see see `Load Data`, `Ingestion`, and `Query` tabs.
2. **Click on the `Load Data`**
     If you've loaded data recently, you will see two buttons that give you the choice of "Start a new spec" or "Continue from previous spec".  If you see these buttons, click Start a new spec.
     You should now be looking at a panel of tiles that represent your choices for where you will load data from.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/9a44962-Screen_Shot_2021-06-08_at_4.23.17_PM.png",
        "Screen Shot 2021-06-08 at 4.23.17 PM.png",
        1394,
        778,
        "#2c3046"
      ],
      "sizing": "full"
    }
  ]
}
[/block]
If you were loading your own data you would now click on a data source such as `Google BigQuery` and then you would specify a path to your BigQuery table. In this example data, we will load sample data that is available within Druid,
3.  **Click on `Example data` and then `Load example`**
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/f02cc48-connector_tiles_example_highlighted.png",
        "connector_tiles_example_highlighted.png",
        1389,
        797,
        "#2d3147"
      ]
    }
  ]
}
[/block]
4. This loads the example data and displays your data, giving you a chance to verify that the data is what you expect. In this example we are looking at a wikipedia dataset.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/e6e3af4-preview_data.png",
        "preview_data.png",
        1510,
        893,
        "#4f5469"
      ]
    }
  ]
}
[/block]
5. The tabs along the top of the page:  `Start`, `Connect`, `Parse data`, ... `Publish` represent stages in the ingestion process. In this example you will move from stage to stge by clicking the `Next` button in the bottom right of the page. Each time you click the `Next` button, Druid will move to the next stage, making its best guess as to the appropriate parameters.The highlighted tab at the top will indicate the stage you have just moved to and if you want to re-execute that stage with different parameters, you can change the parameters in the form at the right amd then click `Apply`. 

In the next steps you will walk through these stages of the ingestion process by clicking the button in the bottom right (currently `Next: Parse date`), but you can also move back and forth among the steps by clicking on the tabs at the top.
 
6. **Click on `Next: Parse data`**
    The data loader parses the data based on its best guess about the type and displays a preview of the data. In this case the data is json and it chooses json, as shown by the `input format` field. 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/41bbfe7-data_parsed.png",
        "data_parsed.png",
        1509,
        898,
        "#303548"
      ]
    }
  ]
}
[/block]
7. If the data was not json, you could change this and click apply. Click on the `input format` field to get a sense of the other choices, and feel free to click apply, but make sure JSON is selected and the display is showing your data before you proceed to the next step.

8. **Click `Next: Parse time`** 
    This step analyzes the data to identify a time column, and moves that column to the far left, with the column name __time. You can see that it is coming from the column originally labeled 'time'.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/264cffe-time_parsed.png",
        "time_parsed.png",
        1510,
        894,
        "#32374c"
      ]
    }
  ]
}
[/block]
9. Druid requires that you specify a timestamp column and it does optimizations based on this column. If your data does not have a timestamp column, you can select `Constant value`, or if your data has multiple timestamp columns, this is your opportunity to select a different one. If you specify a different time column than the default, you click `Apply` to apply your new setting. In this example, the data loader determines that the `time` column is the only candidate to be used as a time column. That's our only time column in this data set, so we leave it set as is.

10. **Click `Next: Transform`** 
    In this step we have the opportunity to transform  one or more columns or add new columns. 
      
    Let's add a new column. In the panel on the right, click `Add column transform`. In the form that expands, set `Name` to `comment_prefix`, leave the `Type` field as is, and change the value of the third field, `Expression` to `substring("comment", 0, 6)`, then click `Apply`. This creates a new field called `comment_prefix` that contains the first six characters of the comment field. You can alternately transform an existing column in place by putting the column name in the `Name` field. 

[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/46b79b1-add_column_transform.png",
        "add_column_transform.png",
        1509,
        894,
        "#313547"
      ]
    }
  ]
}
[/block]
11. After clicking `Apply`, you will see the new column to the right of the `Comment` column. 

[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/73b9724-column_transform_added.png",
        "column_transform_added.png",
        1507,
        892,
        "#2e3244"
      ]
    }
  ]
}
[/block]
12. You have full SQL available to you for making these transformations. To see the expressions and functions available, click on the little `i` with the circle around it and then click on the `expression` link. This will bring up
    
    https://druid.apache.org/docs/0.20.0/misc/math-expr.html
    
    which describes the expression syntax.

13. **Click on `Next: Filter`**
    Here you have the opportunity to filter out rows. To see the syntax, click on `filter` link in the help, which will take you to 
    
    https://druid.apache.org/docs/0.20.0/querying/filters.html. 
    
    We'll skip doing any filters here.

14. **Click on `Next: Configure Schema`**. 
    This takes you to a stage where you have the ability to do all of the following:
    + specify what dimensions and metrics are included in your dataset
    + aggregate measures to a coarser level of detail
    + create new measures that represent aggregations based on Sketches or HLL. For example you can create a measure that represents an approximation of a count or a unique count of a dimension. 

    Together these options allow you to make optimizations that can significantly improve performance. In the next steps, we'll walk through some of these options. For a tutorial on how to do this, read the next section, *Aggregation in Druid*.  For now we'll leave the data as is.

 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/0141846-configure_schema.png",
        "configure_schema.png",
        1510,
        892,
        "#303c4f"
      ]
    }
  ]
}
[/block]
     
15. **Click `Next: Partition`** 
    Here in the partition panel you can choose an optimal way to segment your data across the druid cluster. You'll be segmenting based on your time dimension and you can segment at the same granularity as your time aggregation or a coarser granularity. For example, if you've chosen to aggregate your data to the hour, you can choose to physically segment it by hour, by day, by week or by month. For now leave it set to the default, `Hour`'.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/d2f6c4f-partition.png",
        "partition.png",
        1510,
        899,
        "#282d41"
      ]
    }
  ]
}
[/block]

18. **Click `Next: tune`**
     This allows you to tune the ingestion. Skip this step.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/401f64e-tune.png",
        "tune.png",
        1511,
        892,
        "#282c40"
      ]
    }
  ]
}
[/block]
20. **Click `Next: publish`**
     Here you can choose the name of your new dataset by filling in the `Datasource name` field.  
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/c3af47f-publish.png",
        "publish.png",
        1509,
        894,
        "#282c40"
      ]
    }
  ]
}
[/block]
21. **Click `Next: Edit Spec`**
     The json representation of the spec that you just created is displayed. You could edit this by hand (and you can also generate this by hand). Right now we want to go ahead and create our dataset based on the ingestion spec as is, so we won't make any changes.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/3542288-spec.png",
        "spec.png",
        1509,
        897,
        "#24293b"
      ]
    }
  ]
}
[/block]


 22. **Click `Submit`** 
  You are taken to a panel that shows that status of your job. It will first show the job as running, and then when it's done, the "Running" string will change to "Success". Once it shows success, you can query your data.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/9dc059d-running.png",
        "running.png",
        1509,
        892,
        "#2e3245"
      ]
    }
  ]
}
[/block]

23. **Click `Query` ** (rightmost tab at the top)
  This takes you to the Druid SQL console where you can use SQL to query your data.  For example
[block:code]
{
  "codes": [
    {
      "code": "SELECT\n    countryName,\n    COUNT(*) AS \"Count\"\nFROM \"wikipedia\"\nGROUP BY countryName\nORDER BY \"Count\" DESC",
      "language": "sql"
    }
  ]
}
[/block]
24. From here you can query your data using Druid SQL. Note that by default `Smart query limit` is set to 100. If you want more than 100 rows, turn this toggle off and use the `limit` SQL expression to specify your own limit. A description of Druid's SQL language can be found here: https://druid.apache.org/docs/0.20.0/querying/sql.html