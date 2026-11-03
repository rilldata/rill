# Verified audit: web-admin

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **11 numbered findings**. Every `web-admin-01` through `-11` appears exactly once.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| web-admin-01 | HIGH | P1 | Unvalidated redirect enables open redirect | `routes/[organization]/-/upgrade-callback/+page.svelte:34,97-100`; `billing/plans/upgrade-to-plan.ts:54-58`; `routes/-/auth/device/+page.svelte:17,30-31` still consume attacker-controlled redirect | Central allowlist helper; accept same-origin paths or approved origins only. |
| web-admin-02 | MEDIUM | P1 | View-As cache may show prior impersonated user | `routes/[organization]/[project]/+layout.svelte:196-207,281-290`; runtime query keys in `runtime-client/invalidation.ts:9-24` still exclude JWT/user | Key/invalidate by impersonated subject and remount provider on subject change. |
| web-admin-03 | MEDIUM | P1 | View-As falls back to admin permissions | `features/projects/project-runtime.ts:38-39` still uses `mockUser.permissions ?? projectData?.projectPermissions` | Fail closed while mocked permissions load/fail; show an explicit state. |
| web-admin-04 | LOW | P3 | Embed RPC cleanup callback is discarded | `routes/-/embed/+layout.svelte:108-113` still ignores `createIframeRPCHandler()` cleanup | Return cleanup from `onMount`. |
| web-admin-05 | LOW | P3 | Embed project-page conditional is dead | `routes/-/embed/+layout.svelte:43-53,117-142` still makes `activeResource` always truthy via fallback | Model listing/resource routes explicitly; remove dead branch. |
| web-admin-06 | LOW | P1 | Embed RPC accepts any origin and replies to `*` | `web-common/src/lib/rpc.ts:64-104` still has no origin validation and uses wildcard targets | Pin allowed parent origin through a handshake and exact `targetOrigin`. |
| web-admin-07 | LOW | P3 | Embed globally disables adminServer flag | `routes/-/embed/+layout.svelte:39`; root `routes/+layout.svelte:80` still set global state without ownership/restore | Scope the flag to context or restore on destroy. |
| web-admin-08 | LOW | P3 | Billing handler cache/store are dead | `features/billing/BillingCTAHandler.ts:12-29`; `BillingBannerManagerForAdmins.svelte:17` | Remove unused singleton/cache/store. |
| web-admin-09 | LOW | P2 | Unedited org form is considered changed | `features/organizations/settings/OrgNameSettings.svelte:91-103` still compares slug to display-name field | Compare the form with initial display name and separate slug mutation. |
| web-admin-10 | LOW | P2 | Public URL accepts past expiry | `features/public-urls/form-utils.ts:58-63`; `CreatePublicURLForm.svelte:96-98,183-193` still allow negative TTL | Set min date and reject/clamp non-positive TTL server-side too. |
| web-admin-11 | LOW | P2 | View-As picker silently caps users at 1000 | `features/view-as-user/ViewAsUserPopover.svelte:14-19` still requests one page of 1000 | Implement server search/pagination or infinite loading. |

## Fixed / remove

None. July 29-31 i18n edits touched several files but did not change these behaviors.

## Partial / needs test

None.

Counts: **11 pending, 0 fixed, 0 partial = 11**.

## Machine-readable inventory

```text
PENDING|web-admin.md|web-admin-01|HIGH|P1|-|Open redirect|upgrade-callback/+page.svelte:34,97-100
PENDING|web-admin.md|web-admin-02|MEDIUM|P1|-|View-As prior-user cache|[project]/+layout.svelte:196-207
PENDING|web-admin.md|web-admin-03|MEDIUM|P1|-|View-As admin permission fallback|project-runtime.ts:38-39
PENDING|web-admin.md|web-admin-04|LOW|P3|-|Embed RPC cleanup discarded|embed/+layout.svelte:108-113
PENDING|web-admin.md|web-admin-05|LOW|P3|-|Dead embed conditional|embed/+layout.svelte:43-53
PENDING|web-admin.md|web-admin-06|LOW|P1|-|Embed RPC origin not pinned|rpc.ts:64-104
PENDING|web-admin.md|web-admin-07|LOW|P3|-|Global flag not restored|embed/+layout.svelte:39
PENDING|web-admin.md|web-admin-08|LOW|P3|-|Dead billing state|BillingCTAHandler.ts:12-29
PENDING|web-admin.md|web-admin-09|LOW|P2|-|Org form false dirty state|OrgNameSettings.svelte:91-103
PENDING|web-admin.md|web-admin-10|LOW|P2|-|Negative public URL TTL|form-utils.ts:58-63
PENDING|web-admin.md|web-admin-11|LOW|P2|-|View-As capped at 1000|ViewAsUserPopover.svelte:14-19
```
