---
title: FAQ
sidebar_label: FAQ
sidebar_position: 20
---

## Technical requirements

### Why does macOS say "Rill cannot be opened because it is from an unidentified developer"?
This occurs when Rill binary is downloaded via the browser. You need to change the permissions to make it executable and remove it from Apple Developer identification quarantine. 

The below CLI commands will help you to do that: 
```bash
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```

### Why am I seeing "This macOS version is not supported. Please upgrade"?
Rill uses duckDB internally which requires a newer [macOS version](https://github.com/duckdb/duckdb/issues/3824). 
Please upgrade your macOS version to 10.14 or higher.


### Which browsers work best with Rill?
Rill is optimized for Google Chrome. While other browsers may work, we recommend using the latest version of Chrome for the most reliable experience when accessing Rill Developer or Rill Cloud dashboards.


## Rill Developer

![dev](/img/concepts/rcvsrd/empty-project.png)

### What is Rill Developer?
Rill Developer is a local application used to preview your project and make any necessary changes before deploying to Rill Cloud. For more information, please review [our documentation](https://docs.rilldata.com/concepts/developerVsCloud#rill-developer).

### I'm having issues with Rill Developer...

Please refer to [our tutorials](/tutorials) to get started using Rill. If you still have any questions, please [contact us!](/contact)


### How do I start more than one instance of Rill Developer?

If you try to start two instances of Rill Developer, you will hit the following error:
```bash
Error: serve: server crashed: grpc port 49009 is in use by another process. Either kill that process or pass `--port-grpc PORT` to run Rill on another port
```

In other to run two instances, please use the following flags with a unique port number.
```bash
rill start --port 10010 --port-grpc 10011
```

### How do I share my dashboard with my colleagues?

You need to [deploy your dashboard to Rill Cloud](https://docs.rilldata.com/deploy/existing-project/) to share your dashboard.

## Rill Cloud

![dev](/img/concepts/rcvsrd/Rill-Cloud.png)


### What is Rill Cloud?
Rill Cloud is where your deployed Rill project exists and can be shared to your colleagues, or end-users. For more information, please review [our documentation](https://docs.rilldata.com/concepts/developerVsCloud#rill-cloud).

### How do I deploy to Rill Cloud?
You can deploy your project directly from the UI by selecting [the Deploy button](/deploy/deploy-dashboard/#deploying-a-project-from-rill-developer).

<img src = '/img/deploy/existing-project/deploy-ui.gif' class='rounded-gif' />
<br />


### How do I make changes to my dashboard in Rill Cloud?

You can follow the same steps as above. The button will have changed from `deploy` to `update`. After selecting this, the objects in your Rill project will be updated. Or, after syncing your Rill project to Github, simply push changes directly to the repository and this will automatically update your project on Rill Cloud.

### How do I share my dashboard to other users?

You will need to [invite users to your organization/project](https://docs.rilldata.com/manage/user-management#option-1---admin-invites-user) or send them a URL for them to [request access to your dashboard](https://docs.rilldata.com/manage/user-management#option-2---user-requests-access). If you just want them to see the contents of your dashboard, you can look into using [public URLs](https://docs.rilldata.com/explore/share-url).


## Rill Cloud Trial
### What is Rill Cloud Trial?
We offer a free 30 day trial to any one interested in testing out our online platform. Simply create an account and deploy your project from Rill Developer. If you haven't already created and account and logged in, you will be prompted during the deployment process. 

There are no feature limitations in a free trial but we have set the limit for imported data to 10GB per project with two projects per deployment. You can check the data usage in the settings page. 

:::note 
The banner will show you the remaining days for your trial and will update as the expiration gets closer! Upgrade to a Teams plan and input your payment method to continue using Rill!
:::
![img](/img/FAQ/rill-trial-banner.png)

### When does my trial start?
Your trial will start when you deploy a project to Rill Cloud from Rill Developer. An Organization will be autoamatically created during this process using your email and the project will be the folder that your Rill project exists in. You can change the name using [CLI commands](https://docs.rilldata.com/reference/cli/project/rename). 

### How long does my Rill Cloud Trial last?
A Rill Cloud trial lasts for 30 days. If you have any comments or concerns, please reach out to us on our [various platforms](../contact.md)! 

### What is included in the free trial? 
The free trial is locked at 2 projects and up to 10GB of data each. You can invite as many users as required and there are no locked features. 

### What happens to my project if I do not upgrade to a Team plan?
Your projects will hibernate. Your project metadata will still be available once you've activated your team plan. If you'd like to delete your deployment from Rill Cloud, you can do so via the [CLI commands.](https://docs.rilldata.com/reference/cli/org/delete)

![expired](/img/FAQ/expired-project.png)

### What is project hibernation?
When a project is inactive for a specific number of days or your trial has expired, we automatically hibernate the project. What this means is that all of your information and metadata is saved and resource consumption will be zero. You will need to unhibernate the project to gain access to the dashboard again. 

If the project is hibernated due to payment issues, the project will stay in this state until payment is confirmed. Once the payment is confirmed, you can reaccess the project with the following CLI command.
```
rill project hibernate <project_id> --redeploy
```

## Rill Team Plan
### What is a Rill Team Plan?
A Rill Team Plan unlocks unlimited projects with a 50GB data storage limit per project. Pricing starts at $250/month and includes 10GB of storage. Use the [pricing calculator](https://www.rilldata.com/pricing) on our pricing site for more insight into how much your data might cost! You'll now have access to all of our features on Rill Cloud that you were using during the trial. 

### How many seats am I allowed?
At Rill, we do not charge per seat! From subscription to a Rill Team Plan, you'll have access to unlimited seats! Invite all of your colleagues or just a few, the choice is yours. 

### How are payments calculated?
We charge you by the amount of data that you load into Rill when building your sources and models. Use the [pricing calculator](https://www.rilldata.com/pricing) on our pricing site for more insight into how much your data might cost! If you'd like a further detailed inspection of your objects, [contact us](../contact) and we'll set this up for you. 

### When am I billed? 
You'll be billed on the first of each month via our partner at Stripe. You'll need to set up a valid credit card as explained in [our billing documentation](/manage/account-management/billing#managing-payment-information). If ther are any issues with the card, you'll be notified in the UI and be given a few day grace period to update your information. If you start in the middle of the month, you'll be billed prorated for the number of days you have access to Rill Cloud.

### Why was I billed $XXX? 
You can check your data usage in your organization setting usage page. The graph will display the data that you have over 10GB. Use the [pricing calculator](https://www.rilldata.com/pricing) to see your cost of your current data usage.  If you'd like a further detailed inspection of your objects, [contact us](../contact) and we'll set this up for you. 

## Enterprise Plan

### What is an Enterprise Plan? 
Enterprise plan includes all the features of a Team Plan but also provides further offerings, such as a dedicated Technical Account Manager and less restrictions on data storage. For more information, please visit our price page, [here](https://www.rilldata.com/pricing), or [contact us](../contact.md). Transparent usage-based billing means you only pay for what you need. Flexible pricing based on storage, compute, and network units start at the rates below:

**Storage:**

Storage is the total compressed data in the cluster. It's available in [two performance tiers](/home/FAQ#what-are-the-compute-requirements-for-each-performance-tier), Hot and Cold, which set minimum [compute requirements](/home/FAQ#what-are-the-compute-requirements-for-data-processing).

Data can be also offloaded to an archival tier where it does not consume any compute

`$0.0005 / GB per hour`


**Compute:**

[Rill Compute Units (RCU)](/home/FAQ#what-is-a-rill-compute-unit-rcu) are a combination of CPU, memory, and disk used for ingesting and querying data.

RCU scale up elastically for data ingestion & processing with enterprise discounts on RCUs provisioned for querying.

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

- **Performance tier**: Hot tier consumes more RCU/TB of stored data than cold tier. This provides a good cost vs performance tradeoff. Cost can be reduced by moving data to the cold tier.

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

| **Storage Tier** | **Performance**                     | **RCU Consumption**                 |
|-------------------|-------------------------------------|--------------------------------------|
| Hot Performance   | 8 RCU per 25 GB of data stored     | High performance for frequently accessed data |
| Cold Performance  | 8 RCU per 250 GB of data stored    | Optimized for less frequently accessed data |
| Archival          | No RCU consumed                   | Data is not queryable and can be moved to Hot or Cold tiers as needed |

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
|------------------------|-------------------------|
| Up to 64 RCU          | 8 RCU                  |
| Up to 128 RCU         | 16 RCU                 |
| Up to 256 RCU         | 32 RCU                 |
| Over 256 RCU          | 64 RCU                 |
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