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

Within Rill you can set the default OLAP connection on the [project level](https://docs.rilldata.com/reference/project-files/rill-yaml) or the [dashboard level](https://docs.rilldata.com/reference/project-files/explore-dashboards). 
For this course, we will set it up on the project level so all of our dashboards will be based on our MotherDuck table.

### Connect to MotherDuck
We can create the MotherDuck connection by modifying the connectors/duckdb.yaml.


```yaml
type: connector
driver: duckdb

path: "md:my_db"

init_sql: |
  INSTALL 'motherduck';
  LOAD 'motherduck';
  SET motherduck_token= '{{.env.connector.motherduck.token}}'
```

<img src = '/img/tutorials/md/MotherDuck-connector.png' class='rounded-gif' />
<br />

:::tip change the connector name
While not necessary, I personally like to change the name to motherduck.yaml to make it clear which connector to use.

:::
### Rill Project Default

By default, we explicitly set the project OLAP_connector to duckdb. This is based on the connector.yaml name. If you make any changes to the filename, don't forget to change the name in your rill.yaml file.

```yaml
compiler: rillv1

title: "Rill and MotherDuck Project"
olap_connector: motherduck #default set to duckdb, only change if you modified the filename
```

For more information, take a look at our [MotherDuck documentation](/reference/project-files/connectors#motherduck).

### Securing Your MotherDuck Token

We do not recommend plain text passwords in the connector file, as you likely noticed in the sample YAML, and screenshot. To use the sample, you'll need to create a `.env` in the root directory and add the following:


```
connector.motherduck.token="eyJhb....SAMPLETOKEN"
```

## Confirm Connection to MotherDuck

You'll see a change in the bottom left connector that gives you a look into your MotherDuck tables. 

<img src = '/img/tutorials/md/MotherDuck-confirm.png' class='rounded-gif' />
<br />