<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  let {
    organization,
    plan,
  }: {
    organization: string;
    plan: V1BillingPlan;
  } = $props();

  let open = $state(false);
</script>

<SettingsContainer title={plan?.displayName ?? "Free Trial"}>
  <div slot="body">
    <div>
      You're on the Free Trial plan. Ready to get started with Rill?
      <a
        href="https://www.rilldata.com/pricing"
        target="_blank"
        rel="noreferrer noopener">See pricing details -></a
      >
    </div>
    <PlanQuotas {organization} />
  </div>
  <svelte:fragment slot="contact">
    <span>For any questions,</span>
    <ContactUs />
  </svelte:fragment>

  <Button type="primary" slot="action" onClick={() => (open = true)}>
    Upgrade to Team plan
  </Button>
</SettingsContainer>

<StartTeamPlanDialog bind:open {organization} type="base" />
