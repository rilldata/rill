---
title: "Credential and Environmental Variable Management"
description: how to use credentials in Developer vs Cloud
sidebar_label: "Credential and Environmetal Variable Management"
---
import ComingSoon from '@site/src/components/ComingSoon';


Please review the documentation on [Credential Managment](https://docs.rilldata.com/build/credentials/) and [environmental variable](https://docs.rilldata.com/build/credentials/#variables) / [templating](https://docs.rilldata.com/deploy/templating) before getting started.

## Credentials

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

Sometimes, you will need to makes changes to the project locally and the credentials were not set up. In this case, if the project credentials are already configured, you can run the following to pull them locally. 

```bash
rill env pull
```

Conversly, if you've made any small changes to the .env file, you can push the changes using:

```bash
rill env push
```



## Environmental Variables
Similar to credentials, Environmental variables are usually set on Rill Developer and pushed to Rill Cloud when ready for deployment.

### Configuring Environmental Variables
There are a few ways to use environmental variables locally.

1. [Defining key-value pairs](https://docs.rilldata.com/reference/project-files/rill-yaml#setting-variables) in the `rill.yaml`.
2. If already deployed to Rill Cloud, directly modifying the .env file or using `rill env set`.
3. via the rill start command
```bash
rill start --var <var_name>=<value>
```

### Pushing to Rill Cloud

Depending on the method used during development, pushing to Rill Cloud can be as easy as:

1. Updating the Rill Deployment (if you added the variables to `rill.yaml`)
2. Pushing the .env file after making changes by running `rill env push`.
3. Setting the rill env variables by running `rill env set <key> <value>`

> Each number corresponds to how you configured the environmental variable above.


## Managing Credentials and Variables on Rill Cloud 
<ComingSoon />

<div class='contents_to_overlay'>
Historically (pre 0.48), management was only possible via the CLI. Now, it is also possible to do so via the UI! 

</div>

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />