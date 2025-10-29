---
title: Google Cloud Storage
sidebar_label: Google Cloud Storage
sidebar_position: 30
---

Google Cloud Storage (GCS) is Google's object storage service. Rill supports connecting to GCS buckets to read data files (CSV, Parquet, JSON, etc.) for building dashboards.

:::tip

For information about Google BigQuery, see [BigQuery](/connect/bigquery).

:::

## Authentication Methods

To connect to Google Cloud Storage, you need to provide authentication credentials. Rill supports three methods:

1. **Use Service Account JSON** (recommended for production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

The easiest way to connect to GCS is through Rill's Add Data interface:

1. **Step 1: Configure Authentication** - Add your GCS credentials (Service Account JSON or HMAC keys)
2. **Step 2: Configure Data Model** - Specify the bucket and file path(s) to read

This two-step process separates authentication (stored in `connectors/`) from data configuration (stored in `models/`), making it easier to reuse credentials across multiple data models.

## Method 1: Service Account JSON (Recommended)

Service Account JSON is the recommended authentication method for production deployments. This method uses a JSON key file from a Google Cloud service account.

### Prerequisites

1. A Google Cloud project with billing enabled
2. A service account with appropriate permissions
3. A JSON key file for the service account

### Setting Up a Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **IAM & Admin** → **Service Accounts**
3. Click **Create Service Account**
4. Give it a name and description
5. Grant the **Storage Object Viewer** role (or **Storage Admin** for write access)
6. Click **Done**
7. Click on the created service account
8. Go to the **Keys** tab
9. Click **Add Key** → **Create new key**
10. Choose **JSON** format
11. Click **Create** - the JSON key file will download automatically

### Using the UI

1. In your Rill project, click **Add Data** in the top right
2. Select **Google Cloud Storage** from the list of data sources
3. Choose **Service Account JSON** as the authentication method
4. Upload your JSON key file or paste its contents
5. Click **Save** to create the connector
6. Configure your data model by specifying the bucket and file path
7. Click **Add Data** to create the model

This will create two files:
- `connectors/gcs.yaml` - Contains your authentication credentials
- `models/my_gcs_data.yaml` - References the connector and specifies data location

### Manual Configuration

Alternatively, you can create the connector and model files manually:

**Step 1: Create a connector file** (`connectors/gcs.yaml`):

```yaml
type: connector
driver: gcs

# Reference the service account JSON from your .env file
google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"
```

Add the service account JSON to your `.env` file:

```bash
connector.gcs.google_application_credentials={"type":"service_account","project_id":"your-project",...}
```

**Step 2: Create a model file** (`models/my_gcs_data.yaml`):

```yaml
type: model

# Reference the connector
connector: gcs

# Specify the GCS path
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')

# Add a refresh trigger for production
refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

## Method 2: HMAC Keys

HMAC keys provide S3-compatible authentication for GCS. This method is useful when you need S3-compatible access to GCS buckets.

### Generating HMAC Keys

#### Using Google Cloud Console

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **Cloud Storage** → **Settings**
3. Click on the **Interoperability** tab
4. If you haven't created keys before, click **Create keys for a service account**
5. Select an existing service account or create a new one
6. Click **Create new key**
7. Save the **Access Key** and **Secret** - you won't be able to see the secret again

#### Using gcloud CLI

```bash
# Create HMAC key for a service account
gcloud storage hmac create SERVICE_ACCOUNT_EMAIL

# List existing HMAC keys
gcloud storage hmac list
```

### Using the UI

1. In your Rill project, click **Add Data** in the top right
2. Select **Google Cloud Storage** from the list of data sources
3. Choose **HMAC Keys** as the authentication method
4. Enter your **Access Key ID** and **Secret Access Key**
5. Click **Save** to create the connector
6. Configure your data model by specifying the bucket and file path
7. Click **Add Data** to create the model

### Manual Configuration

**Step 1: Create a connector file** (`connectors/gcs.yaml`):

```yaml
type: connector
driver: gcs

# Use S3-compatible authentication with HMAC keys
aws_access_key_id: "{{ .env.connector.gcs.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.gcs.aws_secret_access_key }}"
```

Add the HMAC credentials to your `.env` file:

```bash
connector.gcs.aws_access_key_id=GOOG1234567890ABCDEF
connector.gcs.aws_secret_access_key=your-secret-key-here
```

**Step 2: Create a model file** (`models/my_gcs_data.yaml`):

```yaml
type: model

# Reference the connector
connector: gcs

# Specify the GCS path (use gs:// prefix)
sql: SELECT * FROM read_csv('gs://my-bucket/path/to/data/*.csv')

# Add a refresh trigger for production
refresh:
  cron: "0 */6 * * *"  # Refresh every 6 hours
```

## Method 3: Local Google Cloud CLI (Development Only)

:::warning
This method only works for local development. Do not use it for production deployments to Rill Cloud, as it will cause authentication failures.
:::

If you have the Google Cloud SDK installed and configured locally, Rill can use your local credentials automatically.

### Prerequisites

1. Install the [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
2. Authenticate with your Google account:
   ```bash
   gcloud auth application-default login
   ```

### Configuration

**Create a model file** (`models/my_gcs_data.yaml`):

```yaml
type: model

# No connector needed - uses local gcloud credentials
# Leave connector empty or omit it
connector: ""

# Specify the GCS path
sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data/*.parquet')

# Add a refresh trigger if needed
refresh:
  cron: "0 */6 * * *"
```

When running `rill start` locally, Rill will automatically use your Google Cloud CLI credentials.

## Common File Formats

GCS supports various file formats through DuckDB's read functions:

### Parquet Files

```yaml
type: model
connector: gcs
sql: SELECT * FROM read_parquet('gs://my-bucket/data/*.parquet')
refresh:
  cron: "0 */6 * * *"
```

### CSV Files

```yaml
type: model
connector: gcs
sql: SELECT * FROM read_csv('gs://my-bucket/data/*.csv', auto_detect=true)
refresh:
  cron: "0 */6 * * *"
```

### JSON Files

```yaml
type: model
connector: gcs
sql: SELECT * FROM read_json('gs://my-bucket/data/*.json')
refresh:
  cron: "0 */6 * * *"
```

### Multiple Files with Wildcards

```yaml
type: model
connector: gcs
sql: |
  SELECT * FROM read_parquet([
    'gs://my-bucket/data/2024-01-*.parquet',
    'gs://my-bucket/data/2024-02-*.parquet'
  ])
refresh:
  cron: "0 */6 * * *"
```

## Deploying to Rill Cloud

When deploying to Rill Cloud, you must use either Service Account JSON or HMAC Keys authentication. Local Google Cloud CLI credentials will not work in production.

### Required Files

Your project must include:
1. **Connector file** (`connectors/gcs.yaml`) - Contains authentication configuration
2. **Model file(s)** (`models/*.yaml`) - References the connector and specifies data location
3. **Environment variables** (`.env`) - Contains sensitive credentials (not committed to git)

### Deployment Steps

1. Ensure your connector file is properly configured with Service Account JSON or HMAC keys
2. Add credentials to your `.env` file (this file should be in `.gitignore`)
3. Deploy to Rill Cloud:
   ```bash
   rill deploy
   ```
4. During deployment, you'll be prompted to securely provide your environment variables
5. Rill Cloud will use these credentials to authenticate with GCS

### Setting Environment Variables in Rill Cloud

After deploying, you can manage environment variables in the Rill Cloud UI:

1. Go to your project in Rill Cloud
2. Navigate to **Settings** → **Environment Variables**
3. Add or update your GCS credentials:
   - For Service Account JSON: `connector.gcs.google_application_credentials`
   - For HMAC Keys: `connector.gcs.aws_access_key_id` and `connector.gcs.aws_secret_access_key`

## Refresh Triggers

For production use, configure refresh triggers to keep your data up to date:

```yaml
type: model
connector: gcs
sql: SELECT * FROM read_parquet('gs://my-bucket/data/*.parquet')

# Refresh every 6 hours
refresh:
  cron: "0 */6 * * *"
```

Common refresh patterns:
- Every hour: `"0 * * * *"`
- Every 6 hours: `"0 */6 * * *"`
- Daily at midnight: `"0 0 * * *"`
- Every Monday at 9am: `"0 9 * * 1"`

## Troubleshooting

### Authentication Errors

**Error**: `Failed to authenticate with GCS`

**Solutions**:
- Verify your service account JSON is valid and complete
- Check that the service account has the necessary permissions (Storage Object Viewer)
- Ensure your HMAC keys are correct and active
- Confirm environment variables are properly set in `.env`

### File Not Found Errors

**Error**: `No files found matching pattern`

**Solutions**:
- Verify the bucket name and path are correct
- Check that the service account has access to the bucket
- Ensure files exist at the specified path
- Try listing files with `gsutil ls gs://my-bucket/path/` to verify access

### Permission Denied

**Error**: `Permission denied when accessing gs://my-bucket/...`

**Solutions**:
- Grant the service account the **Storage Object Viewer** role on the bucket
- Check bucket-level and object-level permissions
- Verify the service account is active and not disabled

### Deployment Failures

**Error**: `Authentication failed after deploying to Rill Cloud`

**Solutions**:
- Ensure you're using Service Account JSON or HMAC keys (not local gcloud credentials)
- Verify environment variables are set in Rill Cloud
- Check that the connector file properly references environment variables
- Confirm the service account JSON is correctly formatted

## Best Practices

1. **Use Service Account JSON for production** - It's the most secure and reliable method
2. **Store credentials in `.env`** - Never commit credentials to git
3. **Use least-privilege access** - Grant only the permissions needed (typically Storage Object Viewer)
4. **Separate connectors from models** - Keep authentication in `connectors/` and data configuration in `models/`
5. **Add refresh triggers** - Configure appropriate refresh schedules for production data
6. **Use wildcards efficiently** - Leverage glob patterns to read multiple files efficiently
7. **Test locally first** - Verify your configuration works with `rill start` before deploying
8. **Monitor costs** - GCS charges for data egress, so be mindful of how much data you're reading

## Additional Resources

- [Google Cloud Storage Documentation](https://cloud.google.com/storage/docs)
- [Service Account Key Management](https://cloud.google.com/iam/docs/creating-managing-service-account-keys)
- [HMAC Keys for GCS](https://cloud.google.com/storage/docs/authentication/hmackeys)
- [DuckDB GCS Extension](https://duckdb.org/docs/extensions/httpfs.html)
