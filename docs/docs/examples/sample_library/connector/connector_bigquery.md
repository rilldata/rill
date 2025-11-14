---
title: GCP BigQuery Connector Example YAML
tags:
- connector
- code
- complete_file
- gcp
- bigquery
docs: https://docs.rilldata.com/connect/data-source/bigquery
hash: d5dbb536870de4b33f676b3ebcce67acd080c3fc0e85a0cc5523b4f4a3458228
---

```yaml
type: connector

driver: bigquery

google_application_credentials: "{{ .env.connector.bigquery.google_application_credentials }}"
project_id: "rilldata"
```
