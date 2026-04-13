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
    isManagedPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  let organization = $derived(data.organization);
  let showUpgradeDialog = $derived(data.showUpgradeDialog);
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let planType = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.planType,
  );
  let planName = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.name ?? "",
  );
  let isEnterprise = $derived(
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_ENTERPRISE ||
      planType === V1BillingPlanType.BILLING_PLAN_TYPE_MANAGED ||
      isManagedPlan(planName) ||
      isEnterprisePlan(planName),
  );

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
    <Plan {organization} {showUpgradeDialog} />
    {#if !isEnterprise}
      <Payment {organization} />
    {/if}
    <BillingContactSetting {organization} />
  </div>
{/if}
