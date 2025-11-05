---
title: GCS Connector using Service Account Example YAML
tags:
- connector
- code
- complete_file
- gcp
- gcs
docs: https://docs.rilldata.com/connect/data-source/gcs
hash: 08647a5920f36d26b2850015fc5f5a1fc1331fc02a4522f33421a8fd75dd346c
---

```yaml
type: connector

driver: gcs

google_application_credentials: "{{ .env.connector.gcs.google_application_credentials }}"
bucket: "gs://bucket"
```
