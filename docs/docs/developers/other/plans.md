---
title: Plans & Pricing
sidebar_label: Plans & Pricing
sidebar_position: 00
---

import { PlanCard, PlanCards } from '@site/src/components/PlanCard';
import FeatureTable from '@site/src/components/FeatureTable';

# Plans & Pricing

Rill offers two deployment modes, each with three pricing tiers.

## Deployment Modes

Available on Growth and Enterprise plans. The Free plan uses Managed mode only.

- **Managed** — Rill ingests, transforms, and hosts your data on managed infrastructure (DuckDB). Ideal for teams that want a fully managed experience.
- **Live Connect** — Rill connects directly to your existing OLAP warehouse (ClickHouse, BigQuery, Snowflake, etc.) without storing data. Ideal for teams with existing infrastructure.


## Managed Plans

<PlanCards>
  <PlanCard
    name="Free"
    price="$0/month"
    features={[
      { label: "Data", value: "Up to 1 GB" },
      { label: "Slots", value: "1 slot (2 GB RAM / 1 vCPU)" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "\"Made with Rill\" badge" },
    ]}
  />
  <PlanCard
    name="Growth"
    price="Usage-based"
    features={[
      { label: "Base fee", value: "None" },
      { label: "Slots", value: "$0.06/slot/hr" },
      { label: "Storage", value: "$10/GB/month above 1 GB" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "Fully customizable" },
    ]}
  />
  <PlanCard
    name="Enterprise"
    price="Custom"
    features={[
      { label: "Data", value: "Custom" },
      { label: "Slots", value: "Custom allocation" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "Fully customizable" },
      { label: "Extras", value: "CSM, SLAs, SSO" },
    ]}
    cta={{ text: "Contact sales →", link: "https://www.rilldata.com/contact" }}
  />
</PlanCards>


## Live Connect Plans

<PlanCards>
  <PlanCard
    name="Growth"
    price="Usage-based"
    features={[
      { label: "Base fee", value: "None" },
      { label: "Slots", value: "$0.06/slot/hr" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "Fully customizable" },
      { label: "Hibernation", value: "Metering pauses automatically" },
    ]}
  />
  <PlanCard
    name="Enterprise"
    price="Custom"
    features={[
      { label: "Slots", value: "Custom — negotiated" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "Fully customizable" },
      { label: "Extras", value: "CSM, SLAs, SSO" },
    ]}
    cta={{ text: "Contact sales →", link: "https://www.rilldata.com/contact" }}
  />
</PlanCards>


## Feature Comparison

### Managed vs Live Connect

<FeatureTable
  columns={["", "Managed", "Live Connect"]}
  rows={[
    ["Base fee", "None", "None"],
    ["Slot rate", "$0.06/slot/hr", "$0.06/slot/hr"],
    ["Storage", "$10/GB/month above 1 GB", "N/A (your warehouse)"],
    ["Data ceiling", "~250 GB", "Unlimited"],
    ["Best for", "Fully managed experience", "Existing OLAP infrastructure"],
  ]}
/>

### By Tier

<FeatureTable
  columns={["", "Free", "Growth", "Enterprise"]}
  rows={[
    ["Slots", "1", "Pay per slot", "Custom"],
    ["AI", true, true, true],
    ["Custom APIs", true, true, true],
    ["Embedded dashboards", "\"Made with Rill\" badge", true, true],
    ["Live Connect", false, true, true],
    ["GitHub sync", false, true, true],
    ["Custom branding", false, true, true],
    ["SAML SSO", false, false, true],
    ["SOC 2 Type II", false, false, true],
    ["SLAs", false, false, true],
    ["Dedicated CSM", false, false, true],
  ]}
/>


## Slots

**1 slot = 2 GB RAM / 1 vCPU** across both deployment modes. Slots power dashboard query performance.

- Slot allocation can be adjusted at any time from the project status page.
- Dev/branch environments default to 1 slot. Additional slots can be allocated from the project pool.


## Upgrading

You can upgrade to the Growth plan from the project status page or the organization settings page.

### Free to Growth
- **Managed:** Triggered when data exceeds the 1 GB limit. You'll receive a notification with a 7-day window to upgrade before the project hibernates.
- **Live Connect:** When enabled, exceeding the free slot limit will automatically trigger a 7-day upgrade window before the project hibernates.

### Growth to Enterprise
- Contact sales when you need SLAs, SSO, governance controls, or custom terms.

:::note
Each project has its own slot allocation. If you have multiple projects, check the status page for each project to review and manage slots independently.
:::

## Enterprise Usage-Based Billing

Enterprise plans include all Growth features plus dedicated support, SLAs, and custom terms. Billing is usage-based across two axes:

### Storage

Storage is the total compressed data in the cluster. It is available in [two performance tiers](/developers/other/FAQ#what-are-the-compute-requirements-for-each-performance-tier), Hot and Cold, which set minimum [compute requirements](/developers/other/FAQ#what-are-the-compute-requirements-for-data-processing).

Data can also be offloaded to an archival tier where it does not consume any compute.

`$0.0005 / GB per hour`

### Compute

[Rill Compute Units (RCU)](/developers/other/FAQ#what-is-a-rill-compute-unit-rcu) are a combination of CPU, memory, and disk used for ingesting and querying data. RCUs scale up elastically for data ingestion and processing, with enterprise discounts on RCUs provisioned for querying.

`$0.09 / RCU per hour`
