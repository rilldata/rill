---
title: Amazon Athena
description: Connect to data in Amazon Athena
sidebar_label: Athena
sidebar_position: 5
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Amazon Athena](https://docs.aws.amazon.com/athena/) is an interactive query service that makes it easy to analyze data directly in Amazon S3 using standard SQL. It is serverless, so there is no infrastructure to manage, and you pay only for the queries you run, making it cost-effective for a wide range of data analysis tasks. Athena is designed for quick, ad-hoc querying of large datasets, enabling businesses to easily integrate it into their analytics and business intelligence tools for immediate insights from their data stored in S3. Rill supports natively connecting to and reading from Athena as a source by leveraging the [AWS SDK for Go](https://aws.github.io/aws-sdk-go-v2/docs/).

![Connecting to Athena](/img/reference/connectors/athena/athena.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), Rill uses the credentials configured in your local environment using the AWS CLI. 

To check if you already have the AWS CLI installed and authenticated, open a terminal window and run:
```bash
aws iam get-user --no-cli-pager
```
If it prints information about your user, there is nothing more to do. Rill will be able to connect to any existing Athena instances that your user has privileges to access.

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

When deploying a project to Rill Cloud (i.e. `rill deploy`), Rill requires you to explicitly provide an access key and secret for an AWS service account with access to Athena used in your project. 

If you subsequently add sources that require new credentials (or if you had simply input the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:
```
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

:::tip Did you know?

If you've configured credentials locally already (in your `<RILL_HOME>/.home` file), you can use `rill env push` to [push these credentials](/build/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::

## Appendix

### Athena/S3 permissions
Athena connector does the following AWS queries while ingesting data from Athena:
1. Athena:[`GetWorkGroup`](https://docs.aws.amazon.com/athena/latest/APIReference/API_GetWorkGroup.html) to determine an output location if not specified explicitly.
2. S3:[`ListObjects`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html) to identify files unloaded by Athena
3. S3:[`GetObject`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html) to ingest files unloaded by Athena.

Make sure your account or a service account have corresponding permissions to perform these requests.
