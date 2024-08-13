---
title: 'Big Query on a Dashboard'
sidebar_label: 'BQ to Rill'
sidebar_position: 1
hide_table_of_contents: false
---

Rill natively supports [connecting to BQ](https://docs.rilldata.com/reference/connectors/bigquery) using the BigQuery SDK.

:::note Requirements
Your GCP credentials will be inferred from your local environment. 

You will need to ensure that you have setup the local credentials via the gcloud CLI. If you have any questions, please refer to our docs, [here](https://docs.rilldata.com/reference/connectors/bigquery#local-credentials)

:::

### Adding a source to Rill
A source can be added via the UI. 

<img src = '/img/guides/Adding-Data.gif' class='rounded-gif' />
<br />


If you have your own Big Query dataset you'd like to use, please prepare the project name and SQL statement. If not, please use the following to follow along:
```
insert a public BQ link here
```

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


That's it! You have successfully deployed a dashboard from Big Query to Rill Cloud. You can know share your dashboard with your team via [public URLs](https://docs.rilldata.com/explore/bookmarks) or [inviting members](https://docs.rilldata.com/manage/user-management) to the project.


### Common Issues




import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />