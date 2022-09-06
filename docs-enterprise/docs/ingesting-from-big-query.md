---
title: "Tutorial: BigQuery Ingestion"
slug: "ingesting-from-big-query"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Creating a dataset from a Google BigQuery table"/>

## Overview
To import data form BigQuery, you will first need to grant Rill access to your BigQuery data. Once that is complete, you'll ingest the data via the Druid console following the same steps shown in [Druid Data Ingestion](https://druid.apache.org/docs/latest/tutorials/tutorial-batch.html) and [Druid Optimization During Ingestion](https://druid.apache.org/docs/latest/ingestion/index.html)

## Grant Rill access to your BigQuery project
Starting in Rill, your workspace has a Google Service Account associated with it. You will then go to your BigQuery project and add that Google Service Account as a member with BigQuery Data Viewer permission. To keep everything in one place for this tutorial, we'll walk you through granting access here. These instructions call also be found [here](/google-bigquery).

1. Find your Google Cloud Service Account by logging into Rill and clicking on Integrations. Your Google Cloud Service Account will be displayed. It will be of the form `organization`-`workspace`@rilldata.iam.gserviceaccount.com. 

2. Go to your Google Cloud Console and select the project to which you want to grant access. 
  ![](https://images.contentful.com/ve6smfzbifwz/6QItw8AUlK7ACgoqbf0UpC/9a92c4ca0e1c68753f65660fa717c703/0fa73d4-Project_selector.png)

3. Open the sidebar menu by clicking the 3 lines button in the top left, then choose IAM & Admin then click on IAM https://console.cloud.google.com/iam-admin/iam 
  ![](https://images.contentful.com/ve6smfzbifwz/5eMuGJMf6mx8cViZ94a8BJ/6160486e5898cec6d6ef41bcb7e1bda8/dc89c3c-iam_menu_selection.png)

4. In the IAM menu click the ADD button. This will display a form where you can input the service accounts that can access your project and the permissions with which they can access it.

5. In the New members field, enter your google service account, found in step 1.  

6. Select the role `BigQuery Data Viewer`. This will permit Rill to fetch your projects tables into BigQuery.
  ![](https://images.contentful.com/ve6smfzbifwz/5eMuGJMf6mx8cViZ94a8BJ/6160486e5898cec6d6ef41bcb7e1bda8/dc89c3c-iam_menu_selection.png)

7. Click on save.

## Ingest a BigQuery table into Rill
Now that you've given Rill permission to access your BigQuery data, you'll load your BigQuery dataset into Rill.

1. Go to your BigQuery dataset, select the table you want to load, and click on Details to find the table id.  Copy the table id into your clipboard. You will use this table id when you load your data so save it away - either keep it in your clipboard or copy it into a document.
![](https://images.contentful.com/ve6smfzbifwz/3ixrM3Du9SgGqUXgQg8mka/e49a5cf7749f8b1cd2666d8883b081e7/5ce7b0f-BigQuery_table_id.png)

1. **In Rill, click on `Druid Console`**

  This is a button in the upper right of RCC. A new tab will be created for you that displays the Druid console. You'll see `Load Data`, `Ingestion`, and `Query` tabs. If you don't see the `Load Data` tab, you will need to ask your Rill admin for Editor privilege. 
2. **Click on the `Load Data`**. 
3. If you see a screen with two buttons that say `Start a new spec` and `Continue from previous spec`, choose `Start a new spec`. 
4. You should see a variety of connector tiles. **Click on the `Google BigQuery`** tile
5. **Click on `Connect Data`**
6. **Paste the table id from BigQuery that you copied into your clipboard into the Table ID field and click `Apply`.
7. You should see a preview of your data
8. Proceed through Druid Data Ingestion, clicking `Next` in the bottom right to step through the various ingestions stages. Remember to name your dataset appropriately in the `Publish` stage
9. Submit your ingestion spec in the final stage. When the status of the job says `Success`, click on the `Query` tab at the top to go to the Druid SQL console and use SQL to query your new dataset.