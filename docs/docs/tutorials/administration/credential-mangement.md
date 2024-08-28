---
title: "Credential Management"
description: how to use credentials in Developer vs Cloud
sidebar_label: "Credential Management"
---
Please review the documentation on [Credential Managment](https://docs.rilldata.com/build/credentials/) before getting started.

### Configuring Credentials


Assuming that you have worked through all of the courses up until now, when running the following command you will get prompted for the following:

```bash
rill env configure
Finish deploying your project by providing access to the connectors. Rill requires credentials for the following connectors:

 - bigquery (used by SQL-incremental-tutorial)
 - clickhouse (used by staging_to_CH)
 - gcs (used by commits__ and others)
 - s3 (used by staging_to_CH)
 - snowflake (used by staging_to_CH)
```

When running locally, Rill will pull environmental credentials or use locally defined credentials. However, when pushing to Rill Cloud, not all of these are available so you will need to define these. 

### Authenticating
Please refer to each connection's documentation on what is required to authenticate. 

```bash
Configuring connector "bigquery":
For instructions on how to configure, see: https://docs.rilldata.com/reference/connectors/bigquery
? connector.bigquery.google_application_credentials (Enter path of file to load from.) 

Configuring connector "clickhouse":
For instructions on how to configure, see: https://docs.rilldata.com/reference/olap-engines/clickhouse
? connector.clickhouse.host 
? connector.clickhouse.port 
? connector.clickhouse.username 
? connector.clickhouse.password 
? connector.clickhouse.ssl 

Configuring connector "gcs":
For instructions on how to configure, see: https://docs.rilldata.com/reference/connectors/gcs
? connector.gcs.google_application_credentials (Enter path of file to load from.) 

Configuring connector "s3":
For instructions on how to configure, see: https://docs.rilldata.com/reference/connectors/s3
? connector.s3.aws_access_key_id 
? connector.s3.aws_secret_access_key 

Configuring connector "snowflake":
? connector.snowflake.dsn 
```

Once you have completed, this you can head back over to Rill Cloud and check if your sources are connecting successfully. 

### Making Changes Locally

Sometimes, you will need to makes changes to the project locally and the credentials were not setup. In this case, if the project credentials are already configured, you can run the following to pull them locally. 

```bash
rill env pull
```

Conversly, if you've made any small changes to the .env file, you can push the changes using:

```bash
rill env push
```

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />