---
title: BigQuery 
description: Connect to data in BigQuery
sidebar_label: BigQuery
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Google BigQuery](https://cloud.google.com/bigquery/docs) is a fully managed, serverless data warehouse that enables scalable and cost-effective analysis of large datasets using SQL-like queries. It supports a highly scalable and flexible architecture, allowing users to analyze large amounts of data in real time, making it suitable for BI/ML applications. Rill supports natively connecting to and reading from BigQuery as a source by leveraging the [BigQuery SDK](https://cloud.google.com/bigquery/docs/reference/libraries).



## Connect to BigQuery

To connect to Google BigQuery, you need to provide authentication credentials. You have two options:

1. **Use Service Account JSON** (recommended for cloud deployment)
2. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

### Service Account JSON 

We recommend using Service Account JSON for authentication as it makes deployment to Rill Cloud easier. The `GOOGLE_APPLICATION_CREDENTIALS` environment variable tells Google Cloud SDK which service account key file to use for authentication.

Create your Service Account JSON with the following command:

```bash
gcloud iam service-accounts keys create ~/key.json \
  --iam-account=my-service-account@PROJECT_ID.iam.gserviceaccount.com
```

:::note Permission denied?
You'll need to contact your internal cloud admin to create your Service Account JSON credentials for you.
:::


Create a connector with your credentials to connect to BigQuery. Here's an example connector configuration file you can copy into your `connectors` directory to get started. The UI will also populate your `.env` with `connector.bigquery.google_application_credentials`.

```yaml
type: connector

driver: bigquery

google_application_credentials: "{{ .env.connector.bigquery.google_application_credentials }}"
project_id: "rilldata"
```

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this connector, you can use `rill env pull` to [pull these cloud credentials](/build/connectors/credentials/#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.

:::


### Local Google Cloud CLI Credentials (Local Development Only)

:::warning Not recommended for production
Local Google Cloud CLI credentials only work for local development. If you deploy to Rill Cloud using this method, your dashboards will fail. Use Method 1 above for production deployments.
:::

Follow these steps to configure your local environment credentials:

1. Open a terminal window and run `gcloud auth list` to check if you already have the Google Cloud CLI installed and authenticated.
2. If it does not print information about your user, follow the steps on [Install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk). Make sure to run `gcloud init` after installation as described in the tutorial.
3. **Important**: Run `gcloud auth application-default login` to set up Application Default Credentials (ADC). If you skip this step, the app will error with missing `GOOGLE_APPLICATION_CREDENTIALS`.

:::tip Service Accounts
If you are using a service account, you will need to run the following command instead:
```
gcloud auth activate-service-account --key-file=path_to_json_key_file
```
:::

You have now configured Google Cloud access from your local environment. Rill will detect and use your credentials the next time you try to ingest a source.

## Deploy to Rill Cloud

When deploying your project to Rill Cloud, you must provide a JSON key file for a Google Cloud service account with access to BigQuery used in your project. If these credentials exist in your `.env` file, they'll be pushed with your project automatically. If you're using inferred credentials only, your deployment will result in errored dashboards.

When you first deploy a project using `rill deploy`, you will be prompted to provide credentials for the remote sources in your project that require authentication.

If you subsequently add sources that require new credentials (or if you enter the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:
```bash
rill env configure
```

:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

## Appendix

### How to create a service account using the Google Cloud Console

Here is a step-by-step guide on how to create a Google Cloud service account with access to BigQuery:

1. Navigate to the [Service Accounts page](https://console.cloud.google.com/iam-admin/serviceaccounts) under "IAM & Admin" in the Google Cloud Console.

2. Click the "Create Service Account" button at the top of the page.

3. In the "Create Service Account" window, enter a name for the service account, then click "Create and continue".

4. In the "Role" field, search for and select the following [BigQuery roles](https://cloud.google.com/bigquery/docs/access-control): 
   - [roles/bigquery.dataViewer](https://cloud.google.com/bigquery/docs/access-control#bigquery.dataViewer) (Lowest-level resources: Table, View)
     - Provides the ability to read data and metadata from the project's datasets/dataset's tables/table or view.
   - [roles/bigquery.readSessionUser](https://cloud.google.com/bigquery/docs/access-control#bigquery.readSessionUser) (Lowest-level resources: Project)
     - Provides the ability to create and use read sessions that can be used to read data from BigQuery managed tables using the Storage API (to read data from BigQuery at high speeds). The role does not provide any other permissions related to BigQuery datasets, tables, or other resources.
   - [roles/bigquery.jobUser](https://cloud.google.com/bigquery/docs/access-control#bigquery.jobUser) (Lowest-level resources: Project)
     - Provides permissions to run BigQuery-specific jobs (including queries), within the project and respecting limits set by roles above.

   Click "Continue", then click "Done".

   **Note**: BigQuery has storage and compute [separated](https://cloud.google.com/blog/products/bigquery/separation-of-storage-and-compute-in-bigquery) from each other, so the lowest-level resource where compute-specific roles are granted is a project, while the lowest-level for data-specific roles is table/view.

5. On the "Service Accounts" page, locate the service account you just created and click on the three dots on the right-hand side. Select "Manage Keys" from the dropdown menu.

6. On the "Keys" page, click the "Add key" button and select "Create new key".

7. Choose the "JSON" key type and click "Create".

8. Download and save the JSON key file to a secure location on your computer.
