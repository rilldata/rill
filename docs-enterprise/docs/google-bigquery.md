---
title: "Google BigQuery"
slug: "google-bigquery"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Direction integration via Rill's homegrown BigQuery connector" />

## Setup instructions
Follow the instructions below to grant Rill access to your Google BigQuery datasets.

1. Find your Google Cloud Service Account by logging into Rill and clicking on Integrations. Your Google Cloud Service Account will be displayed. It will be of the form <organization\>-\<workspace\>@rilldata.iam.gserviceaccount.com. 

2. Go to your Google Cloud Console and select the project to which you want to grant access.
![](https://images.contentful.com/ve6smfzbifwz/4KskMcw6t4az7qdW5i9YDa/7c8fe66bdd9b02864ffd878a29031ac8/2c3627e-Project_selector.png)

3. Open the sidebar menu by clicking the 3 lines button in the top left, then choose IAM & Admin then click on IAM https://console.cloud.google.com/iam-admin/iam 
![](https://images.contentful.com/ve6smfzbifwz/5lkiJLFKP9i0mNJGVcEJpQ/f78c764249c43db1da358df842f3ef0e/8efbbf9-IAM.png)

4. In the IAM menu click the ADD button. This will display a form where you can input the service accounts that can access your project and the permissions with which they can access it.

5. In the New members field, enter your google service account, found in step 1.  

6. Select the role `BigQuery Data Viewer`. This will permit Rill to fetch your projects tables into BigQuery. 
  
  ![](https://images.contentful.com/ve6smfzbifwz/41T3D34qZmZEzFf91mhKo1/013b627be97a308698e04f50a9dccfef/be5a511-Screen_Shot_2020-10-20_at_7.10.46_PM.png)

7. Click on save.