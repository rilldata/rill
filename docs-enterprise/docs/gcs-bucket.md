---
title: "GCS Bucket"
slug: "gcs-bucket"
excerpt: "Batch ingestion: integrating Rill with Google Cloud Storage"
hidden: false
createdAt: "2020-05-31T17:26:09.708Z"
updatedAt: "2021-08-17T23:29:18.894Z"
---
# Setup instructions
Follow the instructions below to grant Rill access to your Google Cloud Storage Bucket.

1. Find your Google Cloud Service Account by logging into Rill and clicking on Integrations. Your Google Cloud Service Account will be displayed. It will be of the form `organization`-`workspace`@rilldata.iam.gserviceaccount.com.

2. Go to Storage Console: https://console.cloud.google.com/storage/browser.

3. Click on the bucket to which you want to grant access (click on the bucket name itself, not the checkbox).

4. Select Permissions tab.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/3df9887-bucket_select_permissions.png",
        "bucket_select_permissions.png",
        2394,
        1004,
        "#e6ebf4"
      ]
    }
  ]
}
[/block]
5. Click Add to open the modal to add members to your bucket.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/8fa34b8-permissions_add.png",
        "permissions_add.png",
        2394,
        1004,
        "#e6ebf4"
      ]
    }
  ]
}
[/block]
6. In the New members field, enter your google service account. You can find your google service account in the Settings page for your workspace in RCC. This will typically have the form  \<organization\>-\<workspace\>@rilldata.iam.gserviceaccount.com.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/42d2803-new_members_modal.png",
        "new_members_modal.png",
        2072,
        974,
        "#c6c8cc"
      ]
    }
  ]
}
[/block]
7. Select the role Cloud Storage -> Storage Object Viewer. 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/46c12ce-select_role_storage_viewer.png",
        "select_role_storage_viewer.png",
        1136,
        738,
        "#f6f7f7"
      ]
    }
  ]
}
[/block]