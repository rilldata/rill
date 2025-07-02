---
title: "Rill Developer vs Rill Cloud"
sidebar_label: "Why Deploy to Rill Cloud?"
position: 1
collapsed: false
sidebar_position: 1
tags:
  - OLAP:DuckDB
  - Rill Developer
  - Getting Started
---

For a full documentation on what the difference between Rill Cloud and Rill Developer is, click [here](/concepts/developerVsCloud). 

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/zW1Xms2qQlc?si=OpKVKN7csHCY_AcX"
    frameBorder="0"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
    allowFullScreen
    style={{
      position: "absolute",
      top: 0,
      left: 0,
      width: "100%",
      height: "100%",
      borderRadius: "10px"
    }}
  ></iframe>
</div>


## Why deploy to Rill Cloud?

Rill Developer is an extremely strong tool to deep dive into your data as it allows users to import sources from many destinations and join these tables together to create something useful in a slice-and-dice visualization. Many times the feedback we receive is, "`I didn't even know my data had an issue in it,`" or, "`In a few minutes, I was able to make new insights into my data that would've taken me hours.`" This is great, but if that's possible in Rill Developer, why pubish the dashboard into Rill Cloud? 



## Rill Cloud Exclusive Features



While we dont have Rill Developer in the Cloud (yet), Rill Cloud offers another set of exclusive features that allow you to share your findings with your colleagues, and allows for more collaboration between teams. How so? 

### - **Scheduled Reports** - 

<img src = '/img/tutorials/rill-advanced/scheduled-report.png' class='rounded-gif' />
<br />
Schedule a report that sends at 9 AM on Mondays to ensure that your team is up to date with the latest week over week information on specific metrics. This email contains both a link back the dashboard (a quick way to dig into the data), or the ability to download the CSV to import into whatever application you need.

### - **Alerts** -

<img src = '/img/tutorials/rill-advanced/alert.png' class='rounded-gif' />
<br />
Put in alerts that go off when your specific measure goes over or below a specific value or threshold. 10% gain of revenue from last week, get an email or slack alert about it! Hit a goal of 30 unique user activity in your logs, get an email or slack alert about it! 

### - **Public URL / Sharing**-

<img src = '/img/tutorials/rill-advanced/public-url.png' class='rounded-gif' />
<br />
Need to provide quick access to a dashboard to an external user OR a user that's not part of your Rill organization yet? Use a public expirable URL that allows quick access to Rill's platform and make faster decisions.

### - **Bookmarks / User Last States**-
<img src = '/img/tutorials/rill-advanced/bookmarks.png' class='rounded-gif' />
<br />
Create easy to navigate views via Bookmarks, create a "default" home view, but also give users the ability to navigate to their last saved view. 

### - **More!**-
We're constantly coming out with new features, so keep up to date with our [release notes](/notes)!