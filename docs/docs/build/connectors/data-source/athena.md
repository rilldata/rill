---
title: Athena
description: Connect to Amazon Athena for serverless querying of data stored in S3
sidebar_label: Athena
sidebar_position: 0
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Amazon Athena](https://docs.aws.amazon.com/athena/) is an interactive query service that makes it easy to analyze data directly in Amazon S3 using standard SQL. It is serverless, so there is no infrastructure to manage, and you pay only for the queries you run, making it cost-effective for a wide range of data analysis tasks. Athena is designed for quick, ad-hoc querying of large datasets, enabling businesses to easily integrate it into their analytics and business intelligence tools for immediate insights from their data stored in S3. Rill supports natively connecting to and reading from Athena as a source by leveraging the [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/docs/).


## Connect to Athena

To connect to Amazon Athena, you need to provide authentication credentials. You have two options:

1. **Use Access Key/Secret Key** (recommended for cloud deployment)
2. **Use Local AWS credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Access Key/Secret Key. Local AWS credentials only work for local development and will cause deployment failures.

### Access Key and Secret Key

Create a connector with your credentials to connect to Athena. Here's an example connector configuration file you can copy into your `connectors` directory to get started:

```yaml
type: connector

driver: athena
aws_access_key_id: "{{ .env.connector.athena.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.athena.aws_secret_access_key }}"
output_location: "s3://bucket/path/folder"
region: "us-east-1"
```

:::tip Using the Add Data Form
You can also use the Add Data form in Rill Developer, which will automatically create the `athena.yaml` file and populate the `.env` file with `connector.athena.aws_access_key_id` and `connector.athena.aws_secret_access_key`.
:::

### Local AWS Credentials (Local Development Only)

:::warning Not recommended for production
Local AWS credentials only work for local development. If you deploy to Rill Cloud using this method, your dashboards will fail. Use Method 1 above for production deployments.
:::

When using Rill Developer on your local machine (i.e., `rill start`), Rill can either use the credentials configured in your local environment using the AWS CLI or use the explicitly set credentials in a [connector](/reference/project-files/connectors#athena) file.

To check if you already have the AWS CLI installed and authenticated, open a terminal window and run:
```bash
aws iam get-user --no-cli-pager
```
If it prints information about your user, there is nothing more to do. Rill will be able to connect to any existing Athena instances that your user has privileges to access.

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

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this connector, you can use `rill env pull` to [pull these cloud credentials](/build/connectors/credentials#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.

:::

## Separating Dev and Prod Environments

When ingesting data locally, consider setting parameters in your connector file to limit how much data is retrieved, since costs can scale with the data source. This also helps other developers clone the project and iterate quickly by reducing ingestion time.

For more details, see our [Dev/Prod setup docs](/build/connectors/templating).

## Cloud deployment

When deploying a project to Rill Cloud, Rill requires you to explicitly provide an access key and secret for an AWS service account with access to Athena used in your project. Please refer to our [connector YAML reference docs](/reference/project-files/connectors#athena) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:
```
rill env configure
```

:::tip Did you know?

If you've already configured credentials locally (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/build/connectors/credentials#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve and reuse the same credentials automatically by running `rill env pull`.

:::

## Appendix

### Athena/S3 permissions
The Athena connector performs the following AWS queries while ingesting data from Athena:
1. Athena: [`GetWorkGroup`](https://docs.aws.amazon.com/athena/latest/APIReference/API_GetWorkGroup.html) to determine an output location if not specified explicitly.
2. S3: [`ListObjects`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html) to identify files unloaded by Athena.
3. S3: [`GetObject`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html) to ingest files unloaded by Athena.

Make sure your account or service account has the corresponding permissions to perform these requests.
