---
title: "5. Deploy to Rill Cloud"
sidebar_label: '5. Deploy to Rill Cloud'
sidebar_position: 9
hide_table_of_contents: false
tags:
  - OLAP:ClickHouse
  - Tutorial
---
:::tip Rill Cloud Trial

If this is the first time you have deployed a project onto Rill Cloud, you will automatically start your [Rill Cloud Trial] () upon deployment of your Rill project. Your trial will last for 30 days. Please refer [here] () for more information on the details of your trial.

:::

## Deploy via the UI!

Select the `Deploy to share` button in the top right corner of a dashboard.

<img src = '/img/tutorials/rill-basics/deploy-ui.gif' class='rounded-gif' />
<br />

Steps to deploy to Rill Cloud:
1. Select the `Deploy to share` button.
2. Select `continue` on the free trial [link to article of free trial explanation]
    - If you have multiple organizations, please select Rill_Learn and `continue`.
3. Select `continue` on user invites.
4. You will be navigated to the /status page of your deployed project.


Take note of the following features in the UI:
<img src = '/img/tutorials/rill-basics/ui-explained.gif' class='rounded-gif' />

## In case of the following error:

```bash
connection: dial tcp 127.0.0.1:9000: connect: connection refused
```

This is likely due to using a locally running ClickHouse server. If so, you will not be able to access your locally running server from Rill Cloud. Instead, we suggest using [ClickHouse Cloud](https://clickhouse.com/cloud). 

For steps to setup ClickHouse Cloud, please refer to [our documentation](https://docs.rilldata.com/build/connectors/olap/clickhouse#connecting-to-clickhouse-cloud).
