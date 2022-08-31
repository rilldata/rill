---
title: "Tutorial: Data Optimization"
slug: "druid-ingestion-optimization"
excerpt: "Optimizing performance by performing aggregation during ingestion"
hidden: true
createdAt: "2020-10-23T22:01:14.525Z"
updatedAt: "2022-06-02T22:48:10.813Z"
---
# Overview
One of the best ways to improve performance is to minimize the size of your data. During ingestion, Druid gives you the ability to perform aggregations on your data that can significantly reduce its size and improve performance. Your goal is to maintain your ability to answer all of the questions that you want to answer, while reducing the size of your data through aggregation and probabilistic approximation. You may need to iterate to find the tradeoffs that will give you the best speed with the least loss of information.

The `Configure Schema` ingestion stage gives you the power to make these tradeoffs. During this stage you can:
 * Specify what dimensions and metrics are included in or excluded from your dataset
 * Aggregate measures to a coarser level of time detail
 * Create new measures based on probabilistic approximations. For example you can create a measure that represents an approximation of a count or a unique count of a dimension

The remainder of this section will walk you through an example to show you the power of this functionality. 

# Tutorial: Aggregate data during ingestion

We'll load the same dataset that you loaded in the *Using Druid* section. To get started, please walk through the steps in that section, up through step 8, which brings you to the `Configure Schema` stage with the Sample data.

1. Review the names of the columns. Note that the time column is listed first, followed by a number of dimensions (turquoise background), followed by measures (gold background).
1. The `count` measure is special and is added by Druid. It represents the number of rows that have been aggregated. Currently there is no aggregation, so you'll see `1` in every row. 
1. The measures that follow `count` represent aggregations of the measures in the original data. For example, the wikipedia data contains five columns: `added,` `commentLength,`  `deleted`, `delta` and `deltaBucket`. Notice that each of these is prefixed by the string `sum_`. This indicates that as aggregation occurs, the value of these measures will be aggregated by summing. If this doesn't make sense yet, hold on!
1. Notice the date column on the far left. This wikipedia data is one days worth of data, at a millisecond granularity. By default the Query granularity is set in the right panel to `Hour`. Set that to `None` and view the time column, which should now show milliseconds (you can drag the column border to make the column wider). A common method of aggregating is to aggregate to a coarser level of time, such as the hour. However, aggregation can only occur if all other dimension values are the same. For example, if we have one row that has a time of 2016-06-27T00:00:01 and another that has a time of 2016-06-27T00:00:02 (second one and second two of 2016-06027), we can aggregate those two rows if all of their other values are the same. In this data set, the `page` and `diffUrl` fields are different for every single row, so in order to achieve aggregation by time, we must remove those columns. Let's do that next.

1. With Query granularity set to NONE, click on the `diffUrl` column and then click the red trash can in the lower right. Then do the same thing for the `page` column. We still aren't aggregating because our time dimension is using milliseconds. You can see this by scrolling to the count measure and noticing that it contains `1` in every column. Next we will aggregate by setting our query granularity to hour.

 1. Enter `Hour` for Query granularity and scroll over and view the count column. Note that it now contains a `5` in row 5, just to the right of user `Kolega2357`. Here you are seeing the effect of our aggregation. Scroll back to the time column and you will see that all rows show hour 0 of the date 2016-06-27. The '5' in the count for `Kolega2357` indicates that in that hour, there were five rows that "rolled up" to create the row with the count of 5. The measures to the right show similar aggregation. For example, `sum_added` now reflects the sum of the `added` field for those 5 rows.

 1. Play around with different query granularity, removing other dimensions and note how the aggregated measures change. Understanding how to aggregate your data so that you have the information that is valuable for your analytics, but the data is as compact as possible, is key to fast query performance.

 1. Sometimes you need to know the count or a unique count of a dimension, but you don't need the detailed values for the dimension. Let's demonstrate this with an example. If you have modified your data, go back to `Start` and 'Next' your way through until you get to `Configure Schema`.  Next, remove all dimensions except for `countryName` and `user`. Looking at the display,and in particular the count column and the `countryName` column, we can see that there are 2 unique users that made wiki comments from  Argentina (users 181.230.118.178 and 181.110.165.189) and 13 unique users that made comments where the countryName was `null`. One user, `Kolega23567` made 5 comments where all of the attributes (comment string, etc) were the same, so that was rolled up to a single row.

    Let's add a new metric that represents the unique number of users. In the right panel. select `Add metric`. Set the `Name` to `unique_users`, set the `Type` to `hyperUnique` and set `Filed name` to `users`. Your display should look like this:
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/8960b83-Screen_Shot_2020-10-23_at_6.29.35_PM.png",
        "Screen Shot 2020-10-23 at 6.29.35 PM.png",
        1456,
        792,
        "#3c4c51"
      ]
    }
  ]
}
[/block]
 9. Click `Apply and you should now see your new `unique_users` field at the far right of the measures

[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/6438a30-Screen_Shot_2020-10-23_at_6.31.07_PM.png",
        "Screen Shot 2020-10-23 at 6.31.07 PM.png",
        1401,
        757,
        "#3e4c50"
      ]
    }
  ]
}
[/block]
 10. Click `Next: Partition` and then `Next: tune`, and then `Next: publish`.  Choose the name of your new dataset. If you don't want to overwrite a previous copy, choose a different name.

11. Click `Submit` and then when the job is complete, click `Query` to go to the Druid SQL Console.