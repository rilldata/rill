---
title: All Inclusive Example env
tags:
- project
- code
- complete_file
docs: https://docs.rilldata.com/manage/project-management/variables-and-credentials
hash: 3db1d37227dd3fdb2bdfb353999bab3ec8793e7bd52c7cea6bd82d57a8938184
---

```
# AWS S3 credentials
connector.s3.access_key_id=AKIAIOSFODNN7EXAMPLE
connector.s3.secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# Google Cloud credentials
connector.gcs.credentials_json={"type":"service_account","project_id":"my-project"}

# Database connection
connector.postgres.dsn=postgres://username:password@localhost:5432/mydb

# Custom variables
my_custom_variable=some_value
```
