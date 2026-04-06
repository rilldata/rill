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
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { DateTime } from "luxon";

  let {
    organization,
    subscription,
    plan,
    showUpgradeDialog,
  }: {
    organization: string;
    subscription: V1Subscription;
    plan: V1BillingPlan;
    showUpgradeDialog: boolean;
  } = $props();

  let categorisedIssues = $derived(
    useCategorisedOrganizationBillingIssues(organization),
  );
  let trialIssue = $derived($categorisedIssues.data?.trial);
  // prefer using end date from BillingIssues since we use that to hibernate projects and take other actions
  let subscriptionEndDate = $derived(
    trialIssue?.metadata?.onTrial?.endDate ?? subscription?.trialEndDate,
  );

  let trialInfo = $derived.by(() => {
    let message = "";
    let ended = false;
    if (trialIssue.type === V1BillingIssueType.BILLING_ISSUE_TYPE_TRIAL_ENDED) {
      message = "Your trial has expired.";
      ended = true;
    } else {
      const today = DateTime.now();
      const endDate = DateTime.fromJSDate(new Date(subscriptionEndDate));
      if (endDate.isValid) {
        const diff = endDate.diff(today);
        message = getTrialMessageForDays(endDate.diff(today));
        ended = diff.milliseconds < 0;
      }
    }
    return { message, ended };
  });

  let title = $derived(
    (plan?.displayName || "Trial plan") + (trialInfo.ended ? " expired" : ""),
  );

  let open = $state(false);
  $effect(() => {
    if (showUpgradeDialog) open = true;
  });
  let type: TeamPlanDialogTypes = $derived(
    trialInfo.ended ? "trial-expired" : "base",
  );
</script>

<SettingsContainer {title}>
  {#snippet body()}
    <div>
      <div>
        {trialInfo.message} Ready to get started with Rill?
        <a
          href="https://www.rilldata.com/pricing"
          target="_blank"
          rel="noreferrer noopener">See pricing details -></a
        >
        {#if plan}
          <!-- if there is no plan then quotas will be set to 0. It doesnt make sense to show this then -->
          <PlanQuotas {organization} />
        {/if}
      </div>
    </div>
  {/snippet}
  {#snippet contact()}
    <span>For custom enterprise needs,</span>
    <ContactUs />
  {/snippet}

  {#snippet action()}
    <Button type="primary" onClick={() => (open = true)}>
      {#if trialInfo.ended}
        Start Team plan
      {:else}
        End trial and start Team plan
      {/if}
    </Button>
  {/snippet}
</SettingsContainer>

{#if !$categorisedIssues.isLoading}
  <StartTeamPlanDialog bind:open {organization} {type} />
{/if}
