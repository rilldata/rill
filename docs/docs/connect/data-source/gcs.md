---
title: Google Cloud Storage (GCS)
description: Connect to data in GCS
sidebar_label: GCS
sidebar_position: 15
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Google Cloud Storage (GCS)](https://cloud.google.com/storage/docs/introduction) is a scalable, fully managed, and highly reliable object storage service offered by Google Cloud, designed to store and access data from anywhere in the world. It provides a secure and cost-effective way to store data, including common data storage formats such as CSV and Parquet. You can connect to GCS using the provided [Google Cloud Storage URI](https://cloud.google.com/bigquery/docs/cloud-storage-transfer-overview#google-cloud-storage-uri) of your bucket to retrieve and read files.

## Authentication Methods

To connect to Google Cloud Storage, you need to provide authentication credentials. Rill supports three methods:

1. **Use Service Account JSON** (recommended for production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

When you add a GCS data model through the Rill UI, the process follows two steps:

1. **Configure Authentication** - Set up your GCS connector with credentials (Service Account JSON or HMAC keys)
2. **Configure Data Model** - Define which bucket and objects to ingest

This two-step flow ensures your credentials are securely stored in the connector configuration, while your data model references remain clean and portable.

---

## Method 1: Service Account JSON (Recommended)

Service Account JSON credentials provide the most secure and reliable authentication for GCS. This method works for both local development and Rill Cloud deployments.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Google Cloud Storage (GCS)** as the data source type
3. In the authentication step:
   - Choose **Service Account JSON**
   - Upload your JSON key file or paste its contents
   - Name your connector (e.g., `gcs`)
4. In the data model configuration step:
   - Enter your bucket name and object path
   - Configure other model settings as needed
5. Click **Create** to finalize

The UI will automatically create both the connector file and model file for you.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/gcs.yaml`:

```yaml
type: connector
driver: gcs
google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"
```

**Step 2: Create model configuration**

Create `models/my_gcs_data.yaml`:

```yaml
type: model
connector: duckdb
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.gcs.google_application_credentials=<json_credentials>
```

:::tip
For detailed instructions on creating a Service Account, see the [Appendix](#how-to-create-a-service-account-using-the-google-cloud-console).
:::

---

## Method 2: HMAC Keys

HMAC keys provide S3-compatible authentication for GCS. This method is useful when you need compatibility with S3-style access patterns.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Google Cloud Storage (GCS)** as the data source type
3. In the authentication step:
   - Choose **HMAC Keys**
   - Enter your Access Key ID
   - Enter your Secret Access Key
   - Name your connector (e.g., `gcs`)
4. In the data model configuration step:
   - Enter your bucket name and object path
   - Configure other model settings as needed
5. Click **Create** to finalize

### Manual Configuration

**Step 1: Create connector configuration**

Create `connectors/gcs.yaml`:

```yaml
type: connector
driver: gcs
key_id: "{{ .env.connector.gcs.key_id }}"
secret: "{{ .env.connector.gcs.secret }}"
```

**Step 2: Create model configuration**

Create `models/my_gcs_data.yaml`:

```yaml
type: model
connector: duckdb
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.gcs.key_id=GOOG1234567890ABCDEFG
connector.gcs.secret=your-secret-access-key
```

:::tip
For detailed instructions on generating HMAC keys, see the [Appendix](#generating-hmac-keys).
:::

:::info
HMAC keys use S3-compatible authentication. When using HMAC keys, GCS transparently handles the authentication in an S3-compatible mode.
:::

---

## Method 3: Local Google Cloud CLI Credentials

For local development, you can use credentials from the Google Cloud CLI. This method is **not suitable for production** or Rill Cloud deployments.

:::warning Not recommended for production
Local Google Cloud CLI credentials only work for local development. If you deploy to Rill Cloud using this method, your dashboards will fail. Use one of the methods above for production deployments.
:::

### Prerequisites

To use the Google Cloud CLI, you will need to [install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk). If you are unsure if this has been done, you can run the following command from the command line to see if it returns your authenticated user:

```bash
gcloud auth list
```

If an error or no users are returned, please follow Google's documentation on setting up your command line before continuing. Make sure to run `gcloud init` after installation as described in the tutorial.

### Setup

1. [Install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk)
2. Initiate the Google Cloud CLI by running `gcloud init`
3. Set up your user by running `gcloud auth application-default login`

:::tip Service Accounts
If you are using a service account, you will need to run the following command:
```bash
gcloud auth activate-service-account --key-file=path_to_json_key_file
```
:::

### Model Configuration

Create `models/my_gcs_data.yaml`:

```yaml
type: model
connector: duckdb
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

When no explicit credentials are configured, Rill will automatically use your local Google Cloud CLI credentials.

---

## Deploy to Rill Cloud

When deploying your project to Rill Cloud, you must provide a JSON key file for a Google Cloud service account with appropriate read access/permissions to the buckets used in your project. If these credentials exist in your `.env` file, they'll be pushed with your project automatically. If you're using inferred credentials only, you'll need to configure explicit credentials to avoid deployment failures.

To manually configure your environment variables, run:

```bash
rill env configure
```

This command will walk you through configuring all required connector credentials interactively.

---

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

:::note Permission denied?
You'll need to contact your internal cloud admin to create your Service Account JSONs for you.
:::

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

### Generating HMAC Keys

You can generate HMAC keys through the Google Cloud Console or using the `gcloud` CLI.

#### Using Google Cloud Console

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **Cloud Storage** > **Settings** > **Interoperability**
3. Click **Create a key for a service account**
4. Select the appropriate service account
5. Copy the **Access Key** and **Secret**

#### Using the gcloud CLI

```bash
gcloud storage hmac create \
  --project=PROJECT_ID \
  --service-account=SERVICE_ACCOUNT_EMAIL
```

Replace `PROJECT_ID` and `SERVICE_ACCOUNT_EMAIL` with your specific values.
