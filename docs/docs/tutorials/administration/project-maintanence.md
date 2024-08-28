---
title: "Project Maintanence"
description: Project Maintanence
sidebar_label: "Project Maintanence"

---

## Status Page

The Status Page gives us an overview of all the components within Rill Cloud, including the underlying source and models. While you will not be able to make any direct changes, the Status page is a good place to start when dashboards are acting strange.

![img](/img/tutorials/203/status.png)

You'll see here that there's an option to connect to GitHub.
During our first deployment onto Rill Cloud, we opted for a one-time upload. By doing so, we are able to directly deploy the project without any further steps, but we lose out on a few powerful capablities that can enhance the user experience, such as version control.

### When a dashboard is failing to load

When a dashboard fails to load, you will see an `Error` in the UI. There are a few potential causes for a dashboard to fail to load, but the best place to start is the Status page. For example, you might see the following in the UI: 

![img](/img/tutorials/admin/failing-dashboard.png)

In order to understand why this is failing, you can navigate to the Status page and find the dashboard's error message:

![img](/img/tutorials/admin/failing-status-page.png)

In this case, we can find that the table, `staging_to_CH` does not exist! We can see that this table fails to create due to the following error:

```bash
connection: dial tcp 127.0.0.1:9000: connect: connection refused
```

Seeing as this is ClickHouse model, it is likely that the credentials or connections are not correct for this connection. 

Whether it's the source or the model that is erroring and causing the dashboard to fail, you may need to [check the credentials](credential-mangement.md) back in Rill Developer.


