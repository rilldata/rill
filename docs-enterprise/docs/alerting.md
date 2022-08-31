---
title: "Alerting"
slug: "alerting"
hidden: false
createdAt: "2021-06-17T00:29:55.235Z"
updatedAt: "2021-08-12T17:11:52.418Z"
---
[block:api-header]
{
  "title": "Alerting"
}
[/block]
To access alerts, click the bell icon at the top of the page to bring up the alerts window.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/17526da-Alerts.png",
        "Alerts.png",
        2002,
        579,
        "#edf4f7"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
Within the alerts window, set the
  * Alert name
  * Metric for the alert
  * Increase/decrease value
  * (*optional*) Add additional recipients

Once the settings are added, Rill calculates the frequency for the current options. 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/dab855c-Alert_Setting.png",
        "Alert Setting.png",
        1541,
        783,
        "#fcfdfd"
      ],
      "sizing": "80"
    }
  ]
}
[/block]

[block:callout]
{
  "type": "info",
  "title": "Applied Filters",
  "body": "In the alerts window, you'll see the applied filters for the existing alert. Any combination of dimensions can be used to set alerts on specific values."
}
[/block]
Alerting is opportune for any troubleshooting or measurement use case. In particular:

  * Campaign delivery rates vs. expectations (e.g. zero values at start)
  * % drop/increase in expected revenue (by product/publisher)
  * Minimum partner spend
  * Changes in inventory
  * % change vs. regional benchmarks 

Once an alert is set, you can view/edit the alert in the Admin panel on the top right (the user icon dropdown).