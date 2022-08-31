---
title: "Google BigQuery"
slug: "google-bigquery"
excerpt: "Direction integration via Rill's homegrown BigQuery connector"
hidden: false
createdAt: "2020-10-21T01:49:25.004Z"
updatedAt: "2021-08-17T23:29:01.547Z"
---
# Setup instructions
Follow the instructions below to grant Rill access to your Google BigQuery datasets.

1. Find your Google Cloud Service Account by logging into Rill and clicking on Integrations. Your Google Cloud Service Account will be displayed. It will be of the form <organization\>-\<workspace\>@rilldata.iam.gserviceaccount.com. 

2. Go to your Google Cloud Console and select the project to which you want to grant access. 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/2c3627e-Project_selector.png",
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
        "https://files.readme.io/8efbbf9-IAM.png",
        "IAM.png",
        743,
        806,
        "#d3d9e3"
      ],
      "sizing": "smart"
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
        "https://files.readme.io/be5a511-Screen_Shot_2020-10-20_at_7.10.46_PM.png",
        "Screen Shot 2020-10-20 at 7.10.46 PM.png",
        519,
        520,
        "#f6f7f8"
      ],
      "sizing": "smart"
    }
  ]
}
[/block]
7. Click on save.