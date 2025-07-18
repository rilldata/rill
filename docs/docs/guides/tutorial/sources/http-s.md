---
title: https://
description: Connect to remote source
sidebar_label: https://
sidebar_position: 18
---


## Adding a remote source

### Using the UI
To add a remote source using the UI, click "+" by Sources in the left-hand navigation pane and select the location where your remote files are stored ("Google Cloud Storage", "Amazon S3", or "http(s)"). Enter your file's URI and click "Add Source".

After import, you can reimport your data whenever you want by clicking the "refresh source" button in the Rill UI.





## Authenticating remote sources

Rill requires an appropriate set of <u>credentials</u> to connect to remote data sources, whether those are buckets (e.g. S3 or GCS) or data warehouses (e.g. Snowflake). When running Rill locally, Rill Developer attempts to find existing credentials that have been configured on your machine. When deploying projects to Rill Cloud, you must explicitly provide service account credentials with correct access permissions.