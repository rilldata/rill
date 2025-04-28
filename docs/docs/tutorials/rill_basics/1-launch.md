---
title: "1. Source to Dashboard on Rill Cloud in 6 Steps"
sidebar_label: "1. Launch Rill Developer"
position: 1
collapsed: false
sidebar_position: 1
tags:
  - OLAP:DuckDB
---
:::note prerequisites
You will need to [install Rill](https://docs.rilldata.com/home/install).
:::

The goal of this six part tutorial is to get started with Rill and deploy your project to Rill Cloud. Upon deployment, your [30 day 
trial will start](./launch). The basics course prepares you for the [Advanced Features Course](../rill_developer_advanced_features/overview.md) that builds on top of this project. Let's get started.

## Start Rill Developer

```yaml
rill start my-rill-tutorial
```

:::tip
While we support macOS and Linux, you can also get Rill Developer running on [Windows machine via WSL](https://docs.rilldata.com/home/install#rill-on-windows-using-wsl). If you are having any issues installing and/or starting Rill, please see our [installation page](https://docs.rilldata.com/home/install). 

:::



After running the command, Rill Developer should automatically open in your default browser. If not, you can access it via the following URL:

```
localhost:9009
``` 
<img src = '/img/tutorials/101/new-rill-project.png' class='rounded-gif' />
<br />

Let's go ahead and select `Start with an empty project`.

<details>
  <summary>Where am I in the terminal?</summary>
  
    You can use the `pwd` command to see which directory in the terminal you are. <br />
    If this is not where you'd like to make the directory use the `cd` command to change directories.

</details>


:::note What is Rill Developer? 
Rill Developer is used to develop your Rill project as editing in Rill Cloud is not yet available. In Rill Developer, you will create connections to your source files, do some last mile ETL, define metrics in the metrics layer and finally create a dashboard. For more details on the differences between Rill Developer and Rill Cloud, see our documention, [here](/concepts/developerVsCloud.md).
:::