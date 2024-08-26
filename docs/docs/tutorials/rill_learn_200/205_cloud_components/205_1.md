---
title: "Let's make an alert"
description:  How to make an alert, what makes sense
sidebar_label: "Create an Alert"
sidebar_position: 2
---
## Alert!

While scheduled reports are good for reporting information, another important feature that we need are [alerts](https://docs.rilldata.com/explore/alerts/). Based on a condition defined, we can create an alert to send out to ensure that proper action is taken.

### Create an alert

Let's say we want to create an alert if any single user has submitted more than 5 commits to the repository in the past two weeks. We can set the filter on the dashboard, then select the alert button.
<img src = '/img/tutorials/205/alert.gif' class='rounded-gif' />
<br />
You'll see in the bottom half of the UI whether or not based on the current filter if your alert will be sent. In our case, we can see that an alert wil be sent. This alert will be sent whenever the data refreshes. In our case as we set the data to refresh every 24 hours, the alert will trigger every 24 hours. 
:::tip
You need to set the desired filters for the alert before opening the alert UI as these will be inherited directly from your current view.
:::

We can see all of the alerts from the project UI page.

![img](/img/tutorials/205/alert.png)

You can edit/delete the alert from this page. In the history section, you can see whether the condition was triggered or not. In the case that the trigger was not hit, your alert should not send you a message. Only project or organization admins can manage alerts.

:::tip
Don't forget about using the snooze option if you want to pause notifications.
:::


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />