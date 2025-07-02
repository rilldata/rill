---
title: "Create Dashboard Policies"
sidebar_label:  "Create Dashboard Access Policies"
hide_table_of_contents: false
tags:
  - Rill Developer
  - Advanced Features
---

For this example, we will use the OpenRTB Programmatic Advertising project available in our [examples repository.](https://github.com/rilldata/rill-examples/tree/main/rill-openrtb-prog-ads) We will create a few dashbaord access policies and test this in Rill Developer before pushing to Rill Cloud!

## Dashboard Access 

This can be set both on the metric view level and at the dashboard level. Simply define an access statement like below:

```yaml
security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
```


## Granular Policies
Including true / false access to a Dashboard, Rill also allows you to customize the view of the dashboard based on [specific criteria](https://docs.rilldata.com/manage/security#restrict-dashboard-access-to-users-matching-specific-criteria). This includes removing specific dimensions and measures and filtering out row data.

This allow for deeper customizability of the dashboard experience without having to manage many dashboards. 

This simplest form of row access policy is mapping a dimension value to the `domain` of the user. However, we find that most of the real-life use cases do not follow such a similar one-to-one relationship. Either a single domain is required to view multiple domains and/or it doesn't map to a dimension. Therefore, in this example, we will cover creating a mapping file, and mapping domains to different values in the column, "Pub Name".

Let's look at the possible values for Pub Name: [Click Here](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/auction_explore)

Some values include: `Disney`, `Pluto TV`, `LG USA`, ...

## Dashboard Access via Mapping Table
### Creating the Mapping Table
There are many ways to set up the mapping file. Whether it's directly in a model SQL, or reading from an S3 bucket. You have the freedom to decide, in this example, we will make it directly in the models/model.sql file.

```SQL
-- mapping.sql in Rill
-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models

    SELECT * FROM (VALUES 
      ('Disney', 'domain.com'),
      ('Disney', 'rilldata.com'),
      ... -- add your email and pub name here!
    ) AS t(PubName, domain)
```

From this SQL file, we create a table that will map the `rilldata.com` domain to `Disney`. You will need to modify the file to add an email domain that fits your email and add a `Pub Name` from one that exists in the demo dashboard.

### Creating the Row level Dashboard Policy
Now that this is created, you have a few options on which level you want to create the security level policy. Some questions to ask yourself are:
1. Am I using the metrics view on other components other than the dashboard? i.e.: APIs, canvas dashboards
2. If I am using the metrics view in other locations, how strict do I want the metrics layer to be? 

For most situations, you would define the dashboard policies at the metrics view level. So let's do that. Let's create a new metrics view, `auction_data_model_metrics_row_policies.yaml`, and copy the contents of `auction_data_model_metrics.yaml` into it.

In our new file, we want to define the following security rule:
```yaml
security:
  access: true
  row_filter: "Pub_Name IN (SELECT PubName FROM mapping WHERE domain = '{{ .user.domain }}')"
```

From our created model `mapping`, we are running the following SQL statement.

```SQL
SELECT PubName FROM mapping WHERE domain = '{{ .user.domain }}'
```

Using the login information from our current user account, .user.domain will extract the domain from your email. In my case, rilldata.com is being extracted. Since the row `rilldata.com` matches, it returns the value in column PubName, 'Disney'. 

This translated back into the query, runs:

```SQL
security:
  access: true
  row_filter: "Pub_Name IN `Disney`"
```

Which results in the current view:

<img src = '/img/tutorials/other/row-policy/row-policy-view.png' class='rounded-gif' />
<br />

### Additional Set Up Possibilities

This is a relatively straightforward example of row policies in Rill. By setting up a mapping file, you can allow specific data to be visible for specific individuals.

Let's say you hired a contractor to assist with several customers. Mapping their domain or even email to the accounts will grant them visibility to only that specific data without having to create a new dashboard, metrics view, etc.

Grant Rill Data to three values in 'Pub Name', but for Roy only LG USA. You would need to modify the SQL statement in the metrics view to also accommodate email.
```SQL
-- Model SQL
-- Reference documentation: https://docs.rilldata.com/reference/project-files/models

    SELECT * FROM (VALUES 
      ('Disney', 'domain.com', ''),
      ('Disney', 'rilldata.com', ''),
      ('LG USA', 'rilldata.com', 'roy.endo@rilldata.com'),
      ('Pluto', 'rilldata.com', '')
    ) AS t(PubName, domain, email)
```
```yaml
  row_filter: "Pub_Name IN (SELECT PubName FROM mapping WHERE domain = '{{ .user.domain }}' {{ if .user.email}} AND email = '{{.user.email}}' {{ end }})"
```

For more possibilities on attributes please see [our documentation](https://docs.rilldata.com/manage/security#user-attributes).


## Dashboard Access (Column-level access)

Another use-case is removing columns for users or groups that do not need them. Going back to our example openrtb project, let's say you are creating this dashboard for bids but do not want to provide specific information to non-company viewers. While you could create a new dashboard and remove some dimensions and measures, you can also use the same dashboard with a specific security policy.

```yaml
security: 
  access: true #access is provided but 
  exclude: #if the user is not part of rilldata, then exclude:
    - if: "'{{ .user.domain }}' != 'rilldata.com'"
      names: 
        - bid_floor_bucket
        - auction_type
        - win_rate
```



In this case, if the user is not part of Rill Data, they will not be able to view the listed dimensions and measures. This is especially useful if your dashboard has customer sensitive information that should not be viewable. 



## Confidence in Publishing Changes
Once you've tested the following in Rill Developer and confirm that the access is behaving as required, you can be confident in publishing your changes to Rill Cloud. 

Did you know? You can also define default project access in the rill.yaml file. See [Rill Project file](/tutorials/rill_developer_advanced_features/rill_profile_file#embedded-dashboard-mock-users) for more information.