---
title: Rill Cloud vs Rill Developer 
sidebar_label: Rill Cloud vs Rill Developer 
sidebar_position: 13
hide_table_of_contents: true
---

## What is Rill Developer and Rill Cloud?

Rill offers two unique but complementary experiences within our broader product suite, **Rill Developer** and **Rill Cloud**.

As the name suggests, Rill _Developer_ is designed with the developer in mind where the project development will actually occur. Rill Developer is meant for the primary developers of project assets and dashboards, where they can import, wrangle, iterate on, and explore the data before presenting it for broader consumption by the team. Rill Developer is meant to run on your local machine - see here for some [recommendations and best practices](/deploy/performance#local-development--rill-developer) - but it is a simple process to [deploy a project](/deploy/deploy-dashboard/) once ready to Rill Cloud.


Rill Cloud on the other hand is designed for our dashboard consumers and allows broader team members to easily collaborate. Once the developer has deployed the dashboard onto Rill Cloud, these users will be able to utilize the dashboards to interact with their data, set alerts / bookmarks, investigate nuances / anomalies, or otherwise perform everyday tasks for their business needs at Rill speed.

<img src = '/img/concepts/rcvsrd/DevCloudComparison.png' class='rounded-gif' />
<br />

:::info Rill Developer vs Rill Cloud

Please note that a common **misnomer** is that Rill Developer can be a sufficient replacement for Rill Cloud. They both serve different purposes but are meant to be used _in conjunction_. Rill enables speed of exploration and is easy to use for developers, allowing the project to be iterated on quickly. Rill Cloud then allows for shared collaboration at scale, especially for production deployments.

:::

## Rill Developer

Rill Developer is designed around developers. Using a familiar IDE-like interface, developers are able to import data, create SQL models, and create metric-views. Many of the underlying files in Rill Developer are either SQL or YAML files. Once a data in imported into Rill (and the underlying OLAP engine), developers are able to perform last mile ETL changes using one or a series of SQL models (as its own [DAG](https://en.wikipedia.org/wiki/Directed_acyclic_graph#:~:text=A%20directed%20acyclic%20graph%20is,a%20path%20with%20zero%20edges)). You can then create and materialize your ["One Big Table"](../build/models/models.md) for your dashboard needs. Finally, any specifications for your dimensions and measures can be defined and tested in Developer's dashboard preview.

<img src = '/img/concepts/rcvsrd/empty-project.png' class='rounded-gif' />
<br />

<details> 
    <summary> What are some things you can do in Rill Developer?</summary>

    Anything from source ingestion to modeling to creating dashboards. 
|         UI  : <img src = '/img/concepts/rcvsrd/DevUI.gif' class='rounded-gif' />          |      Add Sources:  <img src = '/img/concepts/rcvsrd/Adding-Data.gif' class='rounded-gif' />       |
| :---------------------------------------------------------------------------------------: | :-----------------------------------------------------------------------------------------------: |
| **Create Models:** <img src = '/img/concepts/rcvsrd/Add-Model.gif' class='rounded-gif' /> | **Create Dashboards:** <img src = '/img/concepts/rcvsrd/Add-Dashboard.gif' class='rounded-gif' /> |
</details>


## Rill Cloud

Once the dashboard has been [deployed to Rill Cloud](../deploy/deploy-dashboard/), the dashboard can be shared with others and viewed by other members of your Rill Cloud organization. As you can see below, the UI is different from Developer. Upon accessing Rill Cloud, a user will be able to view all the projects they have been granted access to by project admins. 


<img src = '/img/concepts/rcvsrd/rill-cloud-projects.png' class='rounded-gif' />
<br />

 After selecting a specific project, they will be directed to a list of dashboards. From Rill Cloud, the dashboard consumer does not have the ability to make any modifications to sources or models. However, they are given some additional capabilities that are not accessible in Rill Developer, such as alerting, creating bookmarks or sharable public URLs, checking the project status, and more.

 :::info Dashboard 101

 For more details about using a Rill Cloud dashboard, please refer to our [Explore section](/explore/dashboard-101/)!

 :::
 

<img src = '/img/concepts/rcvsrd/Rill-cloud.png' class='rounded-gif' />
<details> 
    <summary> What are some things you can do in Rill Cloud?</summary>

    Anything from source ingestion to modeling to creating dashboards. 
|       Alerts: <img src = '/img/concepts/rcvsrd/alert.gif' class='rounded-gif' />        |          Bookmarks:  <img src = '/img/concepts/rcvsrd/bookmark.gif' class='rounded-gif' />          |
| :-------------------------------------------------------------------------------------: | :-------------------------------------------------------------------------------------------------: |
| **Public URL:** <img src = '/img/concepts/rcvsrd/public-url.gif' class='rounded-gif' /> | **Scheduled Report:** <img src = '/img/concepts/rcvsrd/scheduled-report.gif' class='rounded-gif' /> |
</details>

## Is Rill Cloud a higher offering than Rill Developer?

Based on the naming, it might be confusing and easy to assume that Rill Cloud is our "higher" offering but **that is not the case!** Similarly, Rill Developer is _not meant to be used as a standalone tool either_.


Rill Developer and Rill Cloud are to be used together. Rill Developer gives a space for our developers to define and test any new or needed changes to the data and/or dashboards before pushing to our Rill Cloud users who need stable access to working dashboards. Then, once finalized, these dashboards are deployed to the Rill Cloud project for broader consumption and to power business use cases.

