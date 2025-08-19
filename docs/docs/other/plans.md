---
title: "Billing Plans Explained"
description: How Billing works for non-enterprise accounts
sidebar_label: Billing Plans Explained
sidebar_position: 00
---

Billing cycles begin on the first of the every month (12:00am UTC). If you start your plan mid-month, your first month will be pro-rated accordingly. You can subcribe to a Team Plan at any point via your Rill Cloud billing page. 


### How does it work? 
Rill Data does not use a user-based license system. Instead, we calculate your data usage, after ingestion, and calculate the pricing based on this. For more information on pricing, see our [pricing page](https://www.rilldata.com/pricing). 


## Trial Plan

Get started with Rill Cloud with our 30 day free trial! Upon deployment of your first project, your trial will automatically start.  On a free trial, you will be allowed 1 project up to 10GB of data storage.  Like all plans in Rill Data, this also comes with unlimited seats. As an admin, you'll notice banners at the top of the UI indicating the remaining time left on your trial. Once your time has run out, your projects in Rill Cloud will hibernate. While your project wont be accessible on Rill Cloud, the files will still be available if your choose to upgrade to a Team plan.

<img src = '/img/manage/billing/deploy-project.png' class='rounded-gif' />
<br />


### Upgrading to Team Plan
Once you are ready to upgrade to a Team Plan, you can do so via the organization billing page, or select `Upgrade` in the top banner. Only organization administrators can upgrade the plan.

<img src = '/img/manage/billing/team-plan.png' class='rounded-gif' />
<br />


### Managing Payment Information

Please add a payment method and billing information that is accepted by Stripe. For more information please visit Stripe's website, [here.](https://docs.stripe.com/payments/payment-methods/overview)

<img src = '/img/manage/billing/stripe.png' class='rounded-gif' />
<br />


## Team Plan

Team Plan unlocks unlimited projects with a 50GB data storage limit per project. Like all plans in Rill Data, this also comes with unlimited seats. As an admin, you will have access to your billing and usage page to monitor your project. If you decide to unsubcribe from your subcription, you will have access to Rill Cloud until the end of the month. Afterwards, your project will hibernate.
Your project will not be accessible while hibernating. You will need to renew your subscription in order to access your project on Rill Cloud. 

To calculate your current usage and pricing, see our [pricing page](https://www.rilldata.com/pricing). 

<img src = '/img/manage/billing/team-plan2.png' class='rounded-gif' />
<br />



## Enterprise Plan

Enterprise plan includes all the features of a Team Plan but also provides further offerings, such as a dedicated Technical Account Manager and less restrictions on data storage. For more information, please visit our price page, [here](https://www.rilldata.com/pricing), or [contact us](../../contact).

### Enterprise usage-based billing

**Storage:**

Storage is the total compressed data in the cluster. It's available in [two performance tiers](/other/FAQ#what-are-the-compute-requirements-for-each-performance-tier), Hot and Cold, which set minimum [compute requirements](/other/FAQ#what-are-the-compute-requirements-for-data-processing).

Data can be also offloaded to an archival tier where it does not consume any compute

`$0.0005 / GB per hour`


**Compute:**

[Rill Compute Units (RCU)](/other/FAQ#what-is-a-rill-compute-unit-rcu) are a combination of CPU, memory, and disk used for ingesting and querying data.

RCU scale up elastically for data ingestion & processing with enterprise discounts on RCUs provisioned for querying.

`$0.09 RCU per hour`