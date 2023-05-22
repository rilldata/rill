---
title: Configure connector credentials
sidebar_label: Configure connector credentials
sidebar_position: 10
---

Rill requires credentials to connect to remote data sources such as private buckets in S3 or GCS.

When running Rill locally, Rill attempts to find existing credentials configured on your computer. When deploying projects to Rill Cloud, you must explicitly provide service account credentials with correct access permissions.

For instructions on how to create a service account and set credentials in Rill Cloud, see our reference docs for:

- [Amazon S3](../connectors/s3.md)
- [Google Cloud Storage (GCS)](../connectors/gcs.md)
