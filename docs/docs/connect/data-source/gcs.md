---
title: "Google Cloud Storage"
sidebar_label: "Google Cloud Storage"
sidebar_position: 30
---

Google Cloud Storage (GCS) is Google's object storage service. Rill supports ingesting data from GCS buckets.

:::tip
For more information on how to model data from your GCS buckets, see our [OLAP data modeling documentation](/build/models/models.md).
:::

## Authentication

To connect to Google Cloud Storage, you need to provide authentication credentials. Rill supports three methods:

1. **Use Service Account JSON** (recommended for both development and production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

When you add a GCS model through Rill's UI, you'll follow a two-step authentication flow:

1. **Configure Authentication** - Set up your credentials (Service Account JSON or HMAC keys)
2. **Configure Data Model** - Specify the GCS path and configure your data model

This explicit flow ensures credentials are properly configured before you define your data model.

## Option 1: Service Account JSON (Recommended)

Using a Service Account JSON key file is the recommended method for authenticating with GCS. This works for both local development and Rill Cloud deployments.

### Prerequisites

1. A Google Cloud project with the Cloud Storage API enabled
2. A Service Account with appropriate permissions to read from your GCS buckets
3. A JSON key file for the Service Account

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Google Cloud Storage** as the connector type
3. Choose **Service Account JSON** as the authentication method
4. Upload your Service Account JSON key file using the file picker
5. Click **Save** to store the connector configuration
6. In Step 2, configure your data model:
   - Enter the GCS path (e.g., `gs://my-bucket/path/to/data.parquet`)
   - Configure any additional model settings
7. Click **Save** to create your model

The credentials will be stored in `connectors/gcs.yaml` and your model configuration will be in `models/my_model.yaml`.

### Manual Configuration

If you prefer to configure files manually, create two separate files:

#### Step 1: Create the connector file

Create `connectors/gcs.yaml`:
