---
title: Quotas
description: Information about default quotas on Rill Cloud
sidebar_label: Quotas
sidebar_position: 20
---

## Organization-level quotas

Most quotas on Rill Cloud are enforced at the organization level. The default quotas for free accounts are:

| Resource             | Quota |
| :------------------- | ----: |
| Projects             |     5 |
| Deployments          |    10 |
| Slots                |    20 |
| Slots per deployment |     5 |
| Unaccepted invites   |   200 |

If you need higher quotas, please [reach out](https://www.rilldata.com/contact).

## User-level quotas

Rill Cloud enforces most quotas at the organization level. The only quota enforced on users is the number of free organizations you can create. The default quota is:

| Resource                   | Quota |
| :------------------------- | ----: |
| Free organizations created |     3 |

## Slots

Rill Cloud currently uses a concept of *slots* for the amount of computing resources made available to a deployment.

**One slot is equivalent to 1 vCPU, 2 GB memory, and 5 GB of SSD disk space.**

New deployments are assigned 2 slots by default. You can create larger deployments using the `--slots` flag when running `rill deploy`.
