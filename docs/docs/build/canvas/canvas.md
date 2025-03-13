---
title: Create Canvas Dashboards
description: Create classic dashboards using various metrics views
sidebar_label: Create Canvas Dashboards
sidebar_position: 05
---

> insert screenshot of openRTB canvas

For a more classic dashboard creation experience, Canvas dashboards can create a first look into your data at a top level. Similar to how Explore dashbords are built on top of a metrics view, Canvas does the same but also allows you to connect to various metrics view to display in a single view. 

:::tip
In order to enable this feature, which is in public beta, you will need to add the following to your `rill.yaml` file.

```
features:
   - canvasDashboards
```

:::

- **metrics view**: powers the different components within a single Canvas Dashboard
- **measures**: defined in your metrics view, you can select the measure to display 
- **dimensions**: defined in your metrics view, you can select the dimensions to display 
- **components**: a single object within a Canvas dashbord, including KPIs, charts, tables, and more! 


## A Visual Experience 
> insert image





## Making small changes to the YAML 
As always, while we allow our users to create and modify the dashboard via the UI, you can always change the view to a more traditional YAML text by selecting the toggle.

>insert image 
