---
title: "2. Connect to ClickHouse"
sidebar_label: "2. Connect to ClickHouse"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:ClickHouse
  - Tutorial
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


## Default OLAP connection and Connect to ClickHouse

Within Rill you can set the default OLAP connection on the [project level](https://docs.rilldata.com/reference/project-files/rill-yaml) or the [dashboard level](https://docs.rilldata.com/reference/project-files/explore-dashboards). 
For this course, we will set it up on the project level so all of our dashboards will be based on our ClickHouse table.

:::tip
You have two options for your ClickHouse server:
1. Use a [local running ClickHouse server](https://clickhouse.com/docs/en/install)
2. Use [ClickHouse Cloud](https://clickhouse.com/docs/en/cloud/overview)

Depending what you choose, the contents of your connection will change and I recommend looking through [our ClickHouse documentation](https://docs.rilldata.com/build/connectors/olap/clickhouse) for further information.

:::

### Connect to ClickHouse
We can create the clickhouse connection by selection `+Add Data` > `ClickHouse` and fill in the components on the UI.

<img src = '/img/tutorials/ch/clickhouse-connector.png' class='rounded-gif' />
<br />
:::tip
You can obtain the credentials from your ClickHouse Cloud account by clicking the `Connect` button in the left panel.:

<img src = '/img/tutorials/ch/clickhouse-cloud-credential.png' class='rounded-gif' />
<br />
```
"https://<hostname>:<port>?username=<username>&password=<password>&secure=true&skip_verify=true"
```

:::
Once this is created, a `clickhouse.yaml` file will appear in the `connectors` directory and the following will be added to your rill.yaml.

```yaml
compiler: rillv1

title: "Rill and ClickHouse Project"
olap_connector: clickhouse #automatically added
```

Example for a locally running ClickHouse server:
```yaml
host: "localhost"
port: "9000"
```
or 
```yaml
dsn: "clickhouse://localhost:9000"
```


 You can either add the credentials in plain text or dsn via the yaml file or add the credentials via the CLI.


Please see our documentation to find the DSN for [your ClickHouse Cloud instance](https://docs.rilldata.com/build/connectors/olap/clickhouse#connecting-to-clickhouse-cloud). 

### How to pass the credentials to Rill
There are a few way to define the credentials within Rill.

<Tabs>
<TabItem value="yaml" label="via yaml" default>
Please create a file called clickhouse.yaml and add the following contents.
```yaml
type: connector
driver: clickhouse

host: "localhost"
port: "9000"
```
or 
```yaml
type: connector
driver: clickhouse

dsn: "clickhouse://localhost:9000"
```



</TabItem>
<TabItem value="variable" label="via variables">
Navigate back to the Terminal and stop the Rill process. You can run the following to add a variable and use this is within Rill.

```
rill start --env host='localhost' --env  port='9000'
```

Afterwards, create a file called clickhouse.yaml and add the following contents:

```yaml
type: connector
driver: clickhouse

host: '{{ .env.host }}'
port: '{{ .env.port }}'
```



  </TabItem>


  <TabItem value="env" label="via .env">
There's a few way to generate the .env file. Making a source that requires credentials will automatically generate it. Else, you can create it using `touch .env` in the rill directory.

```yaml
connector.clickhouse.host="localhost"
connector.clickhouse.port=9000
connector.clickhouse.username 
connector.clickhouse.password 
connector.clickhouse.ssl 

or

connector.clickhouse.dsn="..."
```

  </TabItem>
</Tabs>

:::tip Via the UI

If you connect to ClickHouse via the UI, this will automatically create a template with connectors/clickhouse.yaml as well as a reference of your DSN to the .env folder. This will automatically get pushed along with your project to Rill Cloud. 

:::

You should now be able to see the contents of your ClickHouse database in the left panel of your UI.

<img src = '/img/tutorials/ch/olap-connector.png' class='rounded-gif' />
