---
title: Google Cloud Storage (GCS)
description: Connect to data in GCS
sidebar_label: GCS
sidebar_position: 15
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview
[Google Cloud Storage (GCS)](https://cloud.google.com/storage/docs/introduction) is a scalable, fully managed, and highly reliable object storage service offered by Google Cloud, designed to store and access data from anywhere in the world. It provides a secure and cost-effective way to store data, including common data storage formats such as CSV and Parquet. Rill supports natively connecting to GCS using the provided [Google Cloud Storage URI](https://cloud.google.com/bigquery/docs/cloud-storage-transfer-overview#google-cloud-storage-uri) of your bucket to retrieve and read files.

<img src='/img/connect/data-sources/gcs.png' class='rounded-gif' style={{width: '75%', display: 'block', margin: '0 auto'}}/>
<br />

## Rill Developer (Local credentials)

When using Rill Developer on your local machine (i.e., `rill start`), Rill will use either the credentials configured in your local environment using the Google Cloud CLI (`gcloud`) or an [explicitly defined connector YAML](/reference/project-files/connectors#gcs). 

### Inferred Credentials 

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
You have now configured Google Cloud access from your local environment. Rill will detect and use your credentials the next time you try to ingest a source.

### Service Account JSON 

`GOOGLE_APPLICATION_CREDENTIALS` is an environment variable that tells Google Cloud SDK and applications which service account key file to use for authentication. It should point to the full path of your JSON key file. We recommend creating using this credential for Rill, as this makes deployment to Rill Cloud easier. For more information on JSON keys, see the [Google Cloud documentation](https://cloud.google.com/iam/docs/keys-create-delete?hl=en#gcloud).

Assuming you've followed steps 1 and 2 above, you'll need to create your Service Account JSON with the following command.

```bash
gcloud iam service-accounts keys create ~/key.json \
  --iam-account=my-service-account@PROJECT_ID.iam.gserviceaccount.com
```

:::note Permission denied?
You'll need to contact your internal cloud admin to create your Service Account JSONs for you.
:::

To configure Rill to use these credentials, create a `.env` file in your project directory (if one doesn't already exist) and add your service account `google_application_credentials` as a single-line string:

```bash
google_application_credentials='{"type": "service_account", "project_id": "your-project", ...}'
```

Once configured, Rill will automatically use these credentials for all Google Cloud Platform connections, including [BigQuery](/connect/data-source/bigquery).


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

:::warning Security Best Practice

Never commit sensitive credentials directly to your connector YAML files or version control. Instead, use environment variables to reference these values securely.

```yaml
key_id: '{{.env.connector.gcs.key_id}}'
secret: '{{.env.connector.gcs.secret}}'
```

Configure these values in your `.env` file:

```env
connector.gcs.key_id=GOOG1E...
connector.gcs.secret=wRu6iE...
```

:::

## Rill Cloud Deployment

When deploying a project to Rill Cloud, Rill requires a JSON key file to be explicitly provided for a Google Cloud service account with appropriate read access/permissions to the buckets used in your project. If this already exists in your `.env` file, this will be pushed with your project automatically. If you are using inferred credentials, your deployment will result in errored dashboards.

If you want to manually configure your environment variables, run the following command:
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