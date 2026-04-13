<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import { mergedQueryStatus } from "@rilldata/web-admin/client/utils";
  import BillingContactSetting from "@rilldata/web-admin/features/billing/contact/BillingContactSetting.svelte";
  import Payment from "@rilldata/web-admin/features/billing/Payment.svelte";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  let organization = $derived(data.organization);
  let showUpgradeDialog = $derived(data.showUpgradeDialog);
  let billingPortalUrl = $derived(data.billingPortalUrl);

  let allStatus = $derived(
    mergedQueryStatus([
      createAdminServiceGetBillingSubscription(organization),
      createAdminServiceListOrganizationBillingIssues(organization),
    ]),
  );
</script>

<!-- Both the queries are used in both Plan and Payment.
     So instead of showing 2 spinner it is better to show one at the top. -->
{#if $allStatus.isLoading}
  <Spinner status={EntityStatus.Running} size="16px" />
{:else}
  <div class="flex flex-col gap-8">
    <Plan {organization} {showUpgradeDialog} />
    <Payment {organization} />
    <BillingContactSetting {organization} />
  </div>
{/if}
