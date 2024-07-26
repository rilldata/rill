---
title: "Changing the OLAP Engine"
sidebar_label: "Changing the OLAP Engine"
sidebar_position: 2
hide_table_of_contents: false
---

## Let's use Clickhouse!

In order to set the default OLAP connector with Rill, you need to add a [connector] () key pair into the rill.yaml file.


```yaml
connector: clickhouse
```