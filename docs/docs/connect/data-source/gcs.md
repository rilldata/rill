---
title: Google Cloud Storage (GCS)
sidebar_label: Google Cloud Storage (GCS)
sidebar_position: 30
---

<!-- markdownlint-disable MD034 -->

## Overview

Rill supports ingesting data from Google Cloud Storage buckets through a connector-based authentication flow. You can authenticate using Service Account JSON credentials, HMAC keys, or local Google Cloud CLI credentials (for development only).

## Authentication Methods

To connect to Google Cloud Storage, you need to provide authentication credentials using one of these methods:

1. **Use Service Account JSON** (recommended for production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

When adding a GCS data model through Rill's UI, you'll follow a two-step process:

**Step 1: Configure Authentication** - You'll first set up a GCS connector with your credentials (Service Account JSON or HMAC Keys).

**Step 2: Configure Data Model** - After authentication is configured, you'll create a model that references your connector and specifies which GCS files to ingest.

This separation allows you to reuse the same connector across multiple models.

## Method 1: Service Account JSON (Recommended)

Service Account JSON authentication uses a Google Cloud service account key file. This is the recommended method for production deployments.

### Using the UI

1. Navigate to the **Add Data** interface in Rill
2. Select **Google Cloud Storage** as your data source
3. In the authentication step:
   - Choose **Service Account JSON**
   - Upload your service account JSON key file using the file picker
   - Or paste the JSON content directly into the text field
4. Configure your data model:
   - Specify the GCS path (e.g., `gs://my-bucket/path/to/data/*.parquet`)
   - Set a refresh schedule if needed
   - Complete the model configuration

### Manual Configuration

If you prefer to configure files manually, you'll need to create two files:

#### Step 1: Create the Connector File

Create a file at `connectors/my_gcs.yaml`:

```yaml
type: connector
driver: gcs
google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"
```

Add your Service Account JSON credentials to your `.env` file:

```bash
connector.gcs.google_application_credentials=<service-account-json>
```

:::tip
The Service Account JSON should be provided as a single-line string. You can convert a multi-line JSON file to a single line using: `cat service-account.json | jq -c`
:::

#### Step 2: Create the Model File

Create a file at `models/my_gcs_data.yaml`:

```yaml
type: model
connector: my_gcs
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')
refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

Or if you prefer SQL files, create `models/my_gcs_data.sql`:

```sql
-- @connector: my_gcs
-- @refresh.cron: 0 */6 * * *

SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')
```

### Setting up a Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **IAM & Admin** → **Service Accounts**
3. Click **Create Service Account**
4. Provide a name and description, then click **Create and Continue**
5. Grant the service account appropriate permissions (at minimum, `Storage Object Viewer` for read access)
6. Click **Done**
7. Find your new service account in the list and click on it
8. Go to the **Keys** tab
9. Click **Add Key** → **Create new key**
10. Choose **JSON** format and click **Create**
11. The JSON key file will download automatically - keep it secure!

## Method 2: HMAC Keys

HMAC (Hash-based Message Authentication Code) keys provide an alternative authentication method that's compatible with S3-style authentication.

### Using the UI

1. Navigate to the **Add Data** interface in Rill
2. Select **Google Cloud Storage** as your data source
3. In the authentication step:
   - Choose **HMAC Keys**
   - Enter your Access Key ID
   - Enter your Secret Access Key
4. Configure your data model:
   - Specify the GCS path (e.g., `gs://my-bucket/path/to/data/*.parquet`)
   - Set a refresh schedule if needed
   - Complete the model configuration

### Manual Configuration

#### Step 1: Create the Connector File

Create a file at `connectors/my_gcs.yaml`:

```yaml
type: connector
driver: gcs
access_key_id: "{{ .env.connector.gcs.access_key_id }}"
secret_access_key: "{{ .env.connector.gcs.secret_access_key }}"
```

Add your HMAC credentials to your `.env` file:

```bash
connector.gcs.access_key_id=<your-access-key-id>
connector.gcs.secret_access_key=<your-secret-access-key>
```

#### Step 2: Create the Model File

Create a file at `models/my_gcs_data.yaml`:

```yaml
type: model
connector: my_gcs
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')
refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

Or using SQL format at `models/my_gcs_data.sql`:

```sql
-- @connector: my_gcs
-- @refresh.cron: 0 */6 * * *

SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')
```

### Generating HMAC Keys

You can generate HMAC keys using either the Google Cloud Console or the `gsutil` command-line tool.

#### Option A: Using Google Cloud Console

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **Cloud Storage** → **Settings** → **Interoperability**
3. If you haven't already, set a default project for interoperability
4. Under **User account HMAC**, click **Create a key for a user account** (or **Create a key for a service account** if using a service account)
5. Select the appropriate user or service account
6. Click **Create key**
7. Copy the **Access Key** and **Secret** - you won't be able to retrieve the secret again!

#### Option B: Using gsutil CLI

```bash
gsutil hmac create <service-account-email>
```

This will output:
```
Access ID: <access-key-id>
Secret: <secret-access-key>
```

:::warning
Save your HMAC secret immediately - you cannot retrieve it again after creation. If you lose it, you'll need to create a new HMAC key.
:::

## Method 3: Local Google Cloud CLI (Development Only)

For local development, you can use credentials from the Google Cloud CLI. This method is simpler for development but **will not work** when deploying to Rill Cloud.

### Setup

1. Install the [Google Cloud CLI](https://cloud.google.com/sdk/docs/install)
2. Run `gcloud auth application-default login` to authenticate
3. Create your model file without a connector reference

### Model Configuration

Create a file at `models/my_gcs_data.yaml`:

```yaml
type: model
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')
refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

:::danger Production Deployment
Local Google Cloud CLI credentials are stored only on your local machine and will cause deployment failures when pushing to Rill Cloud. For production deployments, use Service Account JSON or HMAC Keys.
:::

## File Path Formats

GCS paths in your SQL queries should use the `gs://` protocol:

```sql
-- Single file
SELECT * FROM read_parquet('gs://my-bucket/data/file.parquet')

-- Multiple files with wildcard
SELECT * FROM read_parquet('gs://my-bucket/data/*.parquet')

-- Nested directory with glob pattern
SELECT * FROM read_parquet('gs://my-bucket/data/**/*.parquet')
```

## Deploying to Rill Cloud

When deploying a project with GCS models to Rill Cloud:

1. **Ensure you're using Service Account JSON or HMAC Keys** - Local CLI credentials will not work in Rill Cloud
2. **Set up credentials as environment variables** in your Rill Cloud project:
   - For Service Account JSON: Set `connector.gcs.google_application_credentials`
   - For HMAC Keys: Set `connector.gcs.access_key_id` and `connector.gcs.secret_access_key`
3. **Deploy your project** - Rill will use the configured connector to access your GCS data

### Setting Environment Variables in Rill Cloud

Via the Rill CLI:
```bash
rill env set connector.gcs.google_application_credentials='<service-account-json>'
```

Or for HMAC keys:
```bash
rill env set connector.gcs.access_key_id='<access-key-id>'
rill env set connector.gcs.secret_access_key='<secret-access-key>'
```

## Common Issues

### Authentication Failures

**Problem**: `Failed to authenticate with GCS`

**Solutions**:
- Verify your Service Account JSON is valid and properly formatted
- Ensure your service account has the necessary permissions (`Storage Object Viewer` or higher)
- For HMAC keys, verify both access key ID and secret are correct
- Check that your credentials are properly set in your `.env` file or Rill Cloud environment variables

### File Not Found Errors

**Problem**: `File or directory not found: gs://my-bucket/...`

**Solutions**:
- Verify the bucket name and path are correct
- Ensure your service account or HMAC key has access to the specific bucket
- Check that files exist at the specified path
- Verify your wildcard patterns match existing files

### Deployment Failures

**Problem**: Model works locally but fails in Rill Cloud

**Solutions**:
- Ensure you're using Service Account JSON or HMAC Keys (not local CLI credentials)
- Verify environment variables are set correctly in Rill Cloud
- Check that your service account has access from external networks (not restricted by IP)
- Confirm your connector file is included in your project

## Supported File Formats

GCS models in Rill support DuckDB's file reading capabilities, including:

- **Parquet**: `read_parquet('gs://bucket/path/*.parquet')`
- **CSV**: `read_csv('gs://bucket/path/*.csv')`
- **JSON**: `read_json('gs://bucket/path/*.json')`
- **NDJSON**: `read_json_auto('gs://bucket/path/*.ndjson')`

## Additional Resources

- [DuckDB GCS Extension Documentation](https://duckdb.org/docs/extensions/httpfs.html#gcs)
- [Google Cloud Storage Documentation](https://cloud.google.com/storage/docs)
- [Service Account Key Management Best Practices](https://cloud.google.com/iam/docs/best-practices-for-managing-service-account-keys)
- [Rill Model Documentation](../../build/models/models.md)
