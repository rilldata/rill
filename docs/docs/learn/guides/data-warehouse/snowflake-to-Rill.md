---
title: 'Snowflake on a Dashboard'
sidebar_label: 'Snowflake to Rill'
sidebar_position: 1
hide_table_of_contents: false
---

Rill natively supports [connecting to Snowflake](https://docs.rilldata.com/reference/connectors/snowflake) using the Go Snowflake driver.

:::note Requirements
You will either need to pass the snowflake when starting Rill developer via the --var flag, or via the .env file. You will need to review our snowflake docs for further information.
```
rill start --var connector.snowflake.dsn=<username>:<password>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>
```
:::

### Adding a source to Rill
A source can be added via the UI.
<img src = '/img/guides/Adding-Data.gif' class='rounded-gif' />
<br />

Please insert the SQL select statment in the UI and select `add data`. If you haven't already defined the Snowflake dsn you can do so here, as well. 

:::note DuckDB
Rill uses DuckDB as the underlying OLAP engine, and will ingest your data to the locally running database. For more information please refer to our documentation, [here](https://docs.rilldata.com/build/olap/).
:::
### Create a Model from the source
Once you've imported the dataset, you can [create a model](https://docs.rilldata.com/build/models/) and use SQL select statements for any last minute transformations. If not needed, you can skip this step and [create a dashboard](#visual-the-data).

<img src = '/img/guides/Add-Model.gif' class='rounded-gif' />
<br />

:::tip materialize model
Models in Rill are, by default, created as views in the underlying database. We recommend [materializing the final table](https://docs.rilldata.com/reference/project-files/models#model-materialization) your dashboard is built off of for improved performance. 
:::



### Visual the data 
Once you're ready to visualize the dataset, you can create a dashboard! An easy way to get started is to select the `Generate dashboard with AI`. This will automatically build you a dashboard and navigate you to the metrics view page. On the top right, you can `preview` the dashboard.

<img src = '/img/guides/generate-ai-dashboard.gif' class='rounded-gif' />
<br />


### Sharing your dashboard
To share you dashboard, all you need to do is select `deploy` and follow the steps in the UI to deploy the dashboard to Rill Cloud.

<img src = '/img/guides/deploy-ui.gif' class='rounded-gif' />
<br />


That's it! You have successfully deployed a dashboard from Snowflake to Rill Cloud. You can know share your dashboard with your team via [public URLs](https://docs.rilldata.com/explore/bookmarks) or [inviting members](https://docs.rilldata.com/manage/user-management) to the project.


## Common Issues

### I am unable to connect to Snowflake
If you are experiencing this issue on Developer, you will need to ensure that you have either defined the connection via .env or passing the dsn via a variable. 

If you are experiencing the issue in Rill Cloud, you will need to check that the connection parameters are passed into Rill cloud. If you have not done so already, please run the following in the CLI:
```
rill env configure
```


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />