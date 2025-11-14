---
title: Azure Shared Access Services Token Example YAML
tags:
- connector
- code
- complete_file
- azure
- sas_token
docs: https://docs.rilldata.com/connect/data-source/azure
hash: 092803522ae7ed9145a357f0f6f2063e13110266be5fd6d3eb761d31637ed624
---

```yaml
type: connector

driver: azure

azure_storage_account: rilltest
azure_storage_sas_token: "{{ .env.connector.azure.azure_storage_sas_token }}"
```
