---
title: Azure Blob Storage
description: Connect to data in Azure Blob Storage
sidebar_label: Azure Blob Storage 
sidebar_position: 05
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview
[Azure Blob Storage (ABS)](https://learn.microsoft.com/en-us/azure/storage/blobs/storage-blobs-introduction) is a scalable, fully managed, and highly reliable object storage solution offered by Microsoft Azure, designed to store and access data from anywhere in the world. It provides a secure and cost-effective way to store data, including common storage formats such as CSV and Parquet. Rill supports connecting to and reading from Azure Blob Storage using the following Resource URI syntax:

```bash
azure://<account>.blob.core.windows.net/<container>/path/file.csv
```

## Authentication Methods

To connect to Azure Blob Storage, you can choose from four authentication options:

1. **Storage Account Key** (recommended for cloud deployment)
2. **Connection String** (alternative for cloud deployment)
3. **Shared Access Signature (SAS) Token** (most secure, fine-grained control)
4. **Public** (for publicly accessible containers - no authentication required)
5. **Azure CLI authentication** (local development only - not recommended for production)

:::tip Authentication Methods
Choose the method that best fits your setup. For production deployments to Rill Cloud, use Storage Account Key, Connection String, or SAS tokens. Public containers don't require authentication and skip connector creation. Azure CLI authentication only works for local development and will cause deployment failures.
:::

## Using the Add Data UI

When you add an Azure Blob Storage data model through the Rill UI, you'll see four authentication options:

- **Storage Account Key**, **Connection String**, or **SAS Token**: The process follows two steps:
  1. **Configure Authentication** - Set up your Azure connector with credentials
  2. **Configure Data Model** - Define which container and objects to ingest
  The UI will automatically create both the connector file and model file for you.

- **Public**: For publicly accessible containers, you skip the connector creation step and go directly to:
  1. **Configure Data Model** - Define which container and objects to ingest
  The UI will only create the model file (no connector file is needed).

:::note Manual Configuration Only
Azure CLI authentication is only available through manual configuration. See [Method 5: Azure CLI Authentication](#method-5-azure-cli-authentication-local-development-only) for setup instructions.
:::

---

## Method 1: Storage Account Key (Recommended)

Storage Account Key credentials provide reliable authentication for Azure Blob Storage. This method works for both local development and Rill Cloud deployments.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Azure Blob Storage** as the data model type
3. In the authentication step:
   - Choose **Storage Account Key**
   - Enter your Storage Account name
   - Enter your Storage Account Key
   - Name your connector (e.g., `my_azure`)
4. In the data model configuration step:
   - Enter your container name and object path
   - Configure other model settings as needed
5. Click **Create** to finalize

The UI will automatically create both the connector file and model file for you.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/my_azure.yaml`:

```yaml
type: connector
driver: azure

azure_storage_account: rilltest
azure_storage_key: "{{ .env.connector.azure.azure_storage_key }}"
```

**Step 2: Create model configuration**

Create `models/my_azure_data.yaml`:

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: my_azure

sql: SELECT * FROM read_parquet('azure://rilltest.blob.core.windows.net/my-container/path/to/data/*.parquet')

refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.azure.azure_storage_key=your_storage_account_key
```

Follow the [Azure Documentation](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-keys-manage?tabs=azure-portal) to retrieve your storage account keys.

---

## Method 2: Connection String

Connection String provides an alternative authentication method for Azure Blob Storage.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Azure Blob Storage** as the data model type
3. In the authentication step:
   - Choose **Connection String**
   - Enter your Connection String
   - Name your connector (e.g., `my_azure_conn`)
4. In the data model configuration step:
   - Enter your container name and object path
   - Configure other model settings as needed
5. Click **Create** to finalize

### Manual Configuration

**Step 1: Create connector configuration**

Create `connectors/my_azure_conn.yaml`:

```yaml
type: connector
driver: azure

azure_storage_connection_string: "{{ .env.connector.azure.azure_storage_connection_string }}"
```

**Step 2: Create model configuration**

Create `models/my_azure_data.yaml`:

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: my_azure_conn

sql: SELECT * FROM read_parquet('azure://rilltest.blob.core.windows.net/my-container/path/to/data/*.parquet')

refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.azure.azure_storage_connection_string=your_connection_string
```

Follow the [Azure Documentation](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-keys-manage?tabs=azure-portal) to retrieve your connection string.

---

## Method 3: Shared Access Signature (SAS) Token

SAS tokens provide fine-grained access control with specific permissions and expiration times for secure access to your storage resources.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Azure Blob Storage** as the data model type
3. In the authentication step:
   - Choose **SAS Token**
   - Enter your Storage Account name
   - Enter your SAS Token
   - Name your connector (e.g., `my_azure_sas`)
4. In the data model configuration step:
   - Enter your container name and object path
   - Configure other model settings as needed
5. Click **Create** to finalize

### Manual Configuration

**Step 1: Create connector configuration**

Create `connectors/my_azure_sas.yaml`:

```yaml
type: connector
driver: azure

azure_storage_account: rilltest 
azure_storage_sas_token: "{{ .env.connector.azure.azure_storage_sas_token }}"
```

**Step 2: Create model configuration**

Create `models/my_azure_data.yaml`:

```yaml
type: model
connector: duckdb
create_secrets_from_connectors: my_azure_sas

sql: SELECT * FROM read_parquet('azure://rilltest.blob.core.windows.net/my-container/path/to/data/*.parquet')

refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.azure.azure_storage_sas_token=your_sas_token
```

Follow the [Azure Documentation](https://learn.microsoft.com/en-us/azure/ai-services/translator/document-translation/how-to-guides/create-sas-tokens?tabs=Containers) to create your Azure SAS token.

---

## Method 4: Public Containers

For publicly accessible Azure Blob Storage containers, you don't need to create a connector. Simply use the Azure URI directly in your model configuration.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Azure Blob Storage** as the data model type
3. In the authentication step:
   - Choose **Public**
   - The UI will skip connector creation and proceed directly to data model configuration
4. In the data model configuration step:
   - Enter your container name and object path
   - Configure other model settings as needed
5. Click **Create** to finalize

The UI will only create the model file (no connector file is created).

### Manual Configuration

For public containers, you only need to create a model file. No connector configuration is required.

Create `models/my_azure_data.yaml`:

```yaml
type: model
connector: duckdb

sql: SELECT * FROM read_parquet('azure://publicaccount.blob.core.windows.net/my-public-container/path/to/data/*.parquet')

refresh:
  cron: "0 */6 * * *"
```

---

## Method 5: Azure CLI Authentication (Local Development Only)

For local development, you can use credentials from the Azure CLI. This method is **not suitable for production** or Rill Cloud deployments. This method is only available through manual configuration, and you don't need to create a connector file.

### Setup

1. Install the [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) if not already installed
2. Authenticate with your Azure account:
   ```bash
   az login
   ```
3. Verify your authentication status:
   ```bash
   az account show
   ```
4. Create your model file (no connector needed)

### Model Configuration

Create `models/my_azure_data.yaml`:

```yaml
type: model
connector: duckdb

sql: SELECT * FROM read_parquet('azure://rilltest.blob.core.windows.net/my-container/path/to/data/*.parquet')

refresh:
  cron: "0 */6 * * *"
```

Rill will automatically detect and use your local Azure CLI credentials when no connector is specified.

:::warning
This method only works for local development. Deploying to Rill Cloud with this configuration will fail because the cloud environment doesn't have access to your local credentials. Always use Storage Account Key, Connection String, or SAS tokens for production deployments.
:::

## Using Azure Blob Storage Data in Models

Once your connector is configured (or for public containers, no connector needed), you can reference Azure Blob Storage paths in your model SQL queries using DuckDB's Azure functions.

### Basic Example

**With a connector (authenticated):**

```yaml
type: model
connector: duckdb

sql: SELECT * FROM read_parquet('azure://rilltest.blob.core.windows.net/my-container/data/*.parquet')

refresh:
  cron: "0 */6 * * *"
```

**Public container (no connector needed):**

```yaml
type: model
connector: duckdb

sql: SELECT * FROM read_parquet('azure://publicaccount.blob.core.windows.net/my-public-container/data/*.parquet')

refresh:
  cron: "0 */6 * * *"
```

### Path Patterns

You can use wildcards to read multiple files:

```sql
-- Single file
SELECT * FROM read_parquet('azure://account.blob.core.windows.net/container/data/file.parquet')

-- All files in a directory
SELECT * FROM read_parquet('azure://account.blob.core.windows.net/container/data/*.parquet')

-- All files in nested directories
SELECT * FROM read_parquet('azure://account.blob.core.windows.net/container/data/**/*.parquet')

-- Files matching a pattern
SELECT * FROM read_parquet('azure://account.blob.core.windows.net/container/data/2024-*.parquet')
```

---

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide either an Azure Blob Storage connection string, Azure Storage Key, or Azure Storage SAS token for the containers used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#azure) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:
```
rill env push
```

