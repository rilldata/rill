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

## Connect to Azure Blob Storage

To connect to Azure Blob Storage, you need to provide authentication credentials. You have four options:

1. **Use Storage Account Key** (recommended for cloud deployment)
2. **Use Connection String** (alternative for cloud deployment)
3. **Use Shared Access Signature (SAS) Token** (most secure, fine-grained control)
4. **Use Azure CLI authentication** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Storage Account Key, Connection String, or SAS tokens. Azure CLI authentication only works for local development and will cause deployment failures.

### Storage Account Key

To ensure seamless deployment to Rill Cloud, configure your Azure Storage Account Key directly in your project's `.env` file instead of relying solely on Azure CLI authentication (which only works locally).

```yaml
type: connector

driver: azure

azure_storage_account: rilltest
azure_storage_key: "{{ .env.connector.azure.azure_storage_key }}"
```

This approach ensures your Azure Blob Storage sources authenticate consistently across both local development and cloud deployment. Follow the [Azure Documentation](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-keys-manage?tabs=azure-portal) to retrieve your storage account keys.

### Connection String

To ensure seamless deployment to Rill Cloud, configure your Azure Blob Storage credentials using a connection string directly in your project's `.env` file instead of relying solely on Azure CLI authentication (which only works locally).

```yaml
type: connector

driver: azure

azure_storage_connection_string: "{{ .env.connector.azure.azure_storage_connection_string }}"
```

This approach ensures your Azure Blob Storage sources authenticate consistently across both local development and cloud deployment. Follow the [Azure Documentation](https://learn.microsoft.com/en-us/azure/storage/common/storage-account-keys-manage?tabs=azure-portal) to retrieve your connection string.

### Shared Access Signature (SAS) Token

Use Shared Access Signature (SAS) tokens as an alternative authentication method for Azure Blob Storage. SAS tokens provide fine-grained access control with specific permissions and expiration times for secure access to your storage resources.

```yaml
type: connector

driver: azure

azure_storage_account: rilltest 
azure_storage_sas_token: "{{ .env.connector.azure.azure_storage_sas_token }}"
```

This method provides fine-grained access control and enhanced security for your Azure Blob Storage connections. Follow the [Azure Documentation](https://learn.microsoft.com/en-us/azure/ai-services/translator/document-translation/how-to-guides/create-sas-tokens?tabs=Containers) to create your Azure SAS token.

###  Azure CLI Authentication (Local Development Only)

:::warning Not recommended for production
Azure CLI authentication only works for local development. If you deploy to Rill Cloud using this method, your dashboards will fail. Use one of the methods above for production deployments.
:::

1. Install the [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) if not already installed.
2. Open a terminal window and run the following command to log in to your Azure account: `az login`
3. Verify your authentication status: `az account show`

You've now configured Azure access from your local environment. Rill will automatically detect and use these credentials when you connect to Azure Blob Storage sources.

:::tip Cloud Credentials Management

If your project is already deployed to Rill Cloud with configured credentials, use `rill env pull` to [retrieve and sync these cloud credentials](/build/connectors/credentials/#rill-env-pull) to your local `.env` file. **Warning**: This operation will overwrite any existing local credentials for this source.

:::

## Deploy to Rill Cloud

When deploying your project to Rill Cloud, you must provide either an Azure Blob Storage connection string, Azure Storage Key, or Azure Storage SAS token for the containers used in your project. If these credentials exist in your `.env` file, they'll be pushed with your project automatically. If you're using inferred credentials only, your deployment will result in errored dashboards.

To manually configure your environment variables, run:
```bash
rill env configure
```

:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), use `rill env push` to [push these credentials](/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This allows other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::
