---
title: "Google Cloud Storage"
sidebar_label: "Google Cloud Storage"
sidebar_position: 30
---

Google Cloud Storage (GCS) is a scalable object storage service offered by Google Cloud Platform. Rill supports connecting to GCS buckets to access data stored in various formats.

## Authentication Methods

To connect to Google Cloud Storage, you need to provide authentication credentials. Rill supports three authentication methods:

1. **Use Service Account JSON** (recommended for production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

When adding a GCS data model through the UI, you'll follow a two-step process:

1. **Step 1: Configure Authentication** - Set up your GCS connector with credentials
2. **Step 2: Configure Data Model** - Select your data file and configure the model

This separation allows you to reuse the same connector for multiple data models.

## Method 1: Service Account JSON (Recommended)

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Google Cloud Storage** as your data source
3. Choose **Service Account JSON** as the authentication method
4. Upload your service account JSON key file or paste the JSON content directly
5. Give your connector a name (e.g., `gcs`)
6. Click **Save** to create the connector
7. In Step 2, provide the GCS URI to your data file (e.g., `gs://my-bucket/data.parquet`)
8. Configure your model settings and click **Generate**

### Manual Configuration

When manually configuring authentication, you'll create two separate files:

#### 1. Create Connector File

Create a connector file (e.g., `connectors/gcs.yaml`):

```yaml
type: gcs
service_account_json: <paste your service account JSON here>
