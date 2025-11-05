---
title: Azure Storage Account Key Example YAML
tags:
- connector
- code
- complete_file
- azure
- blob_storage
docs: https://docs.rilldata.com/connect/data-source/azure
hash: 7eaf6d498319558b8cd4c84ed2443d035ea258be76566b6c7223ecfe75b09f47
---

```yaml
type: connector

driver: azure

azure_storage_account: storage_account_name
azure_storage_key: "{{ .env.connector.azure.azure_storage_key }}"
```
