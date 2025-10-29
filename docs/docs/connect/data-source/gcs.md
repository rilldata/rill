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

To connect to Google Cloud Storage, you need to provide authentication credentials (or skip for public buckets). Rill supports three methods:

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
2. Select **Google Cloud Storage (GCS)** as the data model type
3. In the authentication step:
   - Choose **Service Account JSON**
   - Upload your JSON key file or paste its contents
   - Name your connector (e.g., `my_gcs`)
4. In the data model configuration step:
   - Enter your bucket name and object path
   - Configure other model settings as needed
5. Click **Create** to finalize

The UI will automatically create both the connector file and model file for you.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/my_gcs.yaml`:

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
connector.gcs.google_application_credentials='{"type":"service_account","project_id":"your-project",...}'
```

:::tip
See the [Appendix](#how-to-create-a-service-account-using-the-google-cloud-console) for detailed instructions on creating a service account.
:::

---

## Method 2: HMAC Keys

HMAC keys provide S3-compatible authentication for GCS. This method is useful when you need compatibility with S3-style access patterns.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Google Cloud Storage (GCS)** as the data model type
3. In the authentication step:
   - Choose **HMAC Keys**
   - Enter your Access Key ID
   - Enter your Secret Access Key
   - Name your connector (e.g., `my_gcs_hmac`)
4. In the data model configuration step:
   - Enter your bucket name and object path
   - Configure other model settings as needed
5. Click **Create** to finalize

### Manual Configuration

**Step 1: Create connector configuration**

Create `connectors/my_gcs_hmac.yaml`:

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
See the [Appendix](#how-to-generate-hmac-keys-using-the-google-cloud-console) for detailed instructions on generating HMAC keys.
:::

---

## Method 3: Local Google Cloud CLI Credentials

For local development, you can use credentials from the Google Cloud CLI. This method is **not suitable for production** or Rill Cloud deployments.

### Setup

1. Install the [Google Cloud CLI](https://cloud.google.com/sdk/docs/install)
2. Authenticate with your Google account:
   ```bash
   gcloud auth application-default login
   ```
3. Create your model file referencing a connector without explicit credentials

### Configuration

Create `connectors/my_gcs.yaml`:

```yaml
type: connector
driver: gcs
```

Create `models/my_gcs_data.yaml`:

```yaml
type: model
connector: duckdb
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

When no explicit credentials are provided in the connector, Rill will automatically use your local Google Cloud CLI credentials.

:::warning
This method only works for local development. Deploying to Rill Cloud with this configuration will fail because the cloud environment doesn't have access to your local credentials. Always use Service Account JSON or HMAC keys for production deployments.
:::

---

## Deploying to Rill Cloud

When deploying your project to Rill Cloud, you must provide a JSON key file for a Google Cloud service account with appropriate read access/permissions to the buckets used in your project. If these credentials exist in your `.env` file, they'll be pushed with your project automatically. If you're using inferred credentials only, you'll need to configure explicit credentials to avoid deployment failures.

To configure your environment variables for deployment, run:

```bash
rill env configure
```

The CLI will interactively prompt you to configure credentials for all connectors in your project.

Then deploy with:

```bash
rill deploy
```

:::tip Did you know?
If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/connect/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.
:::

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

### How to create a service account using the gcloud CLI

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

### How to generate HMAC keys using the Google Cloud Console

You can generate HMAC keys through the Google Cloud Console or using the `gcloud` CLI:

**Using the Google Cloud Console:**

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **Cloud Storage** > **Settings** > **Interoperability**
3. If not already enabled, click **Enable Interoperability Access**
4. Scroll to **Service account HMAC** section
5. Select a service account or create a new one
6. Click **Create a key for a service account**
7. Copy the **Access Key** and **Secret** - you won't be able to view the secret again

**Using the gcloud CLI:**

```bash
# Create HMAC keys for a service account
gcloud storage hmac create SERVICE_ACCOUNT_EMAIL

# List existing HMAC keys
gcloud storage hmac list
```

:::info
As an alternative, to ensure that you are running Rill with a specific service account, you can provide the key in the `rill start` command. This is useful when you have multiple profiles or may receive limited access to a bucket.

```bash
rill start --env key_id=<KEY> --env secret=<SECRET>
```
:::
