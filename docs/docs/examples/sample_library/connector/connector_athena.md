---
title: AWS Athena Connector Example YAML
tags:
- connector
- code
- complete_file
- aws
- athena
docs: https://docs.rilldata.com/connect/data-source/athena
hash: 6ce76ae692a8eaa9b18c8218fe54bcf0e458570394d8631fe792a6eb1cecc8c4
---

```yaml
type: connector

driver: athena

aws_access_key_id: "{{ .env.connector.athena.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.athena.aws_secret_access_key }}"
output_location: "s3://bucket/path/folder"
region: "us-east-1"
```
