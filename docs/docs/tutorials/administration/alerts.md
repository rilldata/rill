---
title: "Alerting"
description: Project Maintanence
sidebar_label: "Alerting"
tags:
  - CLI
  - Administration
---

There are two different types of alerting in Rill. 

1. [Project Error Alerting](https://docs.rilldata.com/deploy/project-errors)
2. [Dashboard Alerting](https://docs.rilldata.com/explore/alerts/)

## Project Error Alerting

When deploying to Rill Cloud, there might be a few different reasons why a project goes into an error state. However, this may not be obvious during deployment. Therefore, there is an option to create project alerting based on some sort of query that you can derive from your project. 


### Setting up the YAML file in Rill Developer

Let's go ahead and create a basic alert on the project that sends an email if any of the resources reconcile with error.
[]link the alert YAML once out.
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
![img](/img/tutorials/admin/new-dashboard.png)


Once we've created this, let's [push our changes to Rill Cloud](/tutorials/rill_advanced_features/advanced_developer/update-rill-cloud). 


![img](/img/tutorials/admin/failing-dashboard.png)

Now, you'll receive an email that gives you more information on the failing resource, in this case the dashboard. 


![img](/img/tutorials/admin/alert-email.png)

You can view all the alerts (project and dashboard) from the Alerts page

![img](/img/tutorials/admin/alert-code.png)

## Dashboard Alerts

[Alerts created on individual dashboards](https://docs.rilldata.com/explore/alerts/) can be viewed from a project's alert page. As an admin, you can edit or delete the alert as needed.

![alerts](/img/tutorials/admin/alert-admin.png)