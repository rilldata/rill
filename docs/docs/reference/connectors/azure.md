---
title: Azure Blob Storage
description: Connect to data in an Azure Blob Storage container.
sidebar_label: Azure
sidebar_position: 30
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## How to configure credentials in Rill

How you configure access to Azure Blob Storage depends on whether you are developing a project locally using `rill start` or are setting up a deployment using `rill deploy`.

### Configure credentials for local development

When developing a project locally, Rill uses the credentials configured in your local environment using the Azure CLI (`az`). Follow these steps to configure it:

1. Open a terminal window and run the following command to log in to your Azure account:

    ```bash
    az login
    ```

    Follow the on-screen instructions to complete the login process. This will authenticate you with your Azure account.

    > If you don't have the Azure CLI installed, you can [install it from here](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli).

2. Once you are logged in, Rill will automatically use the credentials obtained from `az login` to authenticate with Azure Blob Storage when you interact with Azure Blob Storage sources.

#### Using Connection String

Alternatively, you can use an Azure Blob Storage connection string to configure the credentials. To do this:

1. Obtain the connection string for your Azure Blob Storage account. You can find this in the Azure Portal under "Access keys" in your storage account settings.

2. Set the `AZURE_STORAGE_CONNECTION_STRING` environment variable in your local environment to the connection string value. You can do this in your terminal:

    ```bash
    export AZURE_STORAGE_CONNECTION_STRING="your_connection_string_here"
    ```

    Replace "your_connection_string_here" with your actual connection string.

3. Rill will automatically use the connection string from the `AZURE_STORAGE_CONNECTION_STRING` environment variable to authenticate with Azure Blob Storage when you interact with Azure Blob Storage sources.

#### Using Shared Access Signature (SAS) Token

Alternatively, you can configure credentials using a Shared Access Signature (SAS) token. To do this:

1. Generate a SAS token for the Azure Blob Storage container or blob you want to access. You can create SAS tokens using the Azure Portal or programmatically using the Azure SDKs.

    > [Learn how to create SAS tokens using this guide](https://learn.microsoft.com/en-us/azure/ai-services/translator/document-translation/how-to-guides/create-sas-tokens?tabs=Containers).

2. Set the `AZURE_STORAGE_SAS_TOKEN` environment variable in your local environment to the SAS token value. You can do this in your terminal:

    ```bash
    export AZURE_STORAGE_SAS_TOKEN="your_sas_token_here"
    ```

    Replace "your_sas_token_here" with your actual SAS token.

3. Rill will use the SAS token from the `AZURE_STORAGE_SAS_TOKEN` environment variable to authenticate with Azure Blob Storage when interacting with Azure Blob Storage sources.

### Configure credentials for deployments on Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide an Azure Blob Storage connection string, Azure Storage Key or Azure Storage SAS token for the Azure Blob Storage containers used in your project. 

When you first deploy a project using `rill deploy`, you will be prompted to provide credentials for the remote sources in your project that require authentication.

If you subsequently add sources that require new credentials (or if you input the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:

```bash
rill env configure
```
Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.
