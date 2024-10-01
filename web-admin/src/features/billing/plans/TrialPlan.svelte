<script lang="ts">
  import type {
    V1BillingPlan,
    V1Subscription,
  } from "@rilldata/web-admin/client";
  import { getTrialMessageForDays } from "@rilldata/web-admin/features/billing/banner/handleTrialPlan";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import PricingDetails from "@rilldata/web-admin/features/billing/PricingDetails.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { DateTime } from "luxon";

  export let organization: string;
  export let plan: V1BillingPlan;
  export let subscription: V1Subscription;

  let trialEndMessage: string;
  $: {
    const today = DateTime.now();
    const endDate = DateTime.fromJSDate(new Date(subscription.trialEndDate));
    if (endDate.isValid) {
      trialEndMessage = getTrialMessageForDays(endDate.diff(today));
    }
  }

  let open = false;
</script>

<SettingsContainer title={plan.displayName ?? plan.name}>
  <div slot="body">
    <div>
      {trialEndMessage} Ready to get started with Rill?
      <PricingDetails />
    </div>
    <PlanQuotas {organization} quotas={plan.quotas} />
  </div>
  <svelte:fragment slot="contact">
    <span>For custom enterprise needs,</span>
    <Button type="link" compact forcedStyle="padding-left:2px !important;">
      contact us
    </Button>
  </svelte:fragment>

  <Button type="primary" slot="action" on:click={() => (open = true)}>
    End trial and start Team plan
  </Button>
</SettingsContainer>

<StartTeamPlanDialog bind:open {organization} type="base" />
