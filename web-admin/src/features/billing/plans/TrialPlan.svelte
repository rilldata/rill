<script lang="ts">
  import {
    V1BillingIssueType,
    type V1BillingPlan,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import { getTrialMessageForDays } from "@rilldata/web-admin/features/billing/issues/getMessageForTrialPlan";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import type { TeamPlanDialogTypes } from "@rilldata/web-admin/features/billing/plans/types";
  import PricingDetails from "@rilldata/web-admin/features/billing/PricingDetails.svelte";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { DateTime } from "luxon";

  export let organization: string;
  export let subscription: V1Subscription;
  export let plan: V1BillingPlan;
  export let showUpgradeDialog: boolean;

  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: trialIssue = $categorisedIssues.data?.trial;
  // prefer using end date from BillingIssues since we use that to hibernate projects and take other actions
  $: subscriptionEndDate =
    trialIssue?.metadata?.onTrial?.endDate ?? subscription?.trialEndDate;

  let trialEndMessage: string;
  let trialEnded = false;
  $: {
    if (trialIssue.type === V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED) {
      trialEndMessage = "Your trial has expired.";
      trialEnded = true;
    } else {
      const today = DateTime.now();
      const endDate = DateTime.fromJSDate(new Date(subscriptionEndDate));
      if (endDate.isValid) {
        const diff = endDate.diff(today);
        trialEndMessage = getTrialMessageForDays(endDate.diff(today));
        trialEnded = diff.milliseconds < 0;
      }
    }
  }

  $: title = plan?.displayName + (trialEnded ? " expired" : "");

  let open = showUpgradeDialog;
  $: type = (trialEnded ? "trial-expired" : "base") as TeamPlanDialogTypes;
</script>

<SettingsContainer {title}>
  <div slot="body">
    <div>
      {trialEndMessage} Ready to get started with Rill?
      <PricingDetails />
      <PlanQuotas {organization} />
    </div>
  </div>
  <svelte:fragment slot="contact">
    <span>For custom enterprise needs,</span>
    <ContactUs />
  </svelte:fragment>

  <Button type="primary" slot="action" on:click={() => (open = true)}>
    {#if trialEnded}
      Start Team plan
    {:else}
      End trial and start Team plan
    {/if}
  </Button>
</SettingsContainer>

{#if !$categorisedIssues.isLoading}
  <StartTeamPlanDialog bind:open {organization} {type} />
{/if}
