---
title: Connector YAML
sidebar_label: Connector YAML 
sidebar_position: 70
hide_table_of_contents: true
---


When you add olap_connector to your rill.yaml file, you will need to set up a `<connector_name>.yaml` file in the 'connectors' directory. This file requires the following parameters,`type` and `driver` (see below for more parameter options). Rill will automatically test the connectivity to the OLAP engine upon saving the file. This can be viewed in the connectors tab in the UI.

:::tip Did you know?

Starting from Rill 0.46, you can directly create OLAP engines from the UI! 
Select + Add -> Data -> Connect an OLAP engine

:::


## Properties

**`type`** - refers to the resource type and must be 'connector'

**`driver`** - refers to the [OLAP engine](../olap-engines/multiple-olap.md)
- _`clickhouse`_ link to[ Clickhouse documentation](https://clickhouse.com/docs/en/intro)
- _`druid`_ link to[ Druid documentation](https://druid.apache.org/docs/latest/design/)
- _`pinot`_ link to[ Pinot documentation](https://docs.pinot.apache.org/)

:::tip A note on OLAP engines

By defining the `connector` parameter in a [dashboard's YAML](explore-dashboards.md) file, you can have multiple OLAP engines in a single project.

:::

**`host`** - refers to the hostname

**`port`** - refers to the port 

**`username`** - the username, in plaintext

**`password`** - the password, in plaintext

**`ssl`** - depending on the engine, this parameter may be required (_pinot_)


You can also connect using a dsn parameter. You cannot use the above parameters along with the **`dsn`** parameter.

**`dsn`** - connection string containing all the details above, in a single string. Note that each engine's syntax is slightly different. Please refer to [our documentation](https://docs.rilldata.com/reference/olap-engines/) for further details.

---
 
_Example #1: Connecting to a local running Clickhouse server (no security enabled)_
```yaml
type: connector
driver: clickhouse

host: "localhost"
port: "9000"
```

_Example #2: Connecting to a ClickHouse Cloud_
```yaml
type: connector
driver: clickhouse


dsn: "https://<hostname>:<port>?username=<username>&password=<password>&secure=true&skip_verify=true"

```
