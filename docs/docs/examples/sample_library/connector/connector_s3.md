---
title: S3 Connector Example YAML
tags:
- connector
- code
- complete_file
- aws
- s3
docs: https://docs.rilldata.com/connect/data-source/s3
hash: d11fee663e3d9b32091fda85b5815291b55cdf08c1fbcbc0887d553fa7ca6e38
---

```yaml
type: connector

driver: s3

aws_access_key_id: "{{ .env.connector.s3.aws_access_key_id }}"
aws_secret_access_key: "{{ .env.connector.s3.aws_secret_access_key }}"
```
