---
title: Visualize your MotherDuck Tables in Rill!
description: tutorial for MotherDuck
sidebar_position: 3
tags:
    - Tutorial
    - OLAP:MotherDuck
---

<img src = '/img/tutorials/ch/MotherDuck-rill.png' class='rounded-gif' />
<br />




In this tutorial, you'll learn how to use MotherDuck as your OLAP database with Rill to create powerful visualizations from your MotherDuck tables. Since both Rill and MotherDuck are built on DuckDB, they work seamlessly together, providing you with a smooth and efficient analytics experience.



The overall steps of the tutorial are:
1. Launch Rill Developer
2. Connect to your MotherDuck server via a Rill live connector.
3. Create a metrics view based on a table in MotherDuck
4. Create an explore dashboard via Generative AI
5. Deploy the dashboard to Rill Cloud, and share to colleagues

Once completed, you will have a basic understanding of how to use Rill Developer with MotherDuck and can customize the Rill Developer project to create further metrics views based on MotherDuck tables and create more dashboards. 

### Let's get started!

:::tip Direct Ingestion into MotherDuck via Rill

While we won't go over it directly in this tutorial, you can use Rill as an ingestor into MotherDuck. See the list of our [supported connectors](/build/connect) for more information.

:::

