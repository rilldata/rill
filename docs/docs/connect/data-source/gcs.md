---
title: Google Cloud Storage (GCS)
sidebar_label: Google Cloud Storage (GCS)
sidebar_position: 30
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Overview

Connect to Google Cloud Storage (GCS) to ingest data from buckets. Rill supports multiple authentication methods including Service Account JSON keys, HMAC keys, and local Google Cloud CLI credentials for development.

## Authentication Methods

To connect to Google Cloud Storage, you need to provide authentication credentials. Rill supports three methods:

1. **Use Service Account JSON** (recommended for production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

When adding a GCS data source through the Rill UI, you'll follow a two-step process:

**Step 1: Configure Authentication**
- Choose your authentication method (Service Account JSON or HMAC Keys)
- Provide the necessary credentials
- Save the connector configuration

**Step 2: Configure Data Model**
- Specify the GCS URI for your data (e.g., `gs://my-bucket/path/to/data.parquet`)
- Configure refresh triggers and other model settings
- The model will automatically reference the connector you created in Step 1

This separation keeps your credentials secure and reusable across multiple models.

## Method 1: Service Account JSON (Recommended)

Service Account JSON authentication provides secure, programmatic access to GCS resources. This is the recommended method for production deployments.

### Prerequisites

1. A Google Cloud Platform project with Cloud Storage API enabled
2. A service account with appropriate permissions
3. A JSON key file for the service account

### Creating a Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **IAM & Admin** → **Service Accounts**
3. Click **Create Service Account**
4. Give it a name and description
5. Grant it the **Storage Object Viewer** role (or **Storage Admin** for write access)
6. Click **Done**
7. Find your new service account in the list and click on it
8. Go to the **Keys** tab
9. Click **Add Key** → **Create new key**
10. Choose **JSON** format
11. Click **Create** - the JSON key file will download automatically

### Using the UI

The Rill UI makes it easy to configure Service Account authentication:

1. Click **Add Data** in your Rill project
2. Select **Google Cloud Storage** as the connector type
3. Choose **Service Account JSON** as the authentication method
4. Upload your JSON key file or paste its contents
5. Click **Save** to create the connector

After creating the connector, you'll be prompted to configure your data model by specifying the GCS URI and other settings.

### Manual Configuration

Create two files in your Rill project:

**Step 1: Create connector file** (`connectors/my_gcs.yaml`):

```yaml
type: connector
name: my_gcs

# Reference credentials from .env file
google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"
```

**Step 2: Create model file** (`models/my_gcs_data.yaml`):

```yaml
type: model
connector: my_gcs

# GCS URI to your data
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data.parquet')

# Refresh configuration
refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

Add your credentials to `.env`:

```bash
connector.gcs.google_application_credentials='{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "service-account@project.iam.gserviceaccount.com",
  "client_id": "1234567890",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/..."
}'
```

## Method 2: HMAC Keys

HMAC keys provide S3-compatible authentication for GCS. This method is useful when you need S3-compatible access or want to avoid service account JSON files.

### Prerequisites

1. A Google Cloud Platform project with Cloud Storage API enabled
2. A service account (or user account) to generate HMAC keys for

### Generating HMAC Keys

#### Using Google Cloud Console

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **Cloud Storage** → **Settings**
3. Click the **Interoperability** tab
4. If not already enabled, click **Enable Interoperability Access**
5. Under **Service account HMAC**, select a service account from the dropdown
6. Click **Create a key for a service account**
7. The Access Key and Secret will be displayed once - save these securely

#### Using gcloud CLI

```bash
# Create HMAC key for a service account
gcloud storage hmac create SERVICE_ACCOUNT_EMAIL

# The command will output:
# Access ID: GOOG1E...
# Secret: base64-encoded-secret
```

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Google Cloud Storage** as the connector type
3. Choose **HMAC Keys** as the authentication method
4. Enter your Access Key ID and Secret Key
5. Click **Save** to create the connector

After creating the connector, you'll be prompted to configure your data model.

### Manual Configuration

Create two files in your Rill project:

**Step 1: Create connector file** (`connectors/my_gcs.yaml`):

```yaml
type: connector
name: my_gcs

# S3-compatible authentication mode using HMAC keys
aws_access_key_id: "{{ .env.connector.gcs.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.gcs.aws_secret_access_key }}"
```

**Step 2: Create model file** (`models/my_gcs_data.yaml`):

```yaml
type: model
connector: my_gcs

# GCS URI to your data (S3-compatible mode)
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data.parquet')

# Refresh configuration
refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

Add your HMAC credentials to `.env`:

```bash
connector.gcs.aws_access_key_id=GOOG1E...
connector.gcs.aws_secret_access_key=your-secret-key
```

:::note
When using HMAC keys, Rill uses S3-compatible authentication internally, which is why the credentials are named `aws_access_key_id` and `aws_secret_access_key`.
:::

## Method 3: Local Google Cloud CLI (Development Only)

For local development, you can use credentials from the Google Cloud CLI. This method is **not recommended for production** and will not work when deploying to Rill Cloud.

### Prerequisites

