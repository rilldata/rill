<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
    createAdminServiceListPublicBillingPlans,
  } from "@rilldata/web-admin/client";
  import { showUpgradeDialog } from "@rilldata/web-admin/features/billing/banner/bannerCTADialogs";
  import { handleBillingIssues } from "@rilldata/web-admin/features/billing/banner/handleBillingIssues";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";

  export let organization: string;

  $: subscription = createAdminServiceGetBillingSubscription(organization);
  const plans = createAdminServiceListPublicBillingPlans();
  $: plan = $plans?.data?.plans?.find(
    (p) => p.id === $subscription.data?.subscription?.planId,
  );
  $: issues = createAdminServiceListOrganizationBillingIssues(organization);

  $: if ($subscription.data?.subscription && plan && $issues.data?.issues) {
    handleBillingIssues(
      $subscription.data.subscription,
      plan,
      $issues.data.issues,
    );
  }
</script>

<StartTeamPlanDialog bind:open={$showUpgradeDialog} {organization} />
