---
title: "GCS Bucket"
slug: "gcs-bucket"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Batch ingestion: integrating Rill with Google Cloud Storage" />

## Setup instructions
Follow the instructions below to grant Rill access to your Google Cloud Storage Bucket.

1. Find your Google Cloud Service Account by logging into Rill and clicking on Integrations. Your Google Cloud Service Account will be displayed. It will be of the form `organization`-`workspace`@rilldata.iam.gserviceaccount.com.

2. Go to Storage Console: https://console.cloud.google.com/storage/browser.

3. Click on the bucket to which you want to grant access (click on the bucket name itself, not the checkbox).

4. Select Permissions tab.
![](https://images.contentful.com/ve6smfzbifwz/4YwoXZUqT2BuTwEvBsG6OA/6b70d11103a3921e64d54a05d99746f2/3df9887-bucket_select_permissions.png)

5. Click Add to open the modal to add members to your bucket.
![](https://images.contentful.com/ve6smfzbifwz/2Ki9BiKaHYMivZ5DPTiwbd/762b2a071d3d6fb58a1b08fd13973dc2/8fa34b8-permissions_add.png)

6. In the New members field, enter your google service account. You can find your google service account in the Settings page for your workspace in RCC. This will typically have the form  \<organization\>-\<workspace\>@rilldata.iam.gserviceaccount.com.
![](https://images.contentful.com/ve6smfzbifwz/50nIholwjMFJkaMTw8bMjy/c3334709d2eb6c8516e056f72f424957/42d2803-new_members_modal.png)

7. Select the role Cloud Storage -> Storage Object Viewer. 
![](https://images.contentful.com/ve6smfzbifwz/7HHypfag0BAVHmLegVeKuJ/2bb32c99aa57abcb9ad10a4de1053b46/46c12ce-select_role_storage_viewer.png)
