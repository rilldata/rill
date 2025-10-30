---
title: "Snowflake"
description: Connect to a Snowflake table as a data source
sidebar_label: "Snowflake"
sidebar_position: 50
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

:::warning 
Snowflake has issued a [deprecation notice](https://www.snowflake.com/en/blog/blacklisting-eol-io-domains/) effective for versions 1.23.0 onward. This means that you **must** use your own Snowflake account (`<account>.snowflakecomputing.com`) and _cannot_ use the IO domain endpoints (e.g. `<account>.us-west-2.snowflakecomputing.io`) after February 28, 2025. Please refer to our [migration guide](/deploy/existing-project/migrate-project) for more details.
:::

[Snowflake](https://docs.snowflake.com/en/user-guide-intro) is a cloud-based data platform designed to facilitate data warehousing, data lakes, data engineering, data science, data application development, and data sharing. It separates compute and storage, enabling users to scale up or down instantly without downtime, providing a cost-effective solution for data management. With its unique architecture and support for multi-cloud environments, including AWS, Azure, and Google Cloud Platform, Snowflake offers seamless data integration, secure data sharing across organizations, and real-time access to data insights, making it a common choice to power many business intelligence applications and use cases. You can connect to and read from Snowflake data warehouses using the [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake).

## Local Modeling

Create a connector YAML file in your Rill project directory (e.g. `connectors/snowflake.yaml`) with your Snowflake connection details and authentication credentials. Replace the placeholder text in this example with your actual Snowflake credentials:

```yaml
type: connector
driver: snowflake
account: "{{ .env.connector.snowflake.account }}"
username: "{{ .env.connector.snowflake.username }}"
password: "{{ .env.connector.snowflake.password }}"
database: "{{ .env.connector.snowflake.database }}"
schema: "{{ .env.connector.snowflake.schema }}"
warehouse: "{{ .env.connector.snowflake.warehouse }}"
role: "{{ .env.connector.snowflake.role }}"
```

Set these Snowflake credentials in your `.env` file:

```bash
connector.snowflake.account=<your_account>.snowflakecomputing.com
connector.snowflake.username=<your_username>
connector.snowflake.password=<your_password>
connector.snowflake.database=<your_database>
connector.snowflake.schema=<your_schema>
connector.snowflake.warehouse=<your_warehouse>
connector.snowflake.role=<your_role>
```

:::info
The `account` should be your full Snowflake account URL (e.g., `mycompany.us-east-1.snowflakecomputing.com`). Do not use the deprecated `.io` domain endpoints.
:::

Create a model YAML file (e.g. `models/my_snowflake_data.yaml`) that references your Snowflake connector to pull data from your Snowflake table or view:

```yaml
type: model
connector: snowflake
sql: SELECT * FROM my_table
```

:::tip
For large tables, consider adding LIMIT clauses during development or using WHERE conditions to reduce query costs and data transfer time.
:::

## Deployment

Once you're ready to deploy your project to Rill Cloud, you can set the credentials as secrets using the `rill env configure` command:

```bash
rill env configure
```

The CLI will walk you through configuring each connector used in your project. For Snowflake, you'll be prompted to provide:
- Your Snowflake account URL
- Username and password
- Database, schema, warehouse, and role

After configuring your credentials, deploy your project:

```bash
rill deploy
```

## Additional Resources

- [Snowflake Documentation](https://docs.snowflake.com/)
- [Go Snowflake Driver](https://pkg.go.dev/github.com/snowflakedb/gosnowflake)
- [Snowflake Account Migration Guide](https://www.snowflake.com/en/blog/blacklisting-eol-io-domains/)
