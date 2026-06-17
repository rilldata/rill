<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    V1BillingIssueType,
  } from "@rilldata/web-admin/client";
  import { getPlanTierForSubscription } from "@rilldata/web-admin/features/billing/plans/selectors";
  import type {
    PlanTier,
    TeamPlanDialogTypes,
  } from "@rilldata/web-admin/features/billing/plans/types";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import ChoosePlanDialog from "@rilldata/web-admin/features/billing/plans/dialog/ChoosePlanDialog.svelte";
  import ProPlan from "@rilldata/web-admin/features/billing/plans/ProPlan.svelte";
  import SelfServePlanCard from "@rilldata/web-admin/features/billing/plans/SelfServePlanCard.svelte";
  import LegacyTeamPlan from "@rilldata/web-admin/features/billing/plans/LegacyTeamPlan.svelte";
  import FreePlan from "@rilldata/web-admin/features/billing/plans/FreePlan.svelte";
  import LegacyTrialPlan from "@rilldata/web-admin/features/billing/plans/LegacyTrialPlan.svelte";
  import EnterprisePlan from "@rilldata/web-admin/features/billing/plans/EnterprisePlan.svelte";

  let {
    organization,
    showUpgradeDialog,
    billingPortalUrl,
  }: {
    organization: string;
    showUpgradeDialog: boolean;
    billingPortalUrl: string | undefined;
  } = $props();

  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let subscription = $derived($subscriptionQuery?.data?.subscription);

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );

  let subHasEnded = $derived(!!$categorisedIssues.data?.cancelled);

  let currentPlan: PlanTier = $derived(
    getPlanTierForSubscription(subscription, $categorisedIssues.data),
  );

  let isTrialExpired = $derived(
    $categorisedIssues.data?.trial?.type ===
      V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED ||
      $categorisedIssues.data?.trial?.type ===
        V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_CREDITS_DEPLETED,
  );

  let dialogType: TeamPlanDialogTypes = $derived.by(() => {
    if (subHasEnded) return "renew";
    if (isTrialExpired) return "trial-expired";
    return "base";
  });

  let renewEndDate = $derived(
    $categorisedIssues.data?.cancelled?.metadata?.subscriptionCancelled
      ?.endDate ?? "",
  );

  // Upgrade dialog
  let upgradeDialogOpen = $state(false);
  $effect(() => {
    if (showUpgradeDialog) upgradeDialogOpen = true;
  });

  function showUpgradeProDialog() {
    upgradeDialogOpen = true;
  }
</script>

{#if currentPlan === "free"}
  <FreePlan {organization} upgrade={showUpgradeProDialog} />
{:else if currentPlan === "pro"}
  <ProPlan {billingPortalUrl} />
{:else if currentPlan === "starter"}
  <SelfServePlanCard tier="starter" {billingPortalUrl} />
{:else if currentPlan === "growth"}
  <SelfServePlanCard tier="growth" {billingPortalUrl} />
{:else if currentPlan === "trial"}
  <LegacyTrialPlan
    {organization}
    {subscription}
    upgrade={showUpgradeProDialog}
  />
{:else if currentPlan === "team"}
  <LegacyTeamPlan {billingPortalUrl} />
{:else if currentPlan === "managed"}
  <EnterprisePlan managed />
{:else if currentPlan === "enterprise"}
  <EnterprisePlan />
{/if}

<ChoosePlanDialog
  bind:open={upgradeDialogOpen}
  {organization}
  type={dialogType}
  endDate={renewEndDate}
/>
