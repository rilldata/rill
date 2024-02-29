---
title: Azure Blob Storage
description: Connect to data in Azure Blob Storage
sidebar_label: ABS
sidebar_position: 3
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview
[Azure Blob Storage (ABS)](https://learn.microsoft.com/en-us/azure/storage/blobs/storage-blobs-introduction) is a scalable, fully managed, and highly reliable object storage solution offered by Microsoft Azure, designed to store and access data from anywhere in the world. It provides a secure and cost-effective way to store data, including common storage formats for data such as CSV and parquet. Rill supports natively connecting to Azure Blob Storage using the appropriate [Resource](https://learn.microsoft.com/en-us/rest/api/storageservices/naming-and-referencing-containers--blobs--and-metadata#resource-uri-syntax) [URI](https://learn.microsoft.com/en-us/azure/storage/blobs/data-lake-storage-introduction-abfs-uri#uri-syntax) of your container / blob to retrieve and read files.

![Connecting to ABS](/img/reference/connectors/azure/abs.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), Rill uses by default the credentials configured in your local environment using the Azure CLI (`az`). Follow these steps to configure it:

1. Open a terminal window and run the following command to log in to your Azure account:

    ```bash
    az login
    ```

    Follow the on-screen instructions to complete the login process. This will authenticate you with your Azure account.

    > If you don't have the Azure CLI installed, you can [install it from here](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli).

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

## Cloud deployment

When deploying a project to Rill Cloud (i.e. `rill deploy`), Rill requires either an Azure Blob Storage connection string, Azure Storage Key, or Azure Storage SAS token to be explicitly provided for the Azure Blob Storage containers used in your project. 

When you first deploy a project using `rill deploy`, you will be prompted to provide credentials for the remote sources in your project that require authentication.

If you subsequently add sources that require new credentials (or if you input the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:

```bash
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::
