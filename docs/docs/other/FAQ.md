---
title: FAQ
sidebar_label: FAQ
sidebar_position: 40

---

## Technical requirements

### Why does macOS say "Rill cannot be opened because it is from an unidentified developer"?
This occurs when the Rill binary is downloaded via the browser. You need to change the permissions to make it executable and remove it from Apple's developer identification quarantine. 

The CLI commands below will help you do that: 
```bash
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```

### Why am I seeing "This macOS version is not supported. Please upgrade"?
Rill uses DuckDB internally, which requires a newer [macOS version](https://github.com/duckdb/duckdb/issues/3824). 
Please upgrade your macOS version to 10.14 or higher.


### Which browsers work best with Rill?
Rill is optimized for Google Chrome. While other browsers may work, we recommend using the latest version of Chrome for the most reliable experience when accessing Rill Developer or Rill Cloud dashboards.


## Rill Developer

<img src = '/img/concepts/rcvsrd/empty-project.png' class='rounded-gif' />
<br />


### What is Rill Developer?
Rill Developer is a local application used to preview your project and make any necessary changes before deploying to Rill Cloud. For more information, please review [our documentation](https://docs.rilldata.com/concepts/cloud-vs-developer#rill-developer). Within Rill Developer, you can ingest new datasets, transform the sources into models, build a metrics layer, and finally visualize your data in an explore dashboard. This preview allows you to develop your project before deploying or updating an existing deployment in Rill Cloud.

### How do I do XXX in Rill Developer? 

Please refer to [our guided tutorial](/guides/rill-basics/launch) to get started using Rill. In the tutorials, we walk you through first project creation, modeling, creating a metrics view and explore dashboard, and finally deploying to Rill Cloud. From there, we go through making local changes in Rill Developer and pushing your changes. In more advanced topics, we discuss custom APIs, Embed Dashboards, and more! 
If you still have any questions, please [contact us!](/contact)


### How do I start more than one instance of Rill Developer?

If you try to start two instances of Rill Developer, you will encounter the following error:
```bash
Error: serve: server crashed: grpc port 49009 is in use by another process. Either kill that process or pass `--port-grpc PORT` to run Rill on another port
```

In order to run two instances, please use the following flags with a unique port number.
```bash
rill start --port 10010 --port-grpc 10011
```

### How do I share my dashboard with my colleagues?

To share your dashboards with your colleagues, you need to [deploy your dashboard to Rill Cloud](https://docs.rilldata.com/deploy/existing-project). Once deployed, you have various ways to share this dashboard with your team. Since Rill does not charge by number of users, you can simply [add them to your organization](../manage/user-management#how-to-add-an-organization-user) and have them sign up to view the dashboard! Other ways to share the dashboard include [public URLs](../explore/public-url) for a limited view and [project invites](../manage/user-management#how-to-add-a-project-user).

## Rill Cloud

<img src = '/img/concepts/rcvsrd/Rill-Cloud.png' class='rounded-gif' />
<br />



### What is Rill Cloud?
Rill Cloud is where your deployed Rill project exists and can be shared with your colleagues or end-users. For more information, please review [our documentation](https://docs.rilldata.com/concepts/cloud-vs-developer#rill-cloud). Unlike Rill Developer, which is developer-based, Rill Cloud is where your dashboards are consumed by your end users. Additional features include bookmarks, public URLs, reporting, alerts, and more! 

### How do I deploy to Rill Cloud?
You can deploy your project directly from the UI by selecting [the Deploy button](/deploy/deploy-dashboard/#deploying-a-project-from-rill-developer). Upon deployment, an organization will be automatically created with your Rill project inside. Each organization can have multiple projects that house multiple sources, models, metrics views, and dashboards. Note that sources are not cross-project compatible. When creating your first project, we will automatically start your 30-day free trial. In order to extend usage, you will need to sign up for a [Team Plan](#rill-team-plan).

<img src = '/img/deploy/existing-project/deploy-ui.gif' class='rounded-gif' />
<br />


### How do I make changes to my dashboard in Rill Cloud?

You can follow the same steps as above. After deploying to Rill Cloud, if you return to Rill Developer, the button will have changed from `deploy` to `update`. When selecting `update`, the objects in your Rill project will be automatically updated. Or, after syncing your Rill project to GitHub, simply push changes directly to the repository and this will automatically update your project on Rill Cloud.

### How do I share my dashboard to other users?

You will need to [invite users to your organization](../manage/user-management#how-to-add-an-organization-user) or [project](../manage/user-management#how-to-add-a-project-user), send them a URL for them to [request access to your dashboard](../manage/user-management#user-requests-access-via-url), or if you just want them to see the contents of your dashboard, you can look into using [public URLs](../explore/public-url).


## Rill Cloud Trial
### What is Rill Cloud Trial?
We offer a free 30-day trial to anyone interested in testing out Rill Cloud. Simply create an account and deploy your project from Rill Developer. If you haven't already created an account and logged in, you will be prompted during the deployment process. 

There are no feature limitations in a free trial, but we have set the limit for imported data to 10 GB per project with two projects per deployment. You can check the data usage in the settings page. 

:::note 
The banner will show you the remaining days for your trial and will update as the expiration gets closer! Upgrade to a Team plan to continue using Rill!
:::
<img src = '/img/FAQ/rill-trial-banner.png' class='rounded-gif' />
<br />


### When does my trial start?
Your trial will start when you deploy a project to Rill Cloud from Rill Developer. An Organization will be automatically created during this process using your email, and the project will be the folder that your Rill project exists in. You can change the name using [CLI commands](https://docs.rilldata.com/reference/cli/project/rename). 

### How long does my Rill Cloud Trial last?
A Rill Cloud trial lasts for 30 days. If you have any comments or concerns, please reach out to us on our [various platforms](/contact)! 

### What is included in the free trial? 
The free trial is limited to 2 projects and up to 10 GB of data each. You can invite as many users as required and there are no locked features. 

### What happens to my project if I do not upgrade to a Team plan?
Your projects will hibernate. Your project metadata will still be available once you've activated your team plan. If you'd like to delete your deployment from Rill Cloud, you can do so via the [CLI commands.](https://docs.rilldata.com/reference/cli/org/delete)

<img src = '/img/FAQ/expired-project.png' class='rounded-gif' />
<br />


### What is project hibernation?
When a project is inactive for a specific number of days or your trial has expired, we automatically hibernate the project. What this means is that all of your information and metadata is saved, and resource consumption will be zero. You will need to unhibernate the project to gain access to the dashboard again. 

If the project is hibernated due to payment issues, the project will stay in this state until payment is confirmed. Once the payment is confirmed, you can re-access the project with the following CLI command:
```
rill project hibernate <project_id> --redeploy
```

## Rill Team Plan
### What is a Rill Team Plan?
A Rill Team Plan unlocks unlimited projects with a 50 GB data storage limit per project. Pricing starts at $250/month and includes 10 GB of storage. Use the [pricing calculator](https://www.rilldata.com/pricing) on our pricing site for more insight into how much your data might cost! You'll now have access to all of our features on Rill Cloud that you were using during the trial. 

### How many seats am I allowed?
At Rill, we do not charge per seat! From subscription to a Rill Team Plan, you'll have access to unlimited seats! Invite all of your colleagues or just a few—the choice is yours. 

### How are payments calculated?
We charge you by the amount of data that you load into Rill when building your sources and models. Use the [pricing calculator](https://www.rilldata.com/pricing) on our pricing site for more insight into how much your data might cost! If you'd like a more detailed inspection of your objects, [contact us](../contact), and we'll set this up for you. 

### When am I billed? 
You'll be billed on the first of each month via our partner at Stripe. You'll need to set up a valid credit card as explained in [our billing documentation](/other/plans#managing-payment-information). If there are any issues with the card, you'll be notified in the UI and be given a few days' grace period to update your information. If you start in the middle of the month, you'll be billed prorated for the number of days you have access to Rill Cloud.

### Why was I billed $XXX? 
You can check your data usage in your organization settings usage page. The graph will display the data that you have over 10GB. Use the [pricing calculator](https://www.rilldata.com/pricing) to see your cost of your current data usage. If you'd like a more detailed inspection of your objects, [contact us](../contact), and we'll set this up for you. 

## Enterprise Plan

### What is an Enterprise Plan? 
Enterprise plan includes all the features of a Team Plan but also provides further offerings, such as a dedicated Technical Account Manager and fewer restrictions on data storage. For more information, please visit our pricing page, [here](https://www.rilldata.com/pricing), or [contact us](/contact). Transparent usage-based billing means you only pay for what you need. Flexible pricing based on storage, compute, and network units start at the rates below:

**Storage:**

Storage is the total compressed data in the cluster. It's available in [two performance tiers](/other/FAQ#what-are-the-compute-requirements-for-each-performance-tier), Hot and Cold, which set minimum [compute requirements](/other/FAQ#what-are-the-compute-requirements-for-data-processing).

Data can also be offloaded to an archival tier where it does not consume any compute

`$0.0005 / GB per hour`


**Compute:**

[Rill Compute Units (RCU)](/other/FAQ#what-is-a-rill-compute-unit-rcu) are a combination of CPU, memory, and disk used for ingesting and querying data.

RCUs scale up elastically for data ingestion & processing with enterprise discounts on RCUs provisioned for querying.

`$0.09 RCU per hour`

### What is a Rill Compute Unit? (RCU)
A Rill Compute Unit (RCU) is a usage metering unit that tracks, by the minute, the amount of resources consumed by your Rill cloud service, including compute, memory, disk storage, batch or streaming data ingested.

1. For data ingestion & processing, RCU scale up elastically as you load and transform data into our service.

2. For querying, Rill offers enterprise customers a set of dedicated compute units sized to handle their concurrency requirements across all data sources. Provisioned query RCUs are upscaled or downscaled on a daily basis based on the usage of past week to maintain query performance targets.

Tasks that consume more resources will see more RCU usage compared to tasks that consume fewer resources. While there is no one-to-one mapping between the various resources your service consumes and an RCU, 1 RCU is comparable to the resources used by a task that runs for one hour on 1 vCPU with 4 GB RAM.

### How is data size calculated in Rill?
When you load data into Rill's service, it is stored in compressed columnar format. Rill charges based on the amount of data stored in Rill's service, after compression, which is typically 3-8x.
For egress and ingress, actual data transferred over the network is measured.


### What are the compute requirements for each performance tier?
For querying, Rill provisions a fixed number of RCUs for each performance tier based on the following factors:

- **Size**: Rill will charge a minimum number of compute units per unit of hot storage and cold storage to maintain optimal performance

- **Queries**: As the concurrency of data access or complexity of the queries increase Rill will automatically add more RCUs to maintain optimal performance.

- **Performance tier**: Hot tier consumes more RCU/TB of stored data than cold tier. This provides a good cost vs performance trade off. Cost can be reduced by moving data to the cold tier.

Below are the minimum RCU charged for each performance tier:
<div
    style={{
    width: '100%',
    margin: 'auto',
    padding: '20px',
    textAlign: 'center', 
    display: 'flex', 
    justifyContent: 'center',
    alignItems: 'center'
    }}
>

| **Storage Tier** | **Performance**                 | **RCU Consumption**                                                    |
| ---------------- | ------------------------------- | ---------------------------------------------------------------------- |
| Hot Performance  | 8 RCU per 25 GB of data stored  | High performance for frequently accessed data                          |
| Cold Performance | 8 RCU per 250 GB of data stored | Optimized for less frequently accessed data                            |
| Archival         | No RCU consumed                 | Data is not query-able and can be moved to Hot or Cold tiers as needed |

</div>
The number of RCUs for each performance tier will be added in increments depending on the overall provisioned RCUs.
<div
    style={{
    width: '100%',
    margin: 'auto',
    padding: '20px',
    textAlign: 'center', 
    display: 'flex', 
    justifyContent: 'center',
    alignItems: 'center'
    }}
>
| **Overall RCU Count** | **Available Increment** |
| --------------------- | ----------------------- |
| Up to 64 RCU          | 8 RCU                   |
| Up to 128 RCU         | 16 RCU                  |
| Up to 256 RCU         | 32 RCU                  |
| Over 256 RCU          | 64 RCU                  |
</div>
Provisioned RCUs are upscaled or downscaled on a daily basis based on the usage of past week.

### What are the compute requirements for data processing?
For data ingestion & processing, Rill elastically scales up compute slots when you load data into the service.

- The number of RCUs consumed depends on the complexity of the ingestion pipeline.
- A more complex ingestion involving joins will consume more RCUs as compared to a less complex pipeline.

### How can I estimate my RCU usage?
The best way to get an accurate RCU estimate is to load some sample data into Rill service and track the RCU usage. For estimates on larger projects [contact us](../contact) for a pricing calculator that reflects the latest volume incentives and discounts.


### How can I track my RCU usage?
RCU usage and utilization can be viewed in the Rill Control Center. Additionally, Administrators can configure to get Daily, Weekly, Monthly summary reports for RCU usage for visibility.