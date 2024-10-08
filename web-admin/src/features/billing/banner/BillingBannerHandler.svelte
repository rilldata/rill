<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import {
    showUpgradeDialog,
    upgradeDialogType,
  } from "@rilldata/web-admin/features/billing/banner/bannerCTADialogs";
  import { handleBillingIssues } from "@rilldata/web-admin/features/billing/banner/handleBillingIssues";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";

  export let organization: string;

  $: subscription = createAdminServiceGetBillingSubscription(organization);
  $: issues = createAdminServiceListOrganizationBillingIssues(organization);

  $: if (!$subscription.isLoading && !$issues.isLoading) {
    handleBillingIssues(
      organization,
      $subscription.data.subscription,
      $issues.data.issues ?? [],
    );
  }
</script>

<StartTeamPlanDialog
  bind:open={$showUpgradeDialog}
  type={$upgradeDialogType}
  {organization}
/>
