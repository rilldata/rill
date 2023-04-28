---
title: Google Cloud Storage (GCS)
description: Create a Google Cloud service account for connecting to a GCS bucket from Rill Cloud
sidebar_label: GCS
sidebar_position: 20
---

## Create a service account

Follow the steps below to create a Google Cloud service account for connecting to a GCS bucket from Rill Cloud.

### Using the Google Cloud Console

Here is a step-by-step guide on how to create a Google Cloud service account with read-only access to GCS:

1. Navigate to the [Service Accounts page](https://console.cloud.google.com/iam-admin/serviceaccounts) under "IAM & Admin" in the Google Cloud Console.

2. Click the "Create Service Account" button at the top of the page.

3. In the "Create Service Account" window, enter a name for the service account, then click "Create and continue".

4. In the "Role" field, search for and select the "Storage Object Viewer" role. Click "Continue", then click "Done".

5. On the "Service Accounts" page, locate the service account you just created and click on the three dots on the right-hand side. Select "Manage Keys" from the dropdown menu.

6. On the "Keys" page, click the "Add key" button and select "Create new key".

7. Choose the "JSON" key type and click "Create".

8. Download and save the JSON key file to a secure location on your computer.

### Using the `gcloud` CLI

1. Open a terminal window. [Install and initialize the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk) if you haven't already done so.

2. You will need your Google Cloud project ID to complete this tutorial. Run the following command to show it:
```bash
gcloud config get project
```

3. Replace `[PROJECT_ID]` with your project ID in the following command, and run it to create a new service account (optionally also replace `rill-service-account` with a name of your choice):
```bash
gcloud iam service-accounts create rill-service-account --project [PROJECT_ID]
```

3. Replace `[PROJECT_ID]` with your project ID in the following command, and run it to grant the "Storage Object Viewer" role to the service account:
```bash
gcloud projects add-iam-policy-binding [PROJECT_ID] \
    --member="serviceAccount:rill-service-account@[PROJECT_ID].iam.gserviceaccount.com" \
    --role="roles/storage.objectViewer"
```

4. Replace `[PROJECT_ID]` with your project ID in the following command, and run it to create a key file for the service account:
```bash
gcloud iam service-accounts keys create rill-service-account.json \
    --iam-account rill-service-account@[PROJECT_ID].iam.gserviceaccount.com
```

5. You have now created a JSON key file named `rill-service-account.json` in your current working directory.
