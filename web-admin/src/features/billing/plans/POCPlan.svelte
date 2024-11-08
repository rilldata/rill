<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;
  export let hasPayment: boolean;
  export let plan: V1BillingPlan;

  let open = false;
</script>

<SettingsContainer title={plan?.displayName}>
  <div slot="body">
    <div>You’re currently on a custom contract.</div>
    <PlanQuotas {organization} />
  </div>
  <svelte:fragment slot="contact">
    <span>To make changes to your contract,</span>
    <ContactUs variant="enterprise" />
  </svelte:fragment>
  {#if hasPayment}
    <Button type="primary" slot="action" on:click={() => (open = true)}>
      Start Team plan
    </Button>
  {/if}
</SettingsContainer>

<StartTeamPlanDialog bind:open {organization} type="base" />
