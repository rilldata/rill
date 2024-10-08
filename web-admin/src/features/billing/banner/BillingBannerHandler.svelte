<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
  } from "@rilldata/web-admin/client";
  import { handleBillingIssues } from "@rilldata/web-admin/features/billing/banner/handleBillingIssues";
  import StartTeamPlanDialog, {
    type TeamPlanDialogTypes,
  } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";

  export let organization: string;

  $: subscription = createAdminServiceGetBillingSubscription(organization);
  $: issues = createAdminServiceListOrganizationBillingIssues(organization);

  let showStartTeamPlanDialog = false;
  let startTeamPlanType: TeamPlanDialogTypes = "base";
  let teamPlanEndDate = "";

  $: if (!$subscription.isLoading && !$issues.isLoading) {
    handleBillingIssues(
      organization,
      $subscription.data.subscription,
      $issues.data.issues ?? [],
      (type, endDate) => {
        showStartTeamPlanDialog = true;
        startTeamPlanType = type;
        teamPlanEndDate = endDate;
      },
    );
  }
</script>

<StartTeamPlanDialog
  bind:open={showStartTeamPlanDialog}
  type={startTeamPlanType}
  endDate={teamPlanEndDate}
  {organization}
/>
