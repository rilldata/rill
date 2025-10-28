---
title: Google Cloud Storage
sidebar_label: Google Cloud Storage
sidebar_position: 20
---

Google Cloud Storage (GCS) is a cloud-based object storage service provided by Google Cloud Platform. This connector allows you to read data files stored in GCS buckets.

## Authentication Methods

To connect to Google Cloud Storage, you need to provide authentication credentials. Rill supports three authentication methods:

1. **Use Service Account JSON** (recommended for production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

When you add a GCS data model through the UI, Rill uses a two-step process:

1. **Step 1: Configure Authentication** - Set up your GCS connector with authentication credentials
2. **Step 2: Configure Data Model** - Reference your GCS files and create your data model

This separation allows you to:
- Reuse the same connector across multiple data models
- Manage credentials independently from data model definitions
- Keep sensitive credentials separate from your model logic

## Method 1: Service Account JSON (Recommended)

### Using the UI

1. Navigate to your Rill project
2. Click "Add Data" and select "Google Cloud Storage"
3. In Step 1 (Configure Authentication):
   - Enter a name for your connector (e.g., `my_gcs`)
   - Upload your Service Account JSON key file using the file picker
   - Click "Save Connector"
4. In Step 2 (Configure Data Model):
   - Enter the GCS URI for your data (e.g., `gs://my-bucket/path/to/data.parquet`)
   - Configure any additional model settings
   - Click "Create Model"

### Manual Configuration

#### Create Connector File

Create a connector file at `connectors/my_gcs.yaml`:

```yaml
type: gcs
google_application_credentials: path/to/service-account-key.json
