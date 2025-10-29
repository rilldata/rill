---
title: Google Cloud Storage (GCS)
sidebar_label: Google Cloud Storage (GCS)
sidebar_position: 30
---

Google Cloud Storage (GCS) is Google's object storage service. Rill can connect to GCS buckets to ingest data files.

## Authentication Methods

To connect to Google Cloud Storage, you need to provide authentication credentials. There are three ways to authenticate:

1. **Use Service Account JSON** (recommended for production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

Rill's UI provides a streamlined two-step process for connecting to GCS:

1. **Step 1: Configure Authentication** - Set up your GCS connector with credentials
2. **Step 2: Configure Data Model** - Select specific files/paths to ingest

This workflow creates two files in your project:
- A **connector file** (`connectors/gcs.yaml`) that stores authentication configuration
- A **model file** (`models/my-model.yaml`) that defines which data to ingest

:::tip
The UI automatically links these files together, so you don't need to manually reference the connector in your model files.
:::

## Method 1: Service Account JSON (Recommended)

Service Account authentication is the recommended approach for production deployments. It provides secure, scoped access to your GCS resources.

### Prerequisites

1. A Google Cloud project with billing enabled
2. A GCS bucket with data files
3. A service account with appropriate permissions

### Using the UI

1. Navigate to your Rill project
2. Click **Add Data** → **Google Cloud Storage**
3. In Step 1 (Configure Authentication):
   - Enter a name for your connector (e.g., "gcs")
   - Select **Service Account JSON** as the authentication method
   - Upload your service account JSON key file
   - Click **Save Connector**
4. In Step 2 (Configure Data Model):
   - Enter the GCS path to your data (e.g., `gs://my-bucket/data/*.parquet`)
   - Configure any additional options
   - Click **Add Data**

The UI will create both the connector and model files automatically.

### Manual Configuration

If you prefer to configure files manually, create two separate files:

#### Connector File (`connectors/gcs.yaml`)

```yaml
type: gcs
google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"
```

#### Model File (`models/my-gcs-data.yaml`)

```yaml
type: model
connector: gcs
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')
```

### Setting up a Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **IAM & Admin** → **Service Accounts**
3. Click **Create Service Account**
4. Give it a name and description
5. Grant it the **Storage Object Viewer** role (or more permissive roles as needed)
6. Click **Create Key** → **JSON** to download the credentials file
7. Store the JSON content securely in your `.env` file

### Environment Variables

Add your service account JSON to `.env`:

```bash
connector.gcs.google_application_credentials='{
  "type": "service_account",
  "project_id": "your-project",
  "private_key_id": "...",
  "private_key": "...",
  "client_email": "...",
  "client_id": "...",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "..."
}'
```

:::warning
Keep your service account JSON secure and never commit it to version control.
:::

## Method 2: HMAC Keys

HMAC keys provide S3-compatible authentication for GCS. This method is useful when you need S3-compatible access patterns or when integrating with tools that expect S3 credentials.

### Prerequisites

1. A Google Cloud project with billing enabled
2. A GCS bucket with data files
3. HMAC keys generated for your service account

### Using the UI

1. Navigate to your Rill project
2. Click **Add Data** → **Google Cloud Storage**
3. In Step 1 (Configure Authentication):
   - Enter a name for your connector (e.g., "gcs")
   - Select **HMAC Keys** as the authentication method
   - Enter your HMAC Access Key ID
   - Enter your HMAC Secret Access Key
   - Click **Save Connector**
4. In Step 2 (Configure Data Model):
   - Enter the GCS path to your data (e.g., `gs://my-bucket/data/*.parquet`)
   - Configure any additional options
   - Click **Add Data**

### Manual Configuration

If you prefer to configure files manually, create two separate files:

#### Connector File (`connectors/gcs.yaml`)

```yaml
type: gcs
aws_access_key_id: "{{ .env.connector.gcs.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.gcs.aws_secret_access_key }}"
```

#### Model File (`models/my-gcs-data.yaml`)

```yaml
type: model
connector: gcs
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')
```

:::info
Notice that the connector uses `aws_access_key_id` and `aws_secret_access_key`. This is intentional - HMAC keys use S3-compatible authentication.
:::

### Generating HMAC Keys

You can create HMAC keys through the Google Cloud Console or CLI:

#### Using Google Cloud Console

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **Cloud Storage** → **Settings** → **Interoperability**
3. If not already enabled, click **Enable Interoperability Access**
4. Scroll to **Service account HMAC** section
5. Select a service account or create a new one
6. Click **Create a key for a service account**
7. Copy the **Access Key** and **Secret** - you won't be able to view the secret again

#### Using Google Cloud CLI

```bash
# Create HMAC keys for a service account
gcloud storage hmac create SERVICE_ACCOUNT_EMAIL

# List existing HMAC keys
gcloud storage hmac list
```

### Environment Variables

Add your HMAC keys to `.env`:

```bash
connector.gcs.aws_access_key_id=GOOG1234567890ABCDEFG
connector.gcs.aws_secret_access_key=your-secret-access-key
```

:::warning
Keep your HMAC keys secure and never commit them to version control.
:::

## Method 3: Local Google Cloud CLI (Development Only)

For local development, you can use credentials from the Google Cloud CLI (`gcloud`). This method uses your personal Google Cloud credentials.

:::caution
This method only works for local development. It will cause deployment failures on Rill Cloud. For production deployments, use Service Account JSON or HMAC Keys.
:::

### Prerequisites

1. [Google Cloud CLI (`gcloud`)](https://cloud.google.com/sdk/gcloud) installed
2. Authenticated with a Google account that has access to your GCS bucket

### Setup

1. Install and authenticate with `gcloud`:

```bash
gcloud auth application-default login
```

2. Create a connector file without explicit credentials:

#### Connector File (`connectors/gcs.yaml`)

```yaml
type: gcs
```

#### Model File (`models/my-gcs-data.yaml`)

```yaml
type: model
connector: gcs
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')
```

Rill will automatically use your local `gcloud` credentials when no explicit credentials are provided.

:::warning
Remember to switch to Service Account JSON or HMAC Keys before deploying to Rill Cloud.
:::

## Using GCS Data in Models

Once your connector is configured, you can reference GCS paths in your model SQL queries using DuckDB's GCS functions.

### Basic Example

```yaml
type: model
connector: gcs
sql: SELECT * FROM read_parquet('gs://my-bucket/data/*.parquet')
```

### Reading Multiple File Types

```yaml
type: model
connector: gcs
sql: |
  -- Read Parquet files
  SELECT * FROM read_parquet('gs://my-bucket/parquet-data/*.parquet')
  
  UNION ALL
  
  -- Read CSV files
  SELECT * FROM read_csv('gs://my-bucket/csv-data/*.csv', AUTO_DETECT=TRUE)
```

### Path Patterns

You can use wildcards to read multiple files:

```sql
-- Single file
SELECT * FROM read_parquet('gs://my-bucket/data/file.parquet')

-- All files in a directory
SELECT * FROM read_parquet('gs://my-bucket/data/*.parquet')

-- All files in nested directories
SELECT * FROM read_parquet('gs://my-bucket/data/**/*.parquet')

-- Files matching a pattern
SELECT * FROM read_parquet('gs://my-bucket/data/2024-*.parquet')
```

## Deploying to Rill Cloud

When deploying your Rill project to Rill Cloud, you need to ensure your GCS credentials are properly configured.

### Prerequisites

- A Rill Cloud account
- Service Account JSON or HMAC Keys configured (local `gcloud` credentials will not work)

### Deployment Steps

1. Ensure your connector file references environment variables:

```yaml
# For Service Account JSON
type: gcs
google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"

# OR for HMAC Keys
type: gcs
aws_access_key_id: "{{ .env.connector.gcs.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.gcs.aws_secret_access_key }}"
```

2. Set the credentials in Rill Cloud:

```bash
# For Service Account JSON
rill env set connector.gcs.google_application_credentials='{"type":"service_account",...}'

# OR for HMAC Keys
rill env set connector.gcs.aws_access_key_id=GOOG1234567890ABCDEFG
rill env set connector.gcs.aws_secret_access_key=your-secret-access-key
```

3. Deploy your project:

```bash
rill deploy
```

:::tip
Use `rill env set --local` to set credentials for local development without affecting your cloud deployment.
:::

## Supported File Formats

GCS connectors support any file format that DuckDB can read:

| Format  | DuckDB Function  | Example                                                                  |
| ------- | ---------------- | ------------------------------------------------------------------------ |
| Parquet | `read_parquet()` | `read_parquet('gs://bucket/data.parquet')`                               |
| CSV     | `read_csv()`     | `read_csv('gs://bucket/data.csv')`                                       |
| JSON    | `read_json()`    | `read_json('gs://bucket/data.json')`                                     |
| Excel   | `read_excel()`   | Requires [spatial extension](https://duckdb.org/docs/extensions/spatial) |

See the [DuckDB documentation](https://duckdb.org/docs/) for more details on file readers.

## Troubleshooting

### Authentication Errors

**Problem**: `Failed to authenticate with GCS`

**Solutions**:
- Verify your service account JSON is valid and complete
- Check that your HMAC keys are correct and not expired
- Ensure your service account has the necessary permissions
- For local development, run `gcloud auth application-default login`

### Permission Errors

**Problem**: `Access Denied` or `403 Forbidden`

**Solutions**:
- Grant your service account the **Storage Object Viewer** role at minimum
- Check bucket-level IAM permissions
- Verify the bucket exists and is in the correct project

### Deployment Failures

**Problem**: Model works locally but fails in Rill Cloud

**Solutions**:
- Verify you're using Service Account JSON or HMAC Keys (not local `gcloud` credentials)
- Check that credentials are set in Rill Cloud: `rill env list`
- Ensure environment variable references match between local and cloud

### Path Issues

**Problem**: `File not found` errors

**Solutions**:
- Verify the GCS path format: `gs://bucket-name/path/to/file`
- Check for typos in bucket or file names
- Ensure wildcards are properly escaped in YAML strings
- Test paths using the `gcloud storage ls` command

## Additional Resources

- [Google Cloud Storage Documentation](https://cloud.google.com/storage/docs)
- [Service Account Best Practices](https://cloud.google.com/iam/docs/best-practices-service-accounts)
- [DuckDB GCS Extension](https://duckdb.org/docs/extensions/gcs)
- [Rill Models Documentation](/build/models)
