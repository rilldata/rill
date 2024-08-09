---
title: Rill Cloud vs Rill Developer 
sidebar_label: Rill Cloud vs Rill Developer 
sidebar_position: 12
hide_table_of_contents: true
---

## What is Rill Developer and Rill Cloud?

Rill offers two unique experiences within our product, **Rill Developer** and **Rill Cloud**.

As the name suggestions Rill Developer is designed around the Developer. Users on Rill Developer will import, wrangle, and explore the data before presenting it to the team. Rill Developer runs locally and when you're ready you can push your changes onto Rill Cloud.

Rill Cloud is designed for our dashboard consumers. Once the developer has deployed the dashboard onto Rill Cloud, these users will be able to utilize the dashboards in their everyday tasks for fast, business-level speed.





## Rill Developer

Rill Developer is designed around developers. Using a familiar user interface, developers are able to import data, create SQL models, and create metric-views. Many of the underlying files in Rill Developer are either SQL or YAML files. Once a data in imported into Rill and the underlying OLAP engine, you will be able to make any last mile ETL changes using a SQL model file. You can then create and materizlize your ["One Big Table"](../build/models/models.md) for your dashboard needs. Finally, any specifications for your dimensions and measures can be defined and tested in Developer's dashboard preview.

<img src = '/img/concepts/rcvsrd/empty-project.png' class='rounded-gif' />
<br />

<details> 
    <summary> What are some things you can do in Rill Developer?</summary>

    From source ingestion, modeling, to dashboard. 
| UI  : <img src = '/img/concepts/rcvsrd/DevUI.gif' class='rounded-gif' /> | Adding Sources:  <img src = '/img/concepts/rcvsrd/Adding-Data.gif' class='rounded-gif' /> |
|:---:|:---:|
| **Creating Models:** <img src = '/img/concepts/rcvsrd/Add-Model.gif' class='rounded-gif' />  | **Creating Dashboards:** <img src = '/img/concepts/rcvsrd/Add-Dashboard.gif' class='rounded-gif' />  |
</details>


## Rill Cloud

Once the dashboard has been [deployed to Rill Cloud](../deploy/existing-project/existing-project.md), the dashboard can be viewed by your organization's users. As you can see below, the UI is different from Developer. Upon accessing Rill Cloud, a user will be able to view all the projects they have view access to. 


<img src = '/img/concepts/rcvsrd/rill-cloud-projects.png' class='rounded-gif' />
<br />

 After selecting a specific project, they will be direct to a list of dashboards. The dashboard consumer will not have access to make any modifications to the dataset or sources. Instead, they are given a few different features such as, alerts, status, Shareable Public URLs, see[ Explore section](../explore/dashboard-101.md) for more features.

<img src = '/img/concepts/rcvsrd/Rill-cloud.png' class='rounded-gif' />

## Is Rill Cloud a higher offering than Rill Developer?

Based on the naming, it might be confusing and easy to assume that Rill Cloud is our "higher" offering but **that is not the case!** 


Rill Developer and Rill Cloud are to be used together. Rill Developer gives a space for our developers to define and test any new or needed changes to the data and/or dashboards before pushing to our Rill Cloud users who need stable access to working dashboards. 

