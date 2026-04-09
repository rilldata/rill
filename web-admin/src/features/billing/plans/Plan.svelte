<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import PlanCards from "@rilldata/web-admin/features/billing/plans/PlanCards.svelte";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
  import {
    isManagedPlan,
    isTeamPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";

  let {
    organization,
    showUpgradeDialog,
  }: {
    organization: string;
    showUpgradeDialog: boolean;
  } = $props();

  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let subscription = $derived($subscriptionQuery?.data?.subscription);
  let plan = $derived(subscription?.plan);

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );

  let neverSubbed = $derived(!!$categorisedIssues.data?.neverSubscribed);
  let isTrial = $derived(!!$categorisedIssues.data?.trial);
  let subHasEnded = $derived(!!$categorisedIssues.data?.cancelled);
  let subIsTeamPlan = $derived(plan && isTeamPlan(plan.name));
  let subIsManagedPlan = $derived(plan && isManagedPlan(plan.name));

  type PlanTier = "trial" | "pro" | "enterprise";
  let currentPlan: PlanTier = $derived.by(() => {
    if (neverSubbed || isTrial || subHasEnded) return "trial";
    if (subIsTeamPlan) return "pro";
    if (subIsManagedPlan) return "enterprise";
    return "enterprise";
  });

  let dialogType: TeamPlanDialogTypes = $derived.by(() => {
    if (subHasEnded) return "renew";
    if (isTrial) {
      const trialIssue = $categorisedIssues.data?.trial;
      if (trialIssue?.type === "BILLING_ISSUE_TYPE_TRIAL_ENDED")
        return "trial-expired";
      return "base";
    }
    return "base";
  });

  let renewEndDate = $derived(
    $categorisedIssues.data?.cancelled?.metadata?.subscriptionCancelled
      ?.endDate ?? "",
  );
</script>

<PlanCards
  {organization}
  {currentPlan}
  {showUpgradeDialog}
  {dialogType}
  {renewEndDate}
/>
