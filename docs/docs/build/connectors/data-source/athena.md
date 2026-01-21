---
title: Athena
description: Connect to Amazon Athena for serverless querying of data stored in S3
sidebar_label: Athena
sidebar_position: 0
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Amazon Athena](https://docs.aws.amazon.com/athena/) is an interactive query service that makes it easy to analyze data directly in Amazon S3 using standard SQL. It is serverless, so there is no infrastructure to manage, and you pay only for the queries you run, making it cost-effective for a wide range of data analysis tasks. Athena is designed for quick, ad-hoc querying of large datasets, enabling businesses to easily integrate it into their analytics and business intelligence tools for immediate insights from their data stored in S3. Rill supports natively connecting to and reading from Athena as a source by leveraging the [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/docs/).

## Authentication Methods

To connect to Amazon Athena, you need to provide authentication credentials. Rill supports two methods:

1. **Use Access Key/Secret Key** (recommended for production)
2. **Use Local AWS credentials** (local development only - not recommended for production)

:::tip Authentication Methods
Choose the method that best fits your setup. For production deployments to Rill Cloud, use Access Key/Secret Key. Local AWS credentials only work for local development and will cause deployment failures.
:::

## Using the Add Data UI

When you add an Athena data model through the Rill UI, the process follows two steps:

1. **Configure Authentication** - Set up your Athena connector with AWS credentials (Access Key/Secret Key)
2. **Configure Data Model** - Define which database, table, or query to execute

This two-step flow ensures your credentials are securely stored in the connector configuration, while your data model references remain clean and portable.

---

## Method 1: Access Key and Secret Key (Recommended)

Access Key and Secret Key credentials provide the most reliable authentication for Athena. This method works for both local development and Rill Cloud deployments.

### Using the UI

1. Click **Add Data** in your Rill project
2. Select **Amazon Athena** as the data source type
3. In the authentication step:
   - Enter your AWS Access Key ID
   - Enter your AWS Secret Access Key
   - Specify the output location (S3 bucket for query results)
   - Specify the AWS region
4. In the data model configuration step, enter your SQL query
5. Click **Create** to finalize

After the model YAML is generated, you can add additional [model settings](/build/models/source-models) directly to the file.

### Manual Configuration

If you prefer to configure manually, create two files:

**Step 1: Create connector configuration**

Create `connectors/my_athena.yaml`:

```yaml
type: connector
driver: athena

aws_access_key_id: "{{ .env.connector.athena.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.athena.aws_secret_access_key }}"
output_location: "s3://my-bucket/athena-results/"
region: "us-east-1"
```

**Step 2: Create model configuration**

Create `models/my_athena_data.yaml`:

```yaml
type: model
connector: my_athena

sql: SELECT * FROM my_database.my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

**Step 3: Add credentials to `.env`**

```bash
connector.athena.aws_access_key_id=AKIAIOSFODNN7EXAMPLE
connector.athena.aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

:::tip Did you know?
If this project has already been deployed to Rill Cloud and credentials have been set for this connector, you can use `rill env pull` to [pull these cloud credentials](/build/connectors/credentials#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.
:::

---

## Method 2: Local AWS Credentials

For local development, you can use credentials from the AWS CLI. This method is **not suitable for production** or Rill Cloud deployments.

:::warning Not recommended for production
Local AWS credentials only work for local development. If you deploy to Rill Cloud using this method, your dashboards will fail. Always use Access Key/Secret Key for production deployments.
:::

### Setup

1. Install the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) if not already installed
2. Authenticate with your AWS account:
   - If your organization has SSO configured, reach out to your admin for instructions on how to authenticate using `aws sso login`
   - Otherwise, run `aws configure` and provide your access key, secret, and default region
3. Verify your authentication:
   ```bash
   aws iam get-user --no-cli-pager
   ```

### Connector Configuration

Create `connectors/my_athena.yaml`:

```yaml
type: connector
driver: athena

output_location: "s3://my-bucket/athena-results/"
region: "us-east-1"
```

### Model Configuration

Create `models/my_athena_data.yaml`:

```yaml
type: model
connector: my_athena

sql: SELECT * FROM my_database.my_table

# Add a refresh schedule
refresh:
  cron: "0 */6 * * *"
```

When no explicit credentials are provided in the connector, Rill will automatically use your local AWS CLI credentials.

---

## Using Athena Data in Models

Once your connector is configured, you can reference Athena tables and run queries in your model configurations.

### Basic Example

```yaml
type: model
connector: my_athena

sql: SELECT * FROM my_database.my_table

refresh:
  cron: "0 */6 * * *"
```

### Custom SQL Query

```yaml
type: model
connector: my_athena

sql: |
  SELECT
    date_trunc('day', event_time) as event_date,
    event_type,
    COUNT(*) as event_count
  FROM my_database.events
  WHERE event_time >= date_add('day', -30, current_date)
  GROUP BY 1, 2

refresh:
  cron: "0 */6 * * *"
```

---

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

---

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide an access key and secret for an AWS service account with access to Athena used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#athena) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials by pushing the `Deploy` button to update your project or by running the following command in the CLI:
```
rill env push
```

:::tip Did you know?
If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.
:::

---

## Appendix

### Athena/S3 Permissions

The Athena connector performs the following AWS queries while ingesting data from Athena:

1. Athena: [`GetWorkGroup`](https://docs.aws.amazon.com/athena/latest/APIReference/API_GetWorkGroup.html) to determine an output location if not specified explicitly.
2. S3: [`ListObjects`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html) to identify files unloaded by Athena.
3. S3: [`GetObject`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html) to ingest files unloaded by Athena.

Make sure your account or service account has the corresponding permissions to perform these requests.

### How to Create an AWS Service Account

For detailed instructions on creating an AWS service account with the appropriate permissions, see the [S3 connector documentation](./s3#how-to-create-an-aws-service-account-using-the-aws-management-console).