1. [Google Cloud CLI installed](https://cloud.google.com/sdk/docs/install)
2. Authenticated with `gcloud auth application-default login`

### Configuration

**Create model file** (`models/my_gcs_data.yaml`):

```yaml
type: model

# No connector needed - will use local gcloud credentials
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data.parquet')

# Refresh configuration (optional for local development)
refresh:
  cron: "0 */6 * * *"
```

:::warning
Local Google Cloud CLI credentials only work in local development. When you deploy your project to Rill Cloud, you must use either Service Account JSON or HMAC Keys authentication.
:::

## GCS URI Format

GCS URIs follow the format: `gs://bucket-name/path/to/file`

Examples:
- Single file: `gs://my-bucket/data.parquet`
- All files in a folder: `gs://my-bucket/folder/*`
- Pattern matching: `gs://my-bucket/data/*.parquet`

## Supported File Formats

Rill can read the following formats from GCS:

- **Parquet** (`.parquet`) - Recommended for performance
- **CSV** (`.csv`)
- **JSON** (`.json`, `.ndjson`)
- **Avro** (`.avro`)

Use DuckDB's file reading functions in your SQL:
- `read_parquet('gs://...')`
- `read_csv('gs://...')`
- `read_json('gs://...')`

## Common Use Cases

### Reading a Single Parquet File

```yaml
type: model
connector: my_gcs

sql: SELECT * FROM read_parquet('gs://my-bucket/sales/2024/data.parquet')

refresh:
  cron: "0 8 * * *"  # Daily at 8 AM
```

### Reading Multiple Files with Pattern Matching

```yaml
type: model
connector: my_gcs

sql: |
  SELECT * FROM read_parquet('gs://my-bucket/logs/2024-*.parquet')
  WHERE event_date >= CURRENT_DATE - INTERVAL 30 DAY

refresh:
  cron: "0 */2 * * *"  # Every 2 hours
```

### Reading CSV with Custom Options

```yaml
type: model
connector: my_gcs

sql: |
  SELECT * FROM read_csv(
    'gs://my-bucket/exports/data.csv',
    header=true,
    delimiter=',',
    quote='"'
  )

refresh:
  cron: "0 6 * * *"  # Daily at 6 AM
```

### Combining Multiple GCS Sources

```yaml
type: model
connector: my_gcs

sql: |
  SELECT 'sales' as source, * FROM read_parquet('gs://my-bucket/sales/*.parquet')
  UNION ALL
  SELECT 'returns' as source, * FROM read_parquet('gs://my-bucket/returns/*.parquet')

refresh:
  cron: "0 4 * * *"  # Daily at 4 AM
```

## IAM Permissions

Your service account needs the following IAM permissions:

**For read-only access:**
- `storage.objects.get`
- `storage.objects.list`

**Predefined role:** `Storage Object Viewer` (`roles/storage.objectViewer`)

**For read-write access:**
- All of the above, plus:
- `storage.objects.create`
- `storage.objects.delete`

**Predefined role:** `Storage Admin` (`roles/storage.admin`)

## Deploying to Rill Cloud

When deploying your project to Rill Cloud:

1. **Service Account JSON Method:**
   - The JSON key file will be securely stored in Rill Cloud
   - No additional configuration needed

2. **HMAC Keys Method:**
   - Your Access Key ID and Secret will be securely stored in Rill Cloud
   - No additional configuration needed

3. **Local Google Cloud CLI:**
   - ❌ Will NOT work in Rill Cloud
   - You must reconfigure to use Service Account JSON or HMAC Keys before deploying

To deploy with authentication:

```bash
# Deploy with GCS credentials
rill deploy
```

The Rill Cloud deployment will use the connector configuration from your project, and you'll be prompted to provide credentials if they're not already set.

## Troubleshooting

### "Access Denied" Errors

**Problem:** Getting 403 or access denied errors when reading from GCS.

**Solutions:**
1. Verify your service account has the correct IAM permissions
2. Check that the bucket name and path are correct
3. Ensure the service account has access to the specific bucket
4. For HMAC keys, verify the keys haven't been revoked

### "Connector Not Found" Errors

**Problem:** Model can't find the referenced connector.

**Solutions:**
1. Verify the connector name in your model matches the connector file name
2. Check that the connector file is in the `connectors/` directory
3. Ensure the connector YAML is valid and has `type: connector`

### Local Development Works But Cloud Deployment Fails

**Problem:** Data loads locally but fails when deployed to Rill Cloud.

**Solutions:**
1. Check if you're using local Google Cloud CLI credentials (not supported in cloud)
2. Switch to Service Account JSON or HMAC Keys authentication
3. Verify credentials are properly stored in your `.env` file and excluded from git

### "Invalid JSON Key" Errors

**Problem:** Service Account JSON is rejected.

**Solutions:**
1. Ensure the entire JSON key is properly formatted
2. Check for extra spaces or newlines
3. Verify the JSON is valid using a JSON validator
4. Make sure you're using the complete key file downloaded from GCP

### HMAC Key Authentication Fails

**Problem:** HMAC keys don't work with GCS.

**Solutions:**
1. Verify Interoperability Access is enabled for your project
2. Check that the HMAC key hasn't been deactivated or deleted
3. Ensure you're using both the Access ID and Secret correctly
4. Confirm the service account associated with the HMAC key has proper permissions

## Related Documentation

- [Connectors Overview](/connect/connectors)
- [Models Overview](/build/models)
- [DuckDB GCS Extension](https://duckdb.org/docs/extensions/httpfs.html)
- [Google Cloud Storage IAM](https://cloud.google.com/storage/docs/access-control/iam)
