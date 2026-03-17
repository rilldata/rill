---
title: Plans & Pricing
sidebar_label: Plans & Pricing
sidebar_position: 00
---

import { PlanCard, PlanCards } from '@site/src/components/PlanCard';
import FeatureTable from '@site/src/components/FeatureTable';

# Plans & Pricing

Rill offers two deployment modes, each with three pricing tiers. AI is included at all tiers — no token math, no surprises.

## Deployment Modes

- **Managed** — Rill ingests, transforms, and hosts your data on managed infrastructure (DuckDB). Ideal for teams that want a fully managed experience. Practical ceiling ~250 GB.
- **Live Connect** — Rill connects directly to your existing OLAP warehouse (ClickHouse, MotherDuck, etc.) without storing your data. Ideal for teams with existing infrastructure.

Both modes share the same three-tier structure: **Free**, **Growth**, and **Enterprise**.

---

## Managed Plans

<PlanCards>
  <PlanCard
    name="Free"
    price="$250 credit"
    features={[
      { label: "Credit", value: "$250 on signup" },
      { label: "Data limit", value: "Up to 1 GB" },
      { label: "Slots", value: "1 slot (4 GB RAM / 1 vCPU)" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "\"Made with Rill\" badge" },
    ]}
  />
  <PlanCard
    name="Growth"
    price="Usage-based"
    features={[
      { label: "Base fee", value: "None" },
      { label: "Slots", value: "$0.15/slot/hr" },
      { label: "Storage", value: "$1/GB/month above 1 GB" },
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

### Managed Growth — reference bills

| Data Size | Slots | Compute/mo | Storage/mo | Total/mo |
|---|---|---|---|---|
| 1 GB | 1 | $0 | $0 | **$0** |
| 5 GB | 2 | ~$219 | ~$4 | **~$223** |
| 10 GB | 2 | ~$219 | ~$9 | **~$228** |
| 25 GB | 2 | ~$219 | ~$24 | **~$243** |
| 50 GB | 3 | ~$328 | ~$49 | **~$377** |
| 100 GB | 4 | ~$438 | ~$99 | **~$537** |
| 250 GB | 6 | ~$657 | ~$249 | **~$906** |

*Slots at $0.15/slot/hr × 730 hrs/month always-on. Storage at $1/GB/month above 1 GB.*

---

## Live Connect Plans

Live Connect pricing has two components:

- **Base price** — derived from your OLAP cluster size (~20% of your cluster cost, at $0.06/slot/hr)
- **Rill Slots** — user-controlled slots for extra performance and dev environments ($0.15/slot/hr, starts at 0)

<PlanCards>
  <PlanCard
    name="Free"
    price="$250 credit"
    features={[
      { label: "Credit", value: "$250 on deploy" },
      { label: "Base price", value: "~20% of your cluster cost" },
      { label: "Rill Slots", value: "0" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "\"Made with Rill\" badge" },
    ]}
  />
  <PlanCard
    name="Growth"
    price="Usage-based"
    features={[
      { label: "Base fee", value: "None" },
      { label: "Base price", value: "~20% of your cluster cost" },
      { label: "Rill Slots", value: "$0.15/slot/hr (user-controlled)" },
      { label: "Hibernation", value: "Metering pauses automatically" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "Fully customizable" },
    ]}
  />
  <PlanCard
    name="Enterprise"
    price="Custom"
    features={[
      { label: "Base price", value: "Custom rate" },
      { label: "Rill Slots", value: "Custom allocation" },
      { label: "AI", value: "Included" },
      { label: "Branding", value: "Fully customizable" },
      { label: "Extras", value: "CSM, SLAs, SSO" },
    ]}
    cta={{ text: "Contact sales →", link: "https://www.rilldata.com/contact" }}
  />
</PlanCards>

### Live Connect Growth — reference bills

| CHC Cluster | CHC Cost/mo | Rill Base/mo |
|---|---|---|
| Basic (8 GB / 2 vCPU) × 2 | $443 | ~$88 |
| Basic (12 GB / 3 vCPU) × 2 | $662 | ~$131 |
| Scale (16 GB / 4 vCPU) × 2 | $880 | ~$175 |
| Scale (32 GB / 8 vCPU) × 2 | $1,754 | ~$350 |
| Scale (64 GB / 16 vCPU) × 2 | $3,502 | ~$701 |
| Scale (120 GB / 30 vCPU) × 2 | $6,700 | ~$1,314 |

*CHC costs based on ClickHouse Cloud pricing, 2 replicas, always-on. Hibernated clusters are not billed.*

---

## Feature Comparison

### Managed vs Live Connect

<FeatureTable
  columns={["", "Managed", "Live Connect"]}
  rows={[
    ["Pricing", "Rill Slots at $0.15/slot/hr + $1/GB/month above 1 GB", "~20% of cluster cost + Rill Slots at $0.15/slot/hr"],
    ["Storage", "$1/GB/month above 1 GB", "N/A (your warehouse)"],
    ["Data ceiling", "~250 GB", "Unlimited"],
    ["Free tier", "$250 credit, 1 GB limit", "$250 credit on deploy"],
    ["Best for", "Fully managed experience", "Existing OLAP infrastructure"],
  ]}
/>

### By Tier

<FeatureTable
  columns={["", "Free", "Growth", "Enterprise"]}
  rows={[
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

---

## Slots

**1 slot = 4 GB RAM / 1 vCPU** across both deployment modes.

- Slot allocation can be adjusted at any time from the project status page.
- Dev/branch environments default to 1 Rill Slot. Slots can be reallocated between production and development environments.

---

## Upgrading

### Free → Growth

Both modes start with a **$250 credit**. When the credit is exhausted, your project hibernates. You'll receive an in-product warning at 80% usage (~$200 burned) with a prompt to upgrade.

### Growth → Enterprise

Contact sales when you need SLAs, SSO, governance controls, or custom terms.

- **Managed:** Enterprise is a natural conversation at 50 GB+ of data.
- **Live Connect:** Enterprise applies for 200+ users, procurement requirements, or custom pricing.

:::note
Each project has its own slot allocation. If you have multiple projects, check the status page for each project to review and manage slots independently.
:::
