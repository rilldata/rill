<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
    V1BillingPlanType,
  } from "@rilldata/web-admin/client";
  import { mergedQueryStatus } from "@rilldata/web-admin/client/utils";
  import BillingContactSetting from "@rilldata/web-admin/features/billing/contact/BillingContactSetting.svelte";
  import Payment from "@rilldata/web-admin/features/billing/Payment.svelte";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";
  import {
    isEnterprisePlan,
    isProPlan,
    isTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { PageData } from "./$types";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";

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
  let planName = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.name ?? "",
  );
  let isPaidPlan = $derived(
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_PRO ||
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_TEAM ||
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_ENTERPRISE ||
      isProPlan(planName) ||
      isTeamPlan(planName) ||
      isEnterprisePlan(planName),
  );
  let isCancelled = $derived(Boolean($categorisedIssues.data?.cancelled));

  let cancelOpen = $state(false);

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
      <Plan
        {organization}
        {showUpgradeDialog}
        {billingPortalUrl}
        bind:cancelOpen
      />
    {/if}
    {#if isPaidPlan}
      <Payment {organization} />
    {/if}
    <BillingContactSetting {organization} />
    {#if isPaidPlan && !isCancelled}
      <button class="cancel-link" onclick={() => (cancelOpen = true)}>
        Cancel subscription
      </button>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .cancel-link {
    @apply text-sm font-medium text-fg-tertiary bg-transparent border-none cursor-pointer p-0;
    display: inline-block;
  }

  .cancel-link:hover {
    @apply text-fg-secondary underline;
  }
</style>
