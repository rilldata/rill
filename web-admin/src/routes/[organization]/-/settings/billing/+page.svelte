<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import { mergedQueryStatus } from "@rilldata/web-admin/client/utils";
  import BillingContactSetting from "@rilldata/web-admin/features/billing/contact/BillingContactSetting.svelte";
  import Payment from "@rilldata/web-admin/features/billing/Payment.svelte";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";
  import { PaidPlanTypes } from "@rilldata/web-admin/features/billing/plans/utils";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { PageData } from "./$types";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
  import PlanActions from "@rilldata/web-admin/features/billing/plans/PlanActions.svelte";

  let { data }: { data: PageData } = $props();

  let organization = $derived(data.organization);
  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );

  let showUpgradeDialog = $derived(data.showUpgradeDialog);
  let billingPortalUrl = $derived(data.billingPortalUrl);
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let planType = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.planType,
  );
  let isPaidPlan = $derived(PaidPlanTypes[planType]);

  let allStatus = $derived(
    mergedQueryStatus([
      subscriptionQuery,
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
    {#if !$categorisedIssues.data?.neverSubscribed}
      <Plan {organization} {showUpgradeDialog} {billingPortalUrl} />
    {/if}
    {#if isPaidPlan}
      <Payment {organization} />
    {/if}
    <BillingContactSetting {organization} />
    <PlanActions {organization} />
  </div>
{/if}
