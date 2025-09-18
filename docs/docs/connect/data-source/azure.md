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

<img src='/img/connect/data-sources/abs.png' class='rounded-gif' style={{width: '75%', display: 'block', margin: '0 auto'}}/>
<br />

## Rill Developer (Local credentials)

When using Rill Developer on your local machine, Rill will use credentials configured in your local environment using the Azure CLI (`az`) or explicitly defined [credentials in a connector YAML](/reference/project-files/connectors#azure).

### Inferred Credentials

1. Install the [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) if not already installed.
2. Open a terminal window and run the following command to log in to your Azure account: `az login`
3. Verify your authentication status: `az account show`

You have now configured Azure access from your local environment. Rill will detect and use your credentials the next time you try to ingest a source.

### Using Storage Account Keys

For seamless deployment to Rill Cloud, you can configure Azure Storage Account credentials directly in your project's `.env` file instead of relying solely on Azure CLI authentication, which only configures credentials for local usage.

Create or update your `.env` file with the following Azure Storage Account credentials:

```env
azure_storage_account=your_storage_account_name
azure_storage_key=oFUw8vZplXd...
```

This approach ensures that your Azure Blob Storage sources can authenticate consistently across both local development and cloud deployment environments.

### Using Connection String

For seamless deployment to Rill Cloud, you can configure Azure Blob Storage credentials using a connection string directly in your project's `.env` file instead of relying solely on Azure CLI authentication, which only configures credentials for local usage.

Create or update your `.env` file with the following Azure Storage connection string:

```env
azure_storage_connection_string='DefaultEndpointsProtocol=https;AccountName=your_account;AccountKey=your_key;EndpointSuffix=core.windows.net'
```

This approach ensures that your Azure Blob Storage sources can authenticate consistently across both local development and cloud deployment environments.

### Using Shared Access Signature (SAS) Token

An alternative authentication method for Azure Blob Storage is using Shared Access Signature (SAS) tokens. This approach generates a token with specific permissions and expiration time for secure access to your storage resources.

Generate SAS tokens using the Azure Portal or programmatically with the Azure SDKs:

[Learn how to create SAS tokens using this guide](https://learn.microsoft.com/en-us/azure/ai-services/translator/document-translation/how-to-guides/create-sas-tokens?tabs=Containers).

Configure the SAS token in your `.env` file:

```env
azure_storage_sas_token='se=2025-09-18T23%3A59%3A...'
```

This method provides fine-grained access control and enhanced security for your Azure Blob Storage connections.

:::tip Cloud Credentials Management

If your project has already been deployed to Rill Cloud with configured credentials, you can use `rill env pull` to [retrieve and sync these cloud credentials](/connect/credentials/#rill-env-pull) to your local `.env` file. Note that this operation will overwrite any existing local credentials for this source.

:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/connect/templating).

## Cloud deployment

When deploying a project to Rill Cloud, Rill requires either an Azure Blob Storage connection string, Azure Storage Key, or Azure Storage SAS token to be explicitly provided for the Azure Blob Storage containers used in your project. If this already exists in your `.env` file, this will be pushed with your project automatically. If you are using inferred credentials, your deployment will result in errored dashboards.

If you want to manually configure your environment variables, run the following command:
```bash
rill env configure
```

:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/connect/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::
