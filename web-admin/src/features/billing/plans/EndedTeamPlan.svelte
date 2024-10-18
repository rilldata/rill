<script lang="ts">
  import { type V1Subscription } from "@rilldata/web-admin/client";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import PricingDetails from "@rilldata/web-admin/features/billing/PricingDetails.svelte";
  import { useCategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { DateTime } from "luxon";

  export let organization: string;
  export let subscription: V1Subscription;
  export let showUpgrade: boolean;

  $: plan = subscription?.plan;
  $: categorisedIssues = useCategorisedOrganizationBillingIssues(organization);
  $: cancelledSubIssue = $categorisedIssues.data?.cancelled;

  let willEndOnText = "";
  $: if (cancelledSubIssue?.metadata.subscriptionCancelled?.endDate) {
    const endDate = DateTime.fromJSDate(
      new Date(cancelledSubIssue.metadata.subscriptionCancelled.endDate),
    );
    if (endDate.isValid && endDate.toMillis() > Date.now())
      willEndOnText = endDate.toLocaleString(DateTime.DATE_MED);
  }

  let open = showUpgrade;
</script>

<SettingsContainer title="Team Plan" titleIcon="info">
  <div slot="body">
    <div>
      Your plan is cancelled
      {#if willEndOnText}
        but you still have access until {willEndOnText}.
      {:else}
        and your subscription has ended.
      {/if}
      <PricingDetails />
      {#if plan}
        <PlanQuotas {organization} quotas={plan.quotas} />
      {/if}
    </div>
  </div>
  <svelte:fragment slot="contact">
    <span>For custom enterprise needs,</span>
    <Button type="link" compact forcedStyle="padding-left:2px !important;">
      contact us
    </Button>
  </svelte:fragment>

  <Button type="primary" slot="action" on:click={() => (open = true)}>
    Renew Team plan
  </Button>
</SettingsContainer>

{#if !$categorisedIssues.isLoading}
  <StartTeamPlanDialog
    bind:open
    {organization}
    type="renew"
    endDate={cancelledSubIssue?.metadata.subscriptionCancelled?.endDate}
  />
{/if}
