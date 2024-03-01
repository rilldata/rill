---
title: BigQuery 
description: Connect to data in BigQuery
sidebar_label: BigQuery
sidebar_position: 4
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Google BigQuery](https://cloud.google.com/bigquery/docs) is a fully-managed, serverless data warehouse that enables scalable and cost-effective analysis of large datasets using SQL-like queries. It supports a highly scalable and flexible architecture, allowing users to analyze large amounts of data in real-time and making it suitable for BI/ML applications. Rill supports natively connecting to and reading from BigQuery as a source by leveraging the [BigQuery SDK](https://cloud.google.com/bigquery/docs/reference/libraries).

![Connecting to BigQuery](/img/reference/connectors/bigquery/bigquery.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), Rill uses the credentials configured in your local environment using the Google Cloud CLI (`gcloud`). Follow these steps to configure it:

1. Open a terminal window and run `gcloud auth list` to check if you already have the Google Cloud CLI installed and authenticated. 

2. If it did not print information about your user, follow the steps on [Install the Google Cloud CLI](https://cloud.google.com/sdk/docs/install-sdk). Make sure to run `gcloud init` after installation as described in the tutorial.

You have now configured Google Cloud access from your local environment. Rill will detect and use your credentials next time you try to ingest a source.

## Cloud deployment

When deploying a project to Rill Cloud (i.e. `rill deploy`), Rill requires you to explicitly provide a JSON key file for a Google Cloud service account with access to BigQuery used in your project. 

When you first deploy a project using `rill deploy`, you will be prompted to provide credentials for the remote sources in your project that require authentication.

If you subsequently add sources that require new credentials (or if you input the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:
```
rill env configure
```

:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

## Appendix

### How to create a service account using the Google Cloud Console

Here is a step-by-step guide on how to create a Google Cloud service account with access to BigQuery:

1. Navigate to the [Service Accounts page](https://console.cloud.google.com/iam-admin/serviceaccounts) under "IAM & Admin" in the Google Cloud Console.

2. Click the "Create Service Account" button at the top of the page.

3. In the "Create Service Account" window, enter a name for the service account, then click "Create and continue".

4. In the "Role" field, search for and select the following [BigQuery roles](https://cloud.google.com/bigquery/docs/access-control): 
   - [roles/bigquery.dataViewer](https://cloud.google.com/bigquery/docs/access-control#bigquery.dataViewer) (Lowest-level resources: Table, View)
     - provides the ability to read data and metadata from the project's datasets/dataset's tables/table or view
   - [roles/bigquery.readSessionUser](https://cloud.google.com/bigquery/docs/access-control#bigquery.readSessionUser) (Lowest-level resources: Project)
     - provides the ability to create and use read sessions that can be used to read data from BigQuery managed tables using the Storage API (to read data from BigQuery at high speeds). The role does not provide any other permissions related to BigQuery datasets, tables, or other resources.
   - [roles/bigquery.jobUser](https://cloud.google.com/bigquery/docs/access-control#bigquery.jobUser) (Lowest-level resources: Project)
     - provides permissions to run BigQuery-specific jobs (including queries), within the project and respecting limits set by roles above

   Click "Continue", then click "Done".

   **Note**: BigQuery has storage and compute [separated](https://cloud.google.com/blog/products/bigquery/separation-of-storage-and-compute-in-bigquery) from each other so the lowest-level resource where compute-specific roles are granted is a project, while lowest-level for data-specific roles is table/view.

5. On the "Service Accounts" page, locate the service account you just created and click on the three dots on the right-hand side. Select "Manage Keys" from the dropdown menu.

6. On the "Keys" page, click the "Add key" button and select "Create new key".

7. Choose the "JSON" key type and click "Create".

8. Download and save the JSON key file to a secure location on your computer.
