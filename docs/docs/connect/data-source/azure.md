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

<img src='/img/reference/connectors/azure/abs.png' class='centered' />
<br />

## Local credentials

When using Rill Developer on your local machine (i.e., `rill start`), Rill will either use the credentials configured in your local environment using the Azure CLI (`az`) or the explicitly defined [credentials in a connector YAML](/reference/project-files/connectors#azure).

Assuming you have the Azure CLI installed, follow the steps below to configure it:

### Inferred Credentials

1. Open a terminal window and run the following command to log in to your Azure account:

    ```bash
    az login
    ```

    Follow the on-screen instructions to complete the login process. This will authenticate you with your Azure account.

    :::info

    To check if you already have the Azure CLI installed and are authenticated, you can open a terminal window and run the following command:

    ```bash
    az account show
    ```

    If it does not display any information about your Azure account, you can [install the Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) if it is not already installed on your system.

    :::

2. Once you are logged in, Rill will automatically use the credentials obtained from `az login` to authenticate with Azure Blob Storage when you interact with Azure Blob Storage sources.


### Using Connection String

Alternatively, you can use an Azure Blob Storage connection string to configure the credentials. To do this:

1. Obtain the connection string for your Azure Blob Storage account. You can find this in the Azure Portal under "Access keys" in your storage account settings.

2. Set the `AZURE_STORAGE_CONNECTION_STRING` environment variable in your local environment to the connection string value. You can do this in your terminal:

    ```bash
    export AZURE_STORAGE_CONNECTION_STRING="your_connection_string_here"
    ```

    Replace "your_connection_string_here" with your actual connection string.

3. Rill will automatically use the connection string from the `AZURE_STORAGE_CONNECTION_STRING` environment variable to authenticate with Azure Blob Storage when you interact with Azure Blob Storage sources.

### Using Shared Access Signature (SAS) Token

As another alternative, you can configure credentials using a Shared Access Signature (SAS) token. To do this:

1. Generate a SAS token for the Azure Blob Storage container or blob you want to access. You can create SAS tokens using the Azure Portal or programmatically using the Azure SDKs.

    > [Learn how to create SAS tokens using this guide](https://learn.microsoft.com/en-us/azure/ai-services/translator/document-translation/how-to-guides/create-sas-tokens?tabs=Containers).

2. Set the `AZURE_STORAGE_SAS_TOKEN` environment variable in your local environment to the SAS token value. You can do this in your terminal:

    ```bash
    export AZURE_STORAGE_SAS_TOKEN="your_sas_token_here"
    ```

    Replace "your_sas_token_here" with your actual SAS token.

3. Rill will use the SAS token from the `AZURE_STORAGE_SAS_TOKEN` environment variable to authenticate with Azure Blob Storage when interacting with Azure Blob Storage sources.

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](//connect/credentials/#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.

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
