---
title: Google Cloud Storage (GCS)
description: Connect to data in GCS
sidebar_label: GCS
sidebar_position: 15
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview
[Google Cloud Storage (GCS)](https://cloud.google.com/storage/docs/introduction) is a scalable, fully managed, and highly reliable object storage service offered by Google Cloud, designed to store and access data from anywhere in the world. It provides a secure and cost-effective way to store data, including common data storage formats such as CSV and Parquet. You can connect to GCS using the provided [Google Cloud Storage URI](https://cloud.google.com/bigquery/docs/cloud-storage-transfer-overview#google-cloud-storage-uri) of your bucket to retrieve and read files.


## Connect to GCS

To connect to Google Cloud Storage, you need to provide authentication credentials. You have three options:

1. **Use Service Account JSON** (recommended for cloud deployment)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures. 

### Service Account JSON 

We recommend using Service Account JSON for authentication as it makes deployment to Rill Cloud easier. The `google_application_credentials` environment variable tells Google Cloud SDK which service account key file to use for authentication.

Create your Service Account JSON with the following command:

```bash
gcloud iam service-accounts keys create ~/key.json \
  --iam-account=my-service-account@PROJECT_ID.iam.gserviceaccount.com
```

:::note Permission denied?
You'll need to contact your internal cloud admin to create your Service Account JSONs for you.
:::

Then, create a connector via the Add Data UI and it will automatically create the `gcs.yaml` file in your `connectors` directory and populate the `.env` file with `connector.gcs.google_application_credentials`.

```yaml
type: connector

driver: gcs

google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"
bucket: "gs://bucket"
```

:::tip Cloud Credentials Management
If your project has already been deployed to Rill Cloud with configured credentials, you can use `rill env pull` to [retrieve and sync these cloud credentials](/connect/credentials/#rill-env-pull) to your local `.env` file. Note that this operation will overwrite any existing local credentials for this source.
:::

### HMAC Keys

An alternative authentication method for GCP data access is using HMAC keys. This approach generates a key and secret pair (similar to AWS S3 credentials) that can be used for authentication.

Generate HMAC credentials using the following command:

```bash
gcloud storage hmac create \
  --project=PROJECT_ID \
  --service-account=SERVICE_ACCOUNT_EMAIL
```



To use these credentials, configure the `key_id` and `secret` parameters in your [GCS connector](/reference/project-files/connectors#gcs).
```yaml
type: connector

driver: gcs

key_id: "{{ .env.connector.gcs.key_id }}"
secret: "{{ .env.connector.gcs.secret }}"
bucket: "*"
```

### Local Google Cloud CLI Credentials (Local Development Only)

:::warning Not recommended for production
Local Google Cloud CLI credentials only work for local development. If you deploy to Rill Cloud using this method, your dashboards will fail. Use one of the methods above for production deployments.
:::

Follow these steps to configure your CLI credentials:

:::note Prerequisites
To use the Google Cloud CLI, you will need to [install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk). If you are unsure if this has been done, you can run the following command from the command line to see if it returns your authenticated user.
```
gcloud auth list
```
If an error or no users are returned, please follow Google's documentation on setting up your command line before continuing. Make sure to run `gcloud init` after installation as described in the tutorial.
:::

1. [Install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk).
2. Initiate the Google Cloud CLI by running `gcloud init`.
3. Set up your user by running `gcloud auth application-default login`.

:::tip Service Accounts
If you are using a service account, you will need to run the following command:
```
gcloud auth activate-service-account --key-file=path_to_json_key_file
```
:::

You have now configured Google Cloud access from your local environment. Rill will automatically detect and use these credentials when you connect to GCS sources.

## Deploy to Rill Cloud

When deploying your project to Rill Cloud, you must provide a JSON key file for a Google Cloud service account with appropriate read access/permissions to the buckets used in your project. If these credentials exist in your `.env` file, they'll be pushed with your project automatically. If you're using inferred credentials only, you'll need to configure explicit credentials to avoid deployment failures.

To manually configure your environment variables, run:
```bash
rill env configure
```


:::tip Did you know?
If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/connect/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.
:::

## Appendix

### How to create a service account using the Google Cloud Console

Here is a step-by-step guide on how to create a Google Cloud service account with read-only access to GCS:

1. Navigate to the [Service Accounts page](https://console.cloud.google.com/iam-admin/serviceaccounts) under "IAM & Admin" in the Google Cloud Console.
2. Click the "Create Service Account" button at the top of the page.
3. In the "Create Service Account" window, enter a name for the service account, then click "Create and continue".
4. In the "Role" field, search for and select the "Storage Object Viewer" role. Click "Continue", then click "Done".
    - This grants the service account access to data in all buckets. To only grant access to data in a specific bucket, leave the "Role" field blank, click "Done", then follow the steps described in [Add a principal to a bucket-level policy](https://cloud.google.com/storage/docs/access-control/using-iam-permissions#bucket-add).
5. On the "Service Accounts" page, locate the service account you just created and click on the three dots on the right-hand side. Select "Manage Keys" from the dropdown menu.
6. On the "Keys" page, click the "Add key" button and select "Create new key".
7. Choose the "JSON" key type and click "Create".
8. Download and save the JSON key file to a secure location on your computer.

### How to create a service account using the `gcloud` CLI

1. Open a terminal window and follow the steps on [Install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk) if you haven't already done so.
2. You will need your Google Cloud project ID to complete this tutorial. Run the following command to show it:
    ```bash
    gcloud config get project
    ```
3. Replace `[PROJECT_ID]` with your project ID in the following command, and run it to create a new service account (optionally also replace `rill-service-account` with a name of your choice):
    ```bash
    gcloud iam service-accounts create rill-service-account --project [PROJECT_ID]
    ```
4. Grant the service account access to data in Google Cloud Storage:
    - To grant access to data in all buckets, replace `[PROJECT_ID]` with your project ID in the following command, and run it:
        ```bash
        gcloud projects add-iam-policy-binding [PROJECT_ID] \
            --member="serviceAccount:rill-service-account@[PROJECT_ID].iam.gserviceaccount.com" \
            --role="roles/storage.objectViewer"
        ```
    - To only grant access to data in a specific bucket, replace `[BUCKET_NAME]` and `[PROJECT_ID]` with your details in the following command, and run it:
        ```bash
        gcloud storage buckets add-iam-policy-binding gs://[BUCKET_NAME] \
            --member="serviceAccount:rill-service-account@[PROJECT_ID].iam.gserviceaccount.com" \
            --role="roles/storage.objectViewer"
        ```
5. Replace `[PROJECT_ID]` with your project ID in the following command, and run it to create a key file for the service account:
    ```bash
    gcloud iam service-accounts keys create rill-service-account.json \
        --iam-account rill-service-account@[PROJECT_ID].iam.gserviceaccount.com
    ```
6. You have now created a JSON key file named `rill-service-account.json` in your current working directory.

:::info
As an alternative, to ensure that you are running Rill with a specific service account, you can provide the key in the `rill start` command. This is useful when you have multiple profiles or may receive limited access to a bucket.

`GOOGLE_APPLICATION_CREDENTIALS=<path_to_json_key_file> rill start`
:::