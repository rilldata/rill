---
title: "Billing Plans Explained"
description: How billing works for non-enterprise accounts
sidebar_label: Billing Plans Explained
sidebar_position: 00
---

Billing cycles begin on the first of every month (12:00 AM UTC). If you start your plan mid-month, your first month will be prorated accordingly. You can subscribe to a Team Plan at any point via your Rill Cloud billing page.


### How does it work?
Rill Data does not use a user-based license system. Instead, we calculate your data usage after ingestion and calculate pricing based on that usage. For more information on pricing, see our [pricing page](https://www.rilldata.com/pricing).


## Trial Plan

Get started with Rill Cloud with our 30-day free trial! Upon deployment of your first project, your trial will automatically start. On a free trial, you will be allowed one project with up to 10 GB of data storage. Like all plans in Rill Data, this also comes with unlimited seats. As an admin, you'll notice banners at the top of the UI indicating the remaining time left on your trial. Once your time has run out, your projects in Rill Cloud will hibernate. While your project won't be accessible on Rill Cloud, the files will still be available if you choose to upgrade to a Team Plan.

![Deploy Project](/img/manage/billing/deploy-project.png)


### Upgrading to Team Plan
Once you are ready to upgrade to a Team Plan, you can do so via the organization billing page, or select `Upgrade` in the top banner. Only organization administrators can upgrade the plan.

![Team Plan](/img/manage/billing/team-plan.png)


### Managing Payment Information

Please add a payment method and billing information that is accepted by Stripe. For more information, please visit [Stripe's website](https://docs.stripe.com/payments/payment-methods/overview).

![Stripe](/img/manage/billing/stripe.png)


## Team Plan

The Team Plan unlocks unlimited projects with a 50 GB data storage limit per project. Like all plans in Rill Data, this also comes with unlimited seats. As an admin, you will have access to your billing and usage page to monitor your project. If you decide to unsubscribe from your subscription, you will have access to Rill Cloud until the end of the month. Afterward, your project will hibernate.
Your project will not be accessible while hibernating. You will need to renew your subscription in order to access your project on Rill Cloud.

To calculate your current usage and pricing, see our [pricing page](https://www.rilldata.com/pricing).

![Team Plan2](/img/manage/billing/team-plan2.png)



## Enterprise Plan

The Enterprise Plan includes all the features of a Team Plan but also provides further offerings, such as a dedicated Technical Account Manager and fewer restrictions on data storage. For more information, please visit our [pricing page](https://www.rilldata.com/pricing), or [contact us](/contact).

### Enterprise usage-based billing

**Storage:**

Storage is the total compressed data in the cluster. It's available in [two performance tiers](/developers/other/FAQ#what-are-the-compute-requirements-for-each-performance-tier), Hot and Cold, which set minimum [compute requirements](/developers/other/FAQ#what-are-the-compute-requirements-for-data-processing).

Data can also be offloaded to an archival tier where it does not consume any compute.

`$0.0005 / GB per hour`


**Compute:**

[Rill Compute Units (RCU)](/developers/other/FAQ#what-is-a-rill-compute-unit-rcu) are a combination of CPU, memory, and disk used for ingesting and querying data.

RCUs scale up elastically for data ingestion and processing, with enterprise discounts on RCUs provisioned for querying.

`$0.09 RCU per hour`
