---
title: "Alerts"
description: Setting up Alerts in Rill
sidebar_label: "Alerts"
sidebar_position: 40
---

## Overview

Alerting is obviously a key element for any BI or analytics workflow. Because Rill's dashboards are typically driven off of raw or near-raw data, we also expose alerting on a wide variety of filters and with depth of high cardnality fields. Alerts can be accessed from any leaderboard to provide the context of the dimension and will then return you to that analysis if the alert is triggered.

## Setting and managing alerts

To set an alert, expand a leaderboard table and select ```Create alert``` on the top right. The Alert modal will open to walk through alerting options:

### Data Selection
1. Any existing filters will be carried into the alert and may be adjusted or removed as needed.
2. Add a name for the alert (required)
3. Select the data for the alert: the metric to alert on, dimension split for the metric (_optional_), and time grain for analysis (_optional_).

:::tip Maximizing Alert Filters
To avoid getting over alerted, consider adding a metric filter to avoid long tail changes on small values. Couple of examples:
- filter to customers with less than 100 logins but filtered to more than 100 users to alert on active user drops in active accounts
- filter to campaigns with > 20% decrease in spend but filtered to spend > $1000 to avoid large % change on small campaigns

See [metric filters](../explore/filters.md#filter-by-metrics) for more info.
:::

### Criteria
1. For any metric in your data selection, set your alert criteria including operator (e.g. less than, greater than), value or percentage, comparison period and threshold/alert amount.
2. You can also set dependencies or add multiple criteria using boolean (AND/OR) to combine conditions across metrics.

#### Delivery
1. If you only want limited instances of alerts, you can select an optional snooze period after an alert is triggered.
2. Select other users to subscribe if necessary and click save to turn on the alert.

### Managing & Editing Alerts
To make changes to an existing alert, navigate to ```Home``` (top left) and select the Alerts tab. Selecting an alert will give details on the alert criteria including frequency and filters + an option to edit the alert settings.

![alerts](<../../static/img/explore/alerts/alerts.gif>)

## Common alerting use cases

### Troubleshooting 
Troubleshooting alerts are useful for making sure that applications are running as expected, campaigns are set up correctly, or any use case where the outcome is binary. For these alerts, the criteria is often is the amount > 0 or is the amount below a threshold. These alerts are best mixed with dimension filters to be alerted on any instance or split (e.g. Impressions > 0 for all Campaign_ID).

### Pacing 
Pacing alerts are good for budgeting and threshold use cases - where a pre-defined range can be applied to evaluate progress towards a goal. There alerts tend to be more specific (setting up filters and criteria for specifics values) and marking progress towards that goal. Consider setting up multiple threshold alerts like 50%, 75% attainment.

### Monitoring & Comparison
Monitoring alerts are probably the most common - wanting to be alerted based on relative values to prior periods. For these alerts, rather than absolutes, create criteria for % change of values. 



