---
title: Getting Started with Rill Developer 
sidebar_label: Start Guide
sidebar_position: 01
---
import Video from '@site/src/components/Video';


<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->


:::tip Rill's Default Engine
The below guide contents assume that you will be using our 'default' embedded engine. If you're looking for a guide with setting up Rill with ClickHouse, see our [ClickHouse Guide](/guides/rill-clickhouse/)!
:::
<!-- <div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/7TlO6E5gZzY?autoplay=1&mute=1&rel=0&si=CMltjZI4S5oAAAtg"
    frameBorder="0"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
    allowFullScreen
    style={{
      position: "absolute",
      top: 0,
      left: 0,
      width: "100%",
      height: "100%",
      borderRadius: "10px", // Apply to iframe as well for rounded effect
    }}
  ></iframe>
</div>
<br/> -->

Rill Developer is your all-in-one tool for exploring data and building insightful dashboards with ease. Whether you're ingesting new datasets, prototyping dashboards, or analyzing schema relationships, Rill streamlines the process. Use Rill to quickly visualize your data, iterate on dashboard designs, and uncover actionable insights. When you're ready, publish your dashboards to the Cloud and collaborate with your team. 

## Install and start Rill

If you haven't already, install and start Rill with the simple command below!
```bash
curl https://rill.sh | sh
rill start my-rill-project
```


<img src = '/img/tutorials/rill-basics/new-rill-project.png' class='rounded-gif' />
<br />

:::note What is Rill Developer? 
Rill Developer is used to develop your Rill project, as editing in Rill Cloud is not yet available. In Rill Developer, you will create connections to your source files, perform last-mile ETL, define metrics in the metrics layer, and finally create a dashboard. For more details on the differences between Rill Developer and Rill Cloud, see our documentation [here](/concepts/developerVsCloud.md).
:::


### Importing Data and Schema Information

Rill supports many different sources for ingestion into our embedded analytical engine. See our list of available [connectors](/reference/connectors/). The below is using a publically available dataset hosted on our GCS bucket. 

Note that after the dataset is ingested, you can see a sample of the data (first 150 rows) and more information about the column schema, information about the data (types, ranges of values, etc)

<img src = '/img/tutorials/rill-basics/Adding-Data.gif' class='rounded-gif' />
<br />


### Create a Dashboard with AI



## Key Take Aways

Skipped a few steps, metrics-view, last mile ETL modeling, but also showed how quick to go from data to dashboard. For more in depth steps and instructions, see our in-depth guide.