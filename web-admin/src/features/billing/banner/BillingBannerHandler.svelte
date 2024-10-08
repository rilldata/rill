<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import { handleBillingIssues } from "@rilldata/web-admin/features/billing/banner/handleBillingIssues";
  import StartTeamPlanDialog, {
    type TeamPlanDialogTypes,
  } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";

  export let organization: string;

  $: subscription = createAdminServiceGetBillingSubscription(organization);
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);

  let showStartTeamPlanDialog = false;
  let startTeamPlanType: TeamPlanDialogTypes = "base";
  let teamPlanEndDate = "";

  $: if (!$subscription.isLoading && $categorisedIssues.data) {
    handleBillingIssues(
      organization,
      $subscription.data.subscription,
      $categorisedIssues.data,
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
