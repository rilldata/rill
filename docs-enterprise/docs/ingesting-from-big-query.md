---
title: "Tutorial: BigQuery Ingestion"
slug: "ingesting-from-big-query"
excerpt: "Creating a dataset from a Google BigQuery table"
hidden: false
createdAt: "2020-10-28T22:06:59.598Z"
updatedAt: "2022-07-13T07:09:45.234Z"
---
# Overview
To import data form BigQuery, you will first need to grant Rill access to your BigQuery data. Once that is complete, you'll ingest the data via the Druid console following the same steps shown in [Druid Data Ingestion](https://druid.apache.org/docs/latest/tutorials/tutorial-batch.html) and [Druid Optimization During Ingestion](https://druid.apache.org/docs/latest/ingestion/index.html)

# Grant Rill access to your BigQuery project
Starting in Rill, your workspace has a Google Service Account associated with it. You will then go to your BigQuery project and add that Google Service Account as a member with BigQuery Data Viewer permission. To keep everything in one place for this tutorial, we'll walk you through granting access here. These instructions call also be found [here](doc:google-big-query).

1. Find your Google Cloud Service Account by logging into Rill and clicking on Integrations. Your Google Cloud Service Account will be displayed. It will be of the form <organization\>-\<workspace\>@rilldata.iam.gserviceaccount.com. 

2. Go to your Google Cloud Console and select the project to which you want to grant access. 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/0fa73d4-Project_selector.png",
        "Project_selector.png",
        2158,
        130,
        "#7fa5e5"
      ]
    }
  ]
}
[/block]
3. Open the sidebar menu by clicking the 3 lines button in the top left, then choose IAM & Admin then click on IAM https://console.cloud.google.com/iam-admin/iam 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/dc89c3c-iam_menu_selection.png",
        "iam_menu_selection.png",
        743,
        806,
        "#d3d9e3"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
4. In the IAM menu click the ADD button. This will display a form where you can input the service accounts that can access your project and the permissions with which they can access it.

5. In the New members field, enter your google service account, found in step 1.  

6. Select the role `BigQuery Data Viewer`. This will permit Rill to fetch your projects tables into BigQuery.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/b11ba7f-add_member.png",
        "add_member.png",
        519,
        520,
        "#f6f7f8"
      ]
    }
  ]
}
[/block]
7. Click on save.

# Ingest a BigQuery table into Rill
Now that you've given Rill permission to access your BigQuery data, you'll load your BigQuery dataset into Rill.

1. Go to your BigQuery dataset, select the table you want to load, and click on Details to find the table id.  Copy the table id into your clipboard. You will use this table id when you load your data so save it away - either keep it in your clipboard or copy it into a document.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/5ce7b0f-BigQuery_table_id.png",
        "BigQuery_table_id.png",
        660,
        406,
        "#f9f8f8"
      ]
    }
  ]
}
[/block]
1. **In Rill, click on `Druid Console`**
    This is a button in the upper right of RCC. A new tab will be created for you that displays the Druid console. You'll see `Load Data`, `Ingestion`, and `Query` tabs. If you don't see the `Load Data` tab, you will need to ask your Rill admin for Editor privilege. 
2. **Click on the `Load Data`**. 
3. If you see a screen with two buttons that say `Start a new spec` and `Continue from previous spec`, choose `Start a new spec`. 
4. You should see a variety of connector tiles. **Click on the `Google BigQuery`** tile
5. **Click on `Connect Data`**
6. **Paste the table id from BigQuery that you copied into your clipboard into the Table ID field and click `Apply`.
7. You should see a preview of your data
8. Follow the instructions in [Druid Data Ingestion](doc:druid-data-ingestion), clicking `Next` in the bottom right to step through the various ingestions stages. Remember to name your dataset appropriately in the `Publish` stage
9. Submit your ingestion spec in the final stage. When the status of the job says `Success`, click on the `Query` tab at the top to go to the Druid SQL console and use SQL to query your new dataset.