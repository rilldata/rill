---
title: "BigQuery"
description: Connect to a BigQuery dataset
sidebar_label: "BigQuery"
sidebar_position: 10
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

[Google BigQuery](https://cloud.google.com/bigquery/docs) is a fully managed, serverless data warehouse that enables scalable and cost-effective analysis of large datasets using SQL-like queries. It supports a highly scalable and flexible architecture, allowing users to analyze large amounts of data in real time, making it suitable for BI/ML applications. Rill supports natively connecting to and reading from BigQuery as a source by leveraging the [BigQuery SDK](https://cloud.google.com/bigquery/docs/reference/libraries).

## Local Modeling

Create a connector YAML file in your Rill project directory (e.g. `connectors/bigquery.yaml`) defining BigQuery properties and credentials. Replace the placeholder text in this example with your actual BigQuery credentials:

```yaml
type: connector
driver: bigquery
project_id: "{{ .env.connector.bigquery.project_id }}"
location: "us"
google_application_credentials: "{{ .env.connector.bigquery.google_application_credentials }}"
```

The `google_application_credentials` parameter expects a base64 encoded string. If you're using a service account key JSON file, you can convert it to base64 with one of these commands:

<Tabs>
<TabItem value="mac" label="Mac">

```bash
base64 -i /path/to/your/service-account-key.json
```

</TabItem>
<TabItem value="linux" label="Linux">

```bash
base64 /path/to/your/service-account-key.json
```

</TabItem>
<TabItem value="windows" label="Windows (PowerShell)">

```powershell
[Convert]::ToBase64String([System.IO.File]::ReadAllBytes("C:\path\to\your\service-account-key.json"))
```

</TabItem>
</Tabs>

Set these BigQuery credentials in your `.env` file:

```bash
connector.bigquery.project_id=<BigQuery_Project_ID>
connector.bigquery.google_application_credentials=<Base64_Encoded_Service_Account_Key>
```

Create a model YAML file (e.g. `models/my_bigquery_data.yaml`) that references your BigQuery connector to pull data from your BigQuery table or view:

```yaml
type: model
connector: bigquery
sql: SELECT * FROM `project.dataset.table`
```

:::tip
For large tables, consider adding LIMIT clauses during development or using WHERE conditions to reduce data transfer and costs.
:::

## Deployment

Once you're ready to deploy your project to Rill Cloud, you can set the credentials as secrets using the `rill env configure` command:

```bash
rill env configure
```

The CLI will walk you through configuring each connector used in your project. For BigQuery, you'll be prompted to provide:
- Your BigQuery project ID
- Your service account credentials (base64 encoded)

After configuring your credentials, deploy your project:

```bash
rill deploy
```

## Additional Resources

- [Google BigQuery Documentation](https://cloud.google.com/bigquery/docs)
- [BigQuery Go Client Library](https://pkg.go.dev/cloud.google.com/go/bigquery)
- [Creating and Managing Service Accounts](https://cloud.google.com/iam/docs/creating-managing-service-accounts)
