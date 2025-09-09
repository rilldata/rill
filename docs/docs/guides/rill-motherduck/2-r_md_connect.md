---
title: "2. Connect to MotherDuck"
sidebar_label: "2. Connect to MotherDuck"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:MotherDuck
  - Tutorial
---

## Default OLAP connection and Connect to MotherDuck

Within Rill, you can set the default OLAP connection on the [project level](https://docs.rilldata.com/reference/project-files/rill-yaml) or the [dashboard level](https://docs.rilldata.com/reference/project-files/explore-dashboards). 
For this tutorial, we will set it up on the project level so all of our dashboards will be based on our MotherDuck table.

### Connect to MotherDuck
We can create the MotherDuck connection by using the UI to add a MotherDuck live connector.




```yaml
type: connector
driver: motherduck

path: "md:my_db"
token: '{{ .env.connector.motherduck.token }}'
schema_name: "main" 
```

<!-- <img src = '/img/guides/md/MotherDuck-connector.png' class='rounded-gif' />
<br /> -->


### Rill Project Default

By default, Rill explicitly sets the project OLAP_connector to DuckDB. When creating a new live connector, the `olap_connector` key will automatically get updated.

```yaml
compiler: rillv1

title: "Rill and MotherDuck Project"
olap_connector: motherduck # default set to duckdb, only change if you modified the filename
```

For more information, take a look at our [MotherDuck documentation](/reference/project-files/connectors#motherduck).

### Securing Your MotherDuck Token

We do not recommend plain text passwords in the connector file, as you likely noticed in the sample YAML and screenshot. To use the sample, you'll need to create a `.env` in the root directory and add the following:


```
connector.motherduck.token="eyJhb....SAMPLETOKEN"
```

## Confirm Connection to MotherDuck

You'll see a change in the bottom left connector panel that gives you a look into your MotherDuck tables. 

<img src = '/img/guides/md/MotherDuck-confirm.png' class='rounded-gif' />
<br />

:::tip Using a different schema?

You can set the schema manually by setting the following parameter.
```yaml
schema_name: not_main
```
:::