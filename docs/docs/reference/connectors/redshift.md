---
title: Amazon Redshift
description: Connect to data in Amazon Redshift
sidebar_label: Redshift
sidebar_position: 6
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Amazon Redshift](https://docs.aws.amazon.com/redshift/) is a fully managed, petabyte-scale data warehouse service in the cloud, offering fast query and I/O performance for data analysis applications. It enables users to run complex analytical queries against structured data using SQL, ETL processes, and BI tools, leveraging massively parallel processing (MPP) to efficiently handle large volumes of data. Redshift's architecture is designed for high performance on large datasets, supporting data warehousing and analytics of all sizes, making it a pivotal component in a modern data-driven decision-making ecosystem. By leveraging the AWS SDK for Go and utilizing intermediary parquet files in S3 (to ensure performance), Rill is able to connect to and read from Redshift as a source.

![Connecting to Redshift](/img/reference/connectors/redshift/redshift.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), Rill uses the credentials configured in your local environment using the AWS CLI. 

To check if you already have the AWS CLI installed and authenticated, open a terminal window and run:
```bash
aws iam get-user --no-cli-pager
```
If it prints information about your user, there is nothing more to do. Rill will be able to connect to any existing Redshift databases that your user has privileges to access.

If you do not have the AWS CLI installed and authenticated, follow these steps:

1. Open a terminal window and [install the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) if it is not already installed on your system.

2. If your organization has SSO configured, reach out to your admin for instructions on how to authenticate using `aws sso login`.

3. If your organization does not have SSO configured:

    a. Follow the steps described in [How to create an AWS service account using the AWS Management Console](./s3.md#how-to-create-an-aws-service-account-using-the-aws-management-console), which you will find below on this page.

    b. Run the following command and provide the access key, access secret, and default region when prompted (you can leave the "Default output format" blank):
    ```
    aws configure
    ```

You have now configured AWS access from your local environment. Rill will detect and use your credentials next time you try to ingest a source.

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](/build/credentials/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials that you have set locally for this source.

:::

## Cloud deployment

When deploying a project to Rill Cloud (i.e. `rill deploy`), Rill requires you to explicitly provide an access key and secret for an AWS service account with access to the Redshift database used in your project. 

When you first deploy a project using `rill deploy`, you will be prompted to provide credentials for the remote sources in your project that require authentication. If you subsequently add sources that require new credentials (or if you had simply input the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:
```
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

:::tip Did you know?

If you've configured credentials locally already (in your `<RILL_PROJECT_DIRECTORY>/.env` file), you can use `rill env push` to [push these credentials](/build/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::

## Appendix

### Redshift Serverless permissions
When using **Redshift Serverless**, make sure to associate an [IAM role (that has S3 access)](https://docs.aws.amazon.com/redshift/latest/mgmt/serverless-iam.html) with the Serverless namespace or the Redshift cluster. 

:::info What happens when Rill is reading from Redshift Serverless?

Our Redshift connector will place temporary files in parquet format in S3 to help accelerate the extraction process (maximizes performance). To provide some more details, the Redshift connector will execute the following queries / requests while ingesting data from Redshift:

1. Redshift Serverless:[`GetCredentials`](https://docs.aws.amazon.com/redshift-data/latest/APIReference/API_ExecuteStatement.html) if you are using _Workgroup_ name to connect. 
2. Reshift Data API:[`DescribeStatement`, `ExecuteStatement`](https://docs.aws.amazon.com/redshift-data/latest/APIReference/API_ExecuteStatement.html) to unload data to S3.
3. S3:[`ListObjects`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html) to identify files unloaded by Redshift.
4. S3:[`GetObject`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html) to ingest files unloaded by Redshift.

:::

:::warning Check your service account permissions

Your account or service account will need to have the <u>appropriate permissions</u> necessary to perform these requests.

:::

### Redshift Cluster permissions

Similarly, when using **Redshift Cluster**, make sure to associate an [IAM role (that has S3 access)](https://docs.aws.amazon.com/redshift/latest/mgmt/redshift-iam-authentication-access-control.html) with the appropriate Redshift cluster.

:::info What happens when Rill is reading from a Redshift Cluster?

Our Redshift connector will place temporary files in parquet format in S3 to help accelerate the extraction process (maximizes performance). To provide some more details, the Redshift connector will execute the following queries / requests while ingesting data from Redshift:

1. Redshift:[`GetClusterCredentialsWithIAM`](https://docs.aws.amazon.com/redshift-data/latest/APIReference/API_ExecuteStatement.html) if you are using _Cluster Identifier_ to connect. 
2. Redshift Data API:[`DescribeStatement`, `ExecuteStatement`](https://docs.aws.amazon.com/redshift-data/latest/APIReference/API_ExecuteStatement.html) to unload data to S3.
3. S3:[`ListObjects`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html) to identify files unloaded by Redshift.
4. S3:[`GetObject`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html) to ingest files unloaded by Redshift.

:::

:::warning Check your service account permissions

Your account or service account will need to have the <u>appropriate permissions</u> necessary to perform these requests.

:::


