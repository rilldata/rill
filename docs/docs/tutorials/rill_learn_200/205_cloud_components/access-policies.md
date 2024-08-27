---
title: "Access policies"
description:  some further changes to dashboard view using access policies
sidebar_label: "Create an access policy"
---

## What are access policies?

While you can set up user access via user management, Rill also supports granular access for dashboards. [Access policies](https://docs.rilldata.com/manage/security) are set directly in the dashboard's YAML file and can use a user's attributes to allow or block to the dashboard or access to specific data within the dashboard.


### Using User attributes to set access
There are a few [user's attributes](https://docs.rilldata.com/manage/security) that can be used to set access to a dashboard. Our documentation goes into further details about the types of access policies that you can set up. However, as this depends on the emails that you used we can use a general example. 


### How to test access? 
We can use Rill Developer to test access to your dashboards before pushing changes to Rill Cloud. Please add the following lines to your rill.yaml, replacing the emails with the ones you've set up already.

We add the `admin:true` to the first email, your email, and to `your_email@domain.com` as this inherits the admin permission via the group `tutorial-admin`.
```
mock_users:
- email: my_emaill@domain.com
  admin: true
- email: your_email@domain.com
  groups:
    - tutorial-admin
  admin: true
- email: your_email2@domain.com
```

On the dashboard YAML file, you can add the following, as taken from our docs. This will provide access to users who are admins on the project or organization.

```
security:
  access: "{{ .user.admin }} 
```

On the dashboard preview page, you will see a new button next to the `Edit Metrics`, `View as`. Here you can select the emails defined in your rill.yaml file to test out the connection to your dashboard. In the example below, I use my rilldata email and can see the dashboard as I am an admin. However, other users who do not, will be redirected to a 404.


<img src = '/img/tutorials/205/access-policy.gif' class='rounded-gif' />
<br />

:::tip
In our example, we have set the full dashboard access via the security key pair. However, depending on your underlying data and your users, you can use the user's attributes to filter out specific data on the dashboard.

For example the below would filter out the data for column `domain` in the underlying table or model based on the user's email domain.
```
security:
  access: true
  row_filter: "domain = '{{ .user.domain }}'"
```
:::


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />
