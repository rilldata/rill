---
title: Amazon Redshift
description: Connect to data in Amazon Redshift
sidebar_label: Redshift
sidebar_position: 55
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Amazon Redshift](https://docs.aws.amazon.com/redshift/) is a fully managed, petabyte-scale data warehouse service in the cloud, offering fast query and I/O performance for data analysis applications. It enables users to run complex analytical queries against structured data using SQL, ETL processes, and BI tools, leveraging massively parallel processing (MPP) to efficiently handle large volumes of data. Redshift's architecture is designed for high performance on large datasets, supporting data warehousing and analytics of all sizes, making it a pivotal component in a modern data-driven decision-making ecosystem. By leveraging the AWS SDK for Go and utilizing intermediary Parquet files in S3 (to ensure performance), you can connect to and read from Redshift data warehouses.


## Connect to Redshift

To connect to Amazon Redshift, you need to provide authentication credentials. You have two options:

1. **Use Access Key/Secret Key** (recommended for cloud deployment)
2. **Use Local AWS credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Access Key/Secret Key. Local AWS credentials only work for local development and will cause deployment failures.

### Access Key and Secret Key


Create a connector with your credentials to connect to Redshift. Here's an example connector configuration file you can copy into your `connectors` directory to get started:



```yaml
type: connector

driver: redshift
aws_access_key_id: "{{ .env.connector.redshift.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.redshift.aws_secret_access_key }}"
database: "dev"
```

:::tip Using the Add Data Form
You can also use the Add Data form in Rill Developer, which will automatically create the `redshift.yaml` file and populate the `.env` file with `connector.redshift.aws_access_key_id` and `connector.redshift.aws_secret_access_key`.
:::

### Local AWS Credentials (Local Development Only)

:::warning Not recommended for production
Local AWS credentials only work for local development. If you deploy to Rill Cloud using this method, your dashboards will fail. Use Method 1 above for production deployments.
:::

When using Rill Developer on your local machine, you can use credentials configured in your local environment using the AWS CLI instead of explicit credentials in the connector.

To check if you already have the AWS CLI installed and authenticated, open a terminal window and run:
```bash
aws iam get-user --no-cli-pager
```
If it prints information about your user, there is nothing more to do. You'll be able to connect to any existing Redshift databases that your user has privileges to access.

If you do not have the AWS CLI installed and authenticated, follow these steps:

1. Open a terminal window and [install the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) if it is not already installed on your system.

2. If your organization has SSO configured, reach out to your admin for instructions on how to authenticate using `aws sso login`.

3. If your organization does not have SSO configured:

    a. Follow the steps described in [How to create an AWS service account using the AWS Management Console](./s3#how-to-create-an-aws-service-account-using-the-aws-management-console), which you will find below on this page.

    b. Run the following command and provide the access key, access secret, and default region when prompted (you can leave the "Default output format" blank):
    ```
    aws configure
    ```

You have now configured AWS access from your local environment. Rill will detect and use your credentials the next time you try to ingest a source.

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

## Deploy to Rill Cloud

When deploying your project to Rill Cloud, you must explicitly provide an access key and secret for an AWS service account with access to the Redshift database used in your project. If these credentials exist in your `.env` file, they'll be pushed with your project automatically.

When you first deploy a project using `rill deploy`, you will be prompted to provide credentials for the remote sources in your project that require authentication. If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:
```bash
rill env configure
```

:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::

## Appendix

:::warning Check your service account permissions

Your account or service account will need to have the **appropriate permissions** necessary to perform these requests.

:::

### Redshift Serverless permissions
When using **Redshift Serverless**, make sure to associate an [IAM role (that has S3 access)](https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-iam.html) with the Serverless namespace or the Redshift cluster.

:::info What happens when Rill is reading from Redshift Serverless?

Our Redshift connector will place temporary files in Parquet format in S3 to help accelerate the extraction process (maximizing performance). To provide more details, the Redshift connector will execute the following queries/requests while ingesting data from Redshift:

1. Redshift Serverless: [`GetCredentials`](https://docs.aws.amazon.com/redshift-data/latest/APIReference/API_ExecuteStatement.html) if you are using a _Workgroup_ name to connect.
2. Redshift Data API: [`DescribeStatement`, `ExecuteStatement`](https://docs.aws.amazon.com/redshift-data/latest/APIReference/API_ExecuteStatement.html) to unload data to S3.
3. S3: [`ListObjects`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html) to identify files unloaded by Redshift.
4. S3: [`GetObject`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html) to ingest files unloaded by Redshift.

:::

### Redshift Cluster permissions

Similarly, when using **Redshift Cluster**, make sure to associate an [IAM role (that has S3 access)](https://docs.aws.amazon.com/redshift/latest/mgmt/redshift-iam-authentication-access-control.html) with the appropriate Redshift cluster.

:::info What happens when Rill is reading from a Redshift Cluster?

Our Redshift connector will place temporary files in Parquet format in S3 to help accelerate the extraction process (maximizing performance). To provide more details, the Redshift connector will execute the following queries/requests while ingesting data from Redshift:

1. Redshift: [`GetClusterCredentialsWithIAM`](https://docs.aws.amazon.com/redshift-data/latest/APIReference/API_ExecuteStatement.html) if you are using a _Cluster Identifier_ to connect.
2. Redshift Data API: [`DescribeStatement`, `ExecuteStatement`](https://docs.aws.amazon.com/redshift-data/latest/APIReference/API_ExecuteStatement.html) to unload data to S3.
3. S3: [`ListObjects`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html) to identify files unloaded by Redshift.
4. S3: [`GetObject`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html) to ingest files unloaded by Redshift.

:::

:::warning Check your service account permissions

Your account or service account will need to have the <u>appropriate permissions</u> necessary to perform these requests.

:::


