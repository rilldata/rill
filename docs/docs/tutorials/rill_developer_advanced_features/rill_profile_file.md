---
title: "Rill Project File"
sidebar_label:  "Rill Project file"
hide_table_of_contents: false
---

The Rill project file `rill.yaml` is often overlooked but is a powerful tool as it sets defaults, environmental variables, and connector settings. Let's walk though setting up [`mock_users`](/manage/security#in-rill-developer) to test row access policies, default security settings for your metrics views and explore dashboards 

<img src = '/img/tutorials/admin/project.png' class='rounded-gif' />
<br />

## Mock Users 
Access to dashboards and even finer, access to rows, is a big concern for everyone. This is why mock users in Rill Developer is so important to be able to test the access of dashboards and data in dashboards before pushing to Rill Cloud. This is done via the mock_users in the project file. You can create pseudo users with specific domains, or admin and non admin users or user groups to test to ensure that access is correct. 

Let's assume that the following is applied to the metrics view.

```yaml
security:
    access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
    row_filter: "domain = '{{ .user.domain }}'"
```

In order to test both the access to the dashboard, as well as the row filter, we can create the following in the project YAML.

```yaml
mock_users:
  - email: royendo@rilldata.com
    admin: true
  - email: your_email@domain.com
    groups:
      - tutorial-admin
  - email: your_email2@another_domain.com
```





### Embedded dashboard Mock Users
Embedded dashboards are another type of dashboards that you'll want to test in Rill Developer. Mainly because there is no way to pass the custom variables in Rill Cloud to ensure that access and data is being presented correctly. In order to do this, you'll need to add the following to your mock_users:
```yaml
- email: embed@rilldata.com
  name: embed
  custom_variable_1: Value_1 #this is passed at embed creation
  custom_variable_2: Value_2 #this is passed at embed creation
```
See our [Custom Attributes Embedded dashboard](https://rill-embedding-example.netlify.app/rowaccesspolicy/custom) live!

Let's assume a similar set up to the above example. Within the metrics view, we define:

```yaml
security:
  access: true
  row_filter: >
    app_site_name = '{{ .user.app_site_name }}' AND
      pub_name = '{{ .user.pub_name }}'
```

Then within the application we are passing

```yaml
app_site_name='Sling'
pub_name='MobilityWare'
```

You can do a test mock user to ensure that this dashboard is working as designed with the following:

```yaml
- email: embed@rilldata.com
  name: embed
  app_site_name: 'Sling' 
  pub_name: 'MobilityWare'
```
<img src = '/img/tutorials/admin/custom-attribute-mock-user.png' class='rounded-gif' />
<br />



## Metrics Defaults

:::tip Order of Operations 

Rill YAML settings < Metrics View YAML
:::

## Explore Defaults


:::tip Order of Operations 
Rill YAML settings < Explore Dashboard YAML < Bookmarks in Rill Cloud < User Last State
:::


