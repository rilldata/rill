---
title: Configure connector credentials
sidebar_label: Configure connector credentials
sidebar_position: 10
---

Rill requires credentials to connect to remote data sources such as private buckets in S3 or GCS.

When running Rill locally, Rill attempts to find existing credentials configured on your computer. When deploying projects to Rill Cloud, you must explicitly provide service account credentials with correct access permissions.

## Updating connector credentials in Rill Cloud

When you first deploy a project using `rill deploy`, you will be prompted to provide credentials for the remote sources in your project that require authentication.

If you subsequently add sources that require new credentials (or if you input the wrong credentials during the initial deploy), you can update the credentials used by Rill Cloud by running:
```
rill env configure
```

Make sure the credentials you provide are not only valid, but also have access to all the source data used in your project.

## Getting connector credentials

For instructions on how to create a service account, setting permissions, and getting credentials, see our reference docs for:

- [Google Cloud Storage (GCS)](../reference/connectors/gcs.md)
- [Amazon S3](../reference/connectors/s3.md)
