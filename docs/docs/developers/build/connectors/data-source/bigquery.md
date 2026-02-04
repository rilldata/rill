---
title: BigQuery
description: Connect to data in BigQuery
sidebar_label: BigQuery
sidebar_position: 10
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Google BigQuery](https://cloud.google.com/bigquery/docs) is a fully managed, serverless data warehouse that enables scalable and cost-effective analysis of large datasets using SQL-like queries. It supports a highly scalable and flexible architecture, allowing users to analyze large amounts of data in real time, making it suitable for BI/ML applications. Rill supports natively connecting to and reading from BigQuery as a source by leveraging the [BigQuery SDK](https://cloud.google.com/bigquery/docs/reference/libraries).

## Authentication Methods

To connect to Google BigQuery, you need to provide authentication credentials. Rill supports two methods:

1. **Use Service Account JSON** (recommended for production)
2. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

:::tip Authentication Methods
Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.
:::

## Using the Add Data UI

When you add a BigQuery data model through the Rill UI, the process follows two steps:

1. **Configure Authentication** - Set up your BigQuery connector with credentials (Service Account JSON)
2. **Configure Data Model** - Define which dataset, table, or query to execute

This two-step flow ensures your credentials are securely stored in the connector configuration, while your data model references remain clean and portable.

## Method 1: Service Account JSON (Recommended)

Service Account JSON credentials provide the most secure and reliable authentication for BigQuery. This method works for both local development and Rill Cloud deployments.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Google BigQuery** as the data source type
3. In the authentication step:
   - Upload your JSON key file or paste its contents
   - Specify your Google Cloud Project ID
4. In the data model configuration step, enter your SQL query
5. Click **Create** to finalize

After the model YAML is generated, you can add additional [model settings](/developers/build/models/source-models) directly to the file.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/my_bigquery.yaml`:

```yaml
type: connector
driver: bigquery

google_application_credentials: "{{ .env.connector.bigquery.google_application_credentials }}"
project_id: "my-gcp-project"
```

**Step 2: Create model configuration**

Create `models/my_bigquery_data.yaml`:

```yaml
type: model
connector: my_bigquery

sql: SELECT * FROM my_dataset.my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.bigquery.google_application_credentials=<json_credentials>
```

:::tip Did you know?
If this project has already been deployed to Rill Cloud and credentials have been set for this connector, you can use `rill env pull` to [pull these cloud credentials](/developers/build/connectors/credentials/#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.
:::

## Method 2: Local Google Cloud CLI Credentials

For local development, you can use credentials from the Google Cloud CLI. This method is **not suitable for production** or Rill Cloud deployments.

:::warning Not recommended for production
Local Google Cloud CLI credentials only work for local development. If you deploy to Rill Cloud using this method, your dashboards will fail. Always use Service Account JSON for production deployments.
:::

### Setup

1. Install the [Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk) if not already installed
2. Initialize and authenticate:
   ```bash
   gcloud init
   ```
3. **Important**: Set up Application Default Credentials (ADC):
   ```bash
   gcloud auth application-default login
   ```

:::tip Service Accounts
If you are using a service account, run the following command instead:
```bash
gcloud auth activate-service-account --key-file=path_to_json_key_file
```
:::

### Connector Configuration

Create `connectors/my_bigquery.yaml`:

```yaml
type: connector
driver: bigquery

project_id: "my-gcp-project"
```

### Model Configuration

Create `models/my_bigquery_data.yaml`:

```yaml
type: model
connector: my_bigquery

sql: SELECT * FROM my_dataset.my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

When no explicit credentials are provided in the connector, Rill will automatically use your local Google Cloud CLI credentials.

## Using BigQuery Data in Models

Once your connector is configured, you can reference BigQuery tables and run queries in your model configurations.

### Basic Example

```yaml
type: model
connector: my_bigquery

sql: SELECT * FROM my_dataset.my_table

refresh:
  cron: "0 */6 * * *"
```

### Custom SQL Query

```yaml
type: model
connector: my_bigquery

sql: |
  SELECT
    DATE(created_at) as event_date,
    event_type,
    COUNT(*) as event_count
  FROM my_dataset.events
  WHERE created_at >= DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)
  GROUP BY 1, 2

refresh:
  cron: "0 */6 * * *"
```

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/developers/build/connectors/templating).

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide a JSON key file for a Google Cloud service account with access to BigQuery used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#bigquery) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:
```
rill env push
```

:::tip Did you know?
If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/developers/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.
:::

## Appendix

### How to Create a Service Account Using the Google Cloud Console

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

:::note Permission denied?
You'll need to contact your internal cloud admin to create your Service Account JSON credentials for you.
:::
