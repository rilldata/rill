<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceGetOrganization,
  } from "@rilldata/web-admin/client";
  import { handleBillingIssues } from "@rilldata/web-admin/features/billing/banner/handleBillingIssues";
  import StartTeamPlanDialog, {
    type TeamPlanDialogTypes,
  } from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";

  export let organization: string;

  $: org = createAdminServiceGetOrganization(organization);
  $: subscription = createAdminServiceGetBillingSubscription(organization, {
    query: {
      enabled: !!$org.data?.permissions?.manageOrg,
    },
  });
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);

  let showStartTeamPlanDialog = false;
  let startTeamPlanType: TeamPlanDialogTypes = "base";
  let teamPlanEndDate = "";

  $: if (!$subscription.isLoading && $categorisedIssues.data) {
    handleBillingIssues(
      organization,
      $subscription.data?.subscription,
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
