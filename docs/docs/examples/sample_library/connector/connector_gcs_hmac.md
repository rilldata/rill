---
title: GCS Connector using HMAC Example YAML
tags:
- connector
- code
- complete_file
- gcp
- gcs
docs: https://docs.rilldata.com/connect/data-source/gcs
hash: d9c8fe1b7aae8979a24d46223c8b33a8da736938e805bd211584f87920b56ee0
---

```yaml
type: connector

driver: gcs

key_id: "{{ .env.connector.gcs.key_id }}"
secret: "{{ .env.connector.gcs.secret }}"
bucket: "*"
```
