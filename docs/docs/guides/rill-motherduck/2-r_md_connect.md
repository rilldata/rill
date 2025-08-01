---
title: "2. Connect to MotherDuck"
sidebar_label: "2. Connect to MotherDuck"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:MotherDuck
  - Tutorial
---
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


## Default OLAP connection and Connect to MotherDuck

Within Rill you can set the default OLAP connection on the [project level](https://docs.rilldata.com/reference/project-files/rill-yaml) or the [dashboard level](https://docs.rilldata.com/reference/project-files/explore-dashboards). 
For this course, we will set it up on the project level so all of our dashboards will be based on our MotherDuck table.

:::tip
You have two options for your MotherDuck server:
1. Use a [local running MotherDuck server](https://MotherDuck.com/docs/en/install)
2. Use [MotherDuck Cloud](https://MotherDuck.com/docs/en/cloud/overview)

Depending what you choose, the contents of your connection will change and I recommend looking through [our MotherDuck documentation](https://docs.rilldata.com/reference/olap-engines/MotherDuck) for further information.

:::

### Connect to MotherDuck
We can create the MotherDuck connection by selection `+Add Data` > `MotherDuck` and fill in the components on the UI.

<img src = '/img/tutorials/ch/MotherDuck-connector.png' class='rounded-gif' />
<br />
:::tip
You can obtain the credentials from your MotherDuck Cloud account by clicking the `Connect` button in the left panel.:

<img src = '/img/tutorials/ch/MotherDuck-cloud-credential.png' class='rounded-gif' />
<br />
```
"https://<hostname>:<port>?username=<username>&password=<password>&secure=true&skip_verify=true"
```

:::
Once this is created, a `MotherDuck.yaml` file will appear in the `connectors` folder and the following will be added to your rill.yaml.

```yaml
compiler: rillv1

title: "Rill and MotherDuck Project"
olap_connector: MotherDuck #automatically added
```

Example for a locally running MotherDuck server:
```yaml
host: "localhost"
port: "9000"
```
or 
```yaml
dsn: "MotherDuck://localhost:9000"
```


 You can either add the credentials in plain text or dsn via the yaml file or add the credentials via the CLI.


Please see our documentation to find the DSN for [your MotherDuck Cloud instance](https://docs.rilldata.com/reference/olap-engines/MotherDuck#connecting-to-MotherDuck-cloud). 

### How to pass the credentials to Rill
There are a few way to define the credentials within Rill.

<Tabs>
<TabItem value="yaml" label="via yaml" default>
Please create a file called MotherDuck.yaml and add the following contents.
```yaml
type: connector
driver: MotherDuck

host: "localhost"
port: "9000"
```
or 
```yaml
type: connector
driver: MotherDuck

dsn: "MotherDuck://localhost:9000"
```



</TabItem>
<TabItem value="variable" label="via variables">
Navigate back to the Terminal and stop the Rill process. You can run the following to add a variable and use this is within Rill.

```
rill start --env host='localhost' --env  port='9000'
```

Afterwards, create a file called MotherDuck.yaml and add the following contents:

```yaml
type: connector
driver: MotherDuck

host: '{{ .env.host }}'
port: '{{ .env.port }}'
```



  </TabItem>


  <TabItem value="env" label="via .env">
There's a few way to generate the .env file. Making a source that requires credentials will automatically generate it. Else, you can create it using `touch .env` in the rill directory.

```yaml
connector.MotherDuck.host="localhost"
connector.MotherDuck.port=9000
connector.MotherDuck.username 
connector.MotherDuck.password 
connector.MotherDuck.ssl 

or

connector.MotherDuck.dsn="..."
```

  </TabItem>
</Tabs>

:::tip Via the UI

If you connect to MotherDuck via the UI, this will automatically create a template with connectors/MotherDuck.yaml as well as a reference of your DSN to the .env folder. This will automatically get pushed along with your project to Rill Cloud. 

:::

You should now be able to see the contents of your MotherDuck database in the left panel of your UI.

<img src = '/img/tutorials/ch/olap-connector.png' class='rounded-gif' />
