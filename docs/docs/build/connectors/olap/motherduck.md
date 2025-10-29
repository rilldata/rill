---
title: MotherDuck
description: Power Rill dashboards using MotherDuck
sidebar_label: MotherDuck
sidebar_position: 15
---

[MotherDuck](https://motherduck.com/) is a cloud-native DuckDB service that provides scalable analytics and data processing capabilities. Built on the same core engine as DuckDB, MotherDuck offers the familiar SQL interface and performance characteristics while adding cloud-native features like serverless compute, automatic scaling, and collaborative data sharing.


<img src='/img/build/connectors/olap-engines/motherduck/rill-developer.png' class='rounded-gif' />
<br />



Rill supports connecting to MotherDuck and using it as an OLAP engine to power dashboards. This is particularly useful when you want the performance and SQL compatibility of DuckDB with the scalability and collaboration features of a cloud service.

:::note Supported Versions
Rill supports connecting to MotherDuck using the latest DuckDB-compatible drivers and protocols.
:::

## Getting Your MotherDuck Access Token

To connect to MotherDuck, you'll need a access token from your MotherDuck account:

1. Log in to your [MotherDuck account](https://motherduck.com/)
2. Navigate to the **Settings** section
3. Go to **Access Tokens**
4. Create a new access token or copy an existing one
   
<img src='/img/build/connectors/olap-engines/motherduck/service-token.png' class='rounded-gif' />
<br />

:::warning Keep Your Token Secure

Your MotherDuck access token provides access to your data. Keep it secure and never commit it directly to version control. Always use environment variables or secure credential management.

:::


## Configuring Rill Developer with MotherDuck

Connect to your OLAP engine via Add Data. This will automatically create the motherduck.yaml file in your connectors folder and populate the .env file with `.connector.motherduck.token`.

For more information on supported parameters, see our [MotherDuck connector YAML reference docs](/reference/project-files/connectors#motherduck).


```yaml
type: connector 
driver: duckdb 

token: '{{ .env.connector.motherduck.token }}' 
path: "md:my_database" 
schema_name: "my_schema"
```

### Setting the Default OLAP Connection

Creating a connection to MotherDuck will automatically add the `olap_connector` property in your project's [rill.yaml](/reference/project-files/rill-yaml) and change the default OLAP engine to `motherduck`.

```yaml
olap_connector: motherduck
```

:::info Interested in using multiple OLAP engines in the same project?

Please see our [Using Multiple OLAP Engines](/build/connectors/olap/multiple-olap) page.
:::


## Configuring Rill Cloud

When deploying a MotherDuck-backed project to Rill Cloud, you have the following options to pass the appropriate access token:

1. If you have followed the UI to create your MotherDuck connector, the token should already exist in the `.env` file. During the deployment process, this `.env` file is automatically pushed with the deployment.

2. If `motherduck_token` has already been set in your project `.env`, you can push and update these variables directly in your cloud deployment by using the `rill env push` command.

3. If you manually passed the connector when running `rill start`, you will need to use the `rill env configure` command to set `motherduck_token` onto Rill Cloud as well.

:::info
Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.
:::



## Additional Notes

- MotherDuck uses the same SQL syntax as DuckDB, so all standard DuckDB functions and features are available
- For dashboards powered by MotherDuck, [measure definitions](/build/metrics-view/#measures) should follow standard [DuckDB SQL](https://duckdb.org/docs/sql/introduction) syntax

:::info Need help connecting to MotherDuck?

If you would like to connect Rill to MotherDuck or need assistance with setup, please don't hesitate to [contact us](/contact). We'd love to help!

:::