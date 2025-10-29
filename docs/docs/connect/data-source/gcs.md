---
title: Google Cloud Storage (GCS)
sidebar_label: Google Cloud Storage (GCS)
sidebar_position: 30
---

## Overview

Google Cloud Storage (GCS) lets you ingest data from buckets into Rill. You can connect using Service Account JSON credentials, HMAC keys, or local Google Cloud CLI credentials (local development only).

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
connector: gcs
google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"
```

**Step 2: Create model configuration**

Create `models/my_gcs_data.yaml`:

```yaml
type: model
connector: my_gcs
path: gs://my-bucket/path/to/data.parquet

# Optional: Specify SQL transformation
# sql: SELECT * FROM read_parquet('gs://my-bucket/path/to/data.parquet')
```

**Step 3: Add credentials to `.env`**

```bash
connector.gcs.google_application_credentials=<json_credentials>
```

### Creating a Service Account

To obtain Service Account JSON credentials:

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **IAM & Admin** > **Service Accounts**
3. Click **Create Service Account**
4. Provide a name and description, then click **Create and Continue**
5. Grant the service account the **Storage Object Viewer** role (or a more restrictive custom role)
6. Click **Done**
7. Click on the newly created service account
8. Go to the **Keys** tab
9. Click **Add Key** > **Create New Key**
10. Select **JSON** format and click **Create**
11. The JSON key file will download automatically

:::warning
Store your Service Account JSON key securely. It provides access to your GCS resources. Never commit it to version control.
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
connector: gcs
access_key_id: "{{ .env.connector.gcs.access_key_id }}"
secret_access_key: "{{ .env.connector.gcs.secret_access_key }}"
```

**Step 2: Create model configuration**

Create `models/my_gcs_data.yaml`:

```yaml
type: model
connector: my_gcs_hmac
path: gs://my-bucket/path/to/data.parquet
```

**Step 3: Add credentials to `.env`**

```bash
connector.gcs.access_key_id=<your_access_key_id>
connector.gcs.secret_access_key=<your_secret_access_key>
```

### Generating HMAC Keys

You can generate HMAC keys through the Google Cloud Console or using the `gsutil` command-line tool.

#### Using Google Cloud Console

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to **Cloud Storage** > **Settings** > **Interoperability**
3. If you don't have a default project, set one
4. Scroll to **Access keys for your user account**
5. Click **Create a key**
6. Your Access Key and Secret will be displayed - save them securely

#### Using gsutil CLI

```bash
# Generate HMAC keys for your default service account
gsutil hmac create

# Generate HMAC keys for a specific service account
gsutil hmac create <service-account-email>
```

:::info
HMAC keys use S3-compatible authentication. When using HMAC keys, GCS transparently handles the authentication in an S3-compatible mode.
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
3. Create your model file without a connector reference

### Model Configuration

Create `models/my_gcs_data.yaml`:

```yaml
type: model
path: gs://my-bucket/path/to/data.parquet
```

When no `connector` is specified and you're running locally, Rill will automatically use your local Google Cloud CLI credentials.

:::warning
This method only works for local development. Deploying to Rill Cloud with this configuration will fail because the cloud environment doesn't have access to your local credentials. Always use Service Account JSON or HMAC keys for production deployments.
:::

---

## Deploying to Rill Cloud

When deploying your project to Rill Cloud, you must use either Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials will not work in the cloud environment.

### Setting Credentials for Deployment

Use the `rill env configure` command to securely set your credentials:

**For Service Account JSON:**
```bash
rill env configure --project <project-name> connector.gcs.google_application_credentials
```

**For HMAC Keys:**
```bash
rill env configure --project <project-name> connector.gcs.access_key_id
rill env configure --project <project-name> connector.gcs.secret_access_key
```

You'll be prompted to enter the credential values securely. These are stored encrypted in Rill Cloud and injected at runtime.

### Deployment Checklist

Before deploying, ensure:

- [ ] You're using Service Account JSON or HMAC Keys (not local CLI credentials)
- [ ] Your connector file references environment variables (e.g., `{{ .env.connector.gcs.google_application_credentials }}`)
- [ ] You've set the credentials using `rill env configure`
- [ ] Your Service Account has the necessary GCS permissions
- [ ] Your model files reference the connector by name

---

## Supported File Formats

GCS models support the same file formats as other Rill data models:

- **Parquet** (`.parquet`)
- **CSV** (`.csv`)
- **JSON** (`.json`, `.ndjson`)
- **Excel** (`.xlsx`, `.xls`)
- **Avro** (`.avro`)

You can reference files using glob patterns:

```yaml
path: gs://my-bucket/data/*.parquet
```

---

## Common Issues

### "Failed to authenticate" errors

- Verify your credentials are correct
- Check that your Service Account has the **Storage Object Viewer** role
- Ensure credentials are properly set in `.env` (local) or via `rill env configure` (cloud)

### "Access denied" errors

- Confirm your Service Account or HMAC key has read permissions for the bucket
- Check bucket-level IAM policies
- Verify the bucket name and path are correct

### Deployment failures with local credentials

- Local Google Cloud CLI credentials don't work in Rill Cloud
- Switch to Service Account JSON or HMAC Keys for deployment
- Use `rill env configure` to set credentials for your cloud project

---

## Additional Resources

- [Google Cloud Storage Documentation](https://cloud.google.com/storage/docs)
- [Service Account Best Practices](https://cloud.google.com/iam/docs/best-practices-service-accounts)
- [HMAC Keys Documentation](https://cloud.google.com/storage/docs/authentication/hmackeys)
- [Rill Cloud Deployment Guide](/deploy/deploy-with-cli)
