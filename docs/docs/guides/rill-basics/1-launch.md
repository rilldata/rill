---
title: "1. Source to Dashboard on Rill Cloud in 6 Steps"
sidebar_label: "1. Launch Rill Developer"
position: 1
collapsed: false
sidebar_position: 1
tags:
  - Tutorial
  - OLAP:DuckDB
  - Rill Developer
  - Getting Started
---
:::note prerequisites
You need to [install Rill](https://docs.rilldata.com/get-started/install). 

```bash
curl https://rill.sh | sh
```

:::

The goal of this six-part tutorial is to get started with Rill and deploy your project to Rill Cloud. Upon deployment, your [30-day trial will start](/other/plans#trial-plan). Each course will build upon the previous one, allowing you to have a fully functioning project with many of our advanced features. This tutorial can be used in tandem with our documentation to ensure you have up-to-date information.


## Start Rill Developer

```yaml
rill start my-rill-tutorial
```

:::tip
While we support macOS and Linux, you can also get Rill Developer running on a [Windows machine via WSL](https://docs.rilldata.com/get-started/install#rill-on-windows-using-wsl). If you are having any issues installing or starting Rill, please see our [installation page](https://docs.rilldata.com/get-started/install). 

:::



If running Rill in a new directory, you'll be prompted with the following. Type "Y" and press Enter. 

```bash
? Rill will create project files in "~/Desktop/GitHub". Do you want to continue? (Y/n) 

```

Rill Developer will automatically open in your default browser. If not, you can access it via the following URL:

```
localhost:9009
``` 

Welcome to Rill Developer!

:::note What is Rill Developer? 
Rill Developer is used to develop your Rill project, as editing in Rill Cloud is not yet available. In Rill Developer, you will create connections to your source files, perform last-mile ETL, define metrics in the metrics layer, and finally create a dashboard. For more details on the differences between Rill Developer and Rill Cloud, see our documentation [here](/get-started/concepts/cloud-vs-developer)
:::

<img src = '/img/tutorials/rill-basics/new-rill-project.png' class='rounded-gif' />
<br />

Let's go ahead and select `Start with an empty project`. If you want to skip the basics, you can select one of the quick start projects and refer to our Quick Start Guide for the corresponding project. Note that we have many more projects available in our public repo [here](https://github.com/rilldata/rill-examples).

<details>
  <summary>Where am I in the terminal?</summary>
  
    You can use the `pwd` command to see which directory you are in within the terminal. <br />
    If this is not where you'd like to make the directory, use the `cd` command to change directories.

</details>


