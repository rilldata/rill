---
title: "Alerts in Rill"
description: Project Maintanence
sidebar_label: " Project Resource Alerting"
tags:
  - CLI
  - Administration
  - Rill Developer
  - Advanced Features
  - Tutorial
---

There are two different types of alerting in Rill. 

1. [Project Error Alerting](https://docs.rilldata.com/deploy/project-errors)
2. [Rill Cloud Dashboard Alerting](https://docs.rilldata.com/explore/alerts/)

## Project Error Alerting

When deploying to Rill Cloud, there might be a few different reasons why a project goes into an error state. However, this may not be obvious during deployment. Therefore, there is an option to create project alerting based on some sort of query that you can derive from your project. 


### Setting up the YAML file in Rill Developer

Let's go ahead and create a [basic alert](/reference/project-files/alerts) on the project that sends an email if any of the resources reconcile with error.

The default alerting is:
```yaml
type: alert

# Check the alert every 10 minutes.
refresh:
  cron: "*/10 * * * *"

# Query for all resources with a reconcile error.
# The alert will trigger when the query result is not empty.
data:
  resource_status:
    where_error: true

# Send notifications by email
notify:
  email:
    recipients: [john@example.com]
```

So as not to spam our own inbox, let's change the alert to run the 1st of every month by changing the CRON to:
```
  cron: "0 0 1 * *"
```

Once complete, we can go ahead and create a broken dashboard or model. The easiest would be to create a new dashboard via +Add, and leaving it as is. There will be a solid red border around the text editor. 

<img src = '/img/tutorials/alert/new-dashboard.png' class='rounded-gif' />
<br />

Once we've created this, let's [push our changes to Rill Cloud](/tutorials/rill_developer_advanced_features/advanced_developer/update-rill-cloud). 


<img src = '/img/tutorials/alert/failing-dashboard.png' class='rounded-gif' />
<br />
Now, you'll receive an email that gives you more information on the failing resource, in this case the dashboard. 


<img src = '/img/tutorials/alert/alert-email.png' class='rounded-gif' />
<br />
You can view all the alerts (project and dashboard) from the Alerts page

<img src = '/img/tutorials/alert/alert-code.png' class='rounded-gif' />
<br />

