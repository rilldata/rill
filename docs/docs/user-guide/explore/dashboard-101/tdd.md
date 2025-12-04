---
title: "Time Dimension Detail"
description: Time Dimension Detail
sidebar_label: "Time Dimension Detail"
sidebar_position: 25
---

The Time Dimension Detail is a separate visualization accessed by clicking on any time series chart in Explore. Expanding the time series allows for more surface area to compare a larger number of dimensions trended over time. Additionally, you can create rapid pivot tables based on time across any dimension.

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/KnMKjahnjeU?si=bnXO82HXIRsnd3Xv"
    frameBorder="0"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
    allowFullScreen
    style={{
      position: "absolute",
      top: 0,
      left: 0,
      width: "100%",
      height: "100%",
      borderRadius: "10px", 
    }}
  ></iframe>
</div>
<br/>

Prefer video? Check out our [YouTube playlist](https://www.youtube.com/watch?v=wTP46eOzoCk&list=PL_ZoDsg2yFKgi7ud_fOOD33AH8ONWQS7I&index=1) for a quick start!


## Time Dimension Detail (TDD)

Within the TTD screen, you can apply all of the same filters and comparisons as Explore. Any filters applied will be carried into the TDD or will be carried out to Explore if you return to the main page on the top left. 

Underneath the expanded time series chart, you will see two sets of dropdowns - one for Rows to change the comparison (this defaults to Time but can be changed to any dimension) and one for Columns (which can be used to cycle through all of your metrics). On the top right, you can also export the TDD pivot view for quick time series reporting. 

Similar to the filters, the Search and Exclude options from the expanded Leadersboards are available on the top right of the TDD table. Lastly, you can also take the TDD table directly to the Pivot view for multi-dimensional analysis - more details on [Pivot here](/user-guide/explore/dashboard-101/pivot).

    
<img src = '/img/explore/tdd/tdd.gif' class='rounded-gif' />
<br />
 
:::tip Adjusting Time Grains
The TDD screen will carry the time dimension from your previous filters, which may be more granular than needed (displaying too many columns). Change the ```Metric Trends by``` filter on the top right to switch to alternate ranges like days, weeks, etc.
:::