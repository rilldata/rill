---
title: Google Cloud Storage
sidebar_label: Google Cloud Storage
sidebar_position: 10
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Overview

Rill supports ingesting data from Google Cloud Storage (GCS) buckets. You can connect to GCS using one of three authentication methods:

1. **Use Service Account JSON** (recommended for production)
2. **Use HMAC Keys** (alternative authentication method)
3. **Use Local Google Cloud CLI credentials** (local development only - not recommended for production)

Choose the method that best fits your setup. For production deployments to Rill Cloud, use Service Account JSON or HMAC Keys. Local Google Cloud CLI credentials only work for local development and will cause deployment failures.

## Using the Add Data UI

When using Rill's Add Data UI, authentication follows a two-step process:

1. **Step 1: Configure Authentication** - Set up your GCS connector with credentials
2. **Step 2: Configure Data Model** - Create your data model pointing to specific GCS files

This separation allows you to:
- Reuse the same connector across multiple data models
- Manage credentials independently from data models
- Update authentication without modifying data model configurations

## Authentication Methods

### Service Account JSON

Service Account authentication is recommended for production deployments to Rill Cloud.

#### Using the UI

1. Navigate to the Add Data UI in Rill
2. Select **Google Cloud Storage** as your data source
3. In Step 1 (Configure Authentication):
   - Choose **Service Account JSON** as the authentication method
   - Upload your JSON key file using the file picker
   - Click **Save** to create the connector
4. In Step 2 (Configure Data Model):
   - Enter your GCS path (e.g., `gs://bucket-name/path/to/file.parquet`)
   - Configure additional model settings
   - Click **Create** to finalize

#### Manual Configuration

<Tabs>
<TabItem value="connector" label="Connector File (connectors/gcs.yaml)">

```yaml
type: gcs
google_application_credentials: |
  {
    "type": "service_account",
    "project_id": "your-project-id",
    "private_key_id": "key-id",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
    "client_email": "your-service-account@your-project.iam.gserviceaccount.com",
    "client_id": "client-id",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/your-service-account%40your-project.iam.gserviceaccount.com"
  }
