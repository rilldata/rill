---
title: "Alerts"
description: Setting up Alerts in Rill
sidebar_label: "Alerts"
sidebar_position: 40
---

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/n8OAwi5-tk4?si=5k65x9zNLM2huca6"
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


## Overview

Alerting is a key element for any BI or analytics workflow. Because Rill's dashboards are typically built off of raw or near-raw data, we expose alerting on a wide range of filters and depth, including high cardinality fields. Alerts are accessible from any dashboard via the upper-right alarm bell icon and can be used to create context-based triggers or alerts to bring you back to an analysis if an alert is triggered. This allows the analyst or end user to then dive deeper and use Rill dashboards to interactively explore their data as needed.


<img src = '/img/explore/alerts/alerts.gif' class='rounded-gif' />
<br />


## Setting and managing alerts

To set an alert, click on the alarm bell icon in the upper-right nav bar of Rill Cloud (next to your profile and bookmarks). This will trigger the Alert modal to open up and walk you through the alert creation process.


<img src = '/img/explore/alerts/alerts_icon.png' class='centered' />
<br />

### Data Selection

When creating an alert, it is important to note that any existing filters will _automatically_ be carried over into the alert but can be adjusted or removed as needed. In the first tab of creating your alert, you will want to set the following:
1. Add a name for the alert **(required)**
2. Set what measure to alert on **(required)**, including any dimension splits for the metric _(optional)_ and time grains for analysis _(optional)_

Before selecting **Next**, you will have the optional to preview the data for which you alert will be created on.

:::tip Maximize Your Alert Filters
To avoid getting over alerted, consider adding a metric filter to avoid long tail changes on small values. Some examples include:
- Creating a filter on customers with less than 100 logins <u>but</u> filtered to more than 100 users to alert on active user drops within active accounts.
- Creating a filter against campaigns with a greater than 20% decrease in spend <u>but</u> filtered to spend greater than $1000 to avoid large percentage changes on small campaigns.

For more information, refer to our documentation on [metric filters](/explore/filters#filter-by-measures).
:::

### Criteria

On the second tab, you will have the opportunity to specify the criteria for which your alert will be triggered once certain conditions are met.
1. For any metric in your data selection (previous tab), set your alert criteria to include an operator (e.g. less than, greater than), the value or percentage, and a comparison period and/or threshold amount.
2. You will also have the ability to set dependencies or add multiple criteria using boolean conditions (AND/OR) to combine conditions across measures.

### Delivery

On the final tab, you will choose how and where your alert is delivered. By default, the alert will be checked whenever the source data is [refreshed](/build/models/data-refresh). There are a few additional things worth noting:
1. To limit the number of alerts, you can set an optional **Snooze** period after an alert is triggered.
2. Depending on the available notification targets (see next section), choose which targets and/or users to subscribe to the alert.

Afterwards, click **Create** to finish creating the alert.

## Available alert notification targets

Rill Cloud currently supports the following notification targets:
- Email (default)
- Slack (can be enabled)

When creating an alert, all available notification targets that can be configured for an alert will be presented in the **Delivery** tab.

:::note Interested in other alerting notification targets?

If there is a potential alerting destination that you'd like to use with Rill but don't currently see available as a target, please don't hesitate to [contact us](/contact). We're always iterating and would love to learn more about your use case!

:::

### Configuring email targets

Email is the default notification target for alerts and is automatically enabled. When creating an alert, simply specify the email addresses to include for a particular alert and an email will be sent with a link to the alert when the alert is triggered.

<img src = '/img/explore/alerts/email-notifications.png' class='centered' />
<br />

### Configuring Slack targets

Slack is also an available target for alert notifications and Rill can be configured to send alerts to your workspace, either in specified Slack channels (public / private) or as private messages via a configured bot. However, Slack will <u>first need to be enabled</u> to show up as an available notification target for alerts. For more information, refer to our [Configuring Slack integration](/build/connectors/data-source/slack) documentation.


<img src = '/img/explore/alerts/slack-notifications.png' class='centered' />
<br />

:::warning Adding your Slack app to the correct channels

After having the Slack admin create the app / bot with appropriate permissions, please make sure to <u>first</u> add it to your target channels (using the */invite* command). Otherwise, the Slack alert will trigger an error that the channel can't be found when trying to post a notification!

:::

## Managing & Editing Alerts

To view or make changes to existing alerts, navigate to the project home page and select the `Alerts` tab. Selecting an alert will give details on the configured alert criteria, including frequency and filters. You will also have the option to edit the alert settings.

<img src = '/img/explore/alerts/project_home_alerts.png' class='rounded-gif' />
<br />

## Common use cases

### Troubleshooting 
Alerts for troubleshooting purposes are useful for making sure that applications are running as expected, campaigns are set up correctly, or any use case where the outcome is binary. For these alerts, the criteria is often: is the amount > 0 or is the amount below a threshold. These alerts are best mixed with dimension filters to be alerted on any instance or split (e.g. Impressions > 0 for all Campaign_ID).

### Pacing 
Alerts for pacing purposes are good for budgeting and threshold use cases - where a pre-defined range can be applied to evaluate progress towards a goal. There alerts tend to be more specific (setting up filters and criteria for specific values) and marking progress towards that goal. Consider setting up multiple threshold alerts like 50%, 75% attainment.

### Monitoring & Comparison
Alerts for monitoring purposes are probably the most common alerting use case, i.e. wanting to be alerted based on relative values to prior periods. For these alerts, rather than absolutes, create criteria for % change of values. 



