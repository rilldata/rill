<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { startPylonChat } from "@rilldata/web-common/features/help/startPylonChat";

  export let organization: string;
  export let plan: V1BillingPlan;
</script>

<SettingsContainer title="Enterprise plan">
  <div slot="body">
    <div>You’re currently on a custom contract.</div>
    {#if plan}
      <PlanQuotas {organization} quotas={plan.quotas} />
    {/if}
  </div>
  <svelte:fragment slot="contact">
    <span>To make changes to your contract,</span>
    <Button
      type="link"
      compact
      forcedStyle="padding-left:2px !important;"
      on:click={startPylonChat}
    >
      contact support
    </Button>
  </svelte:fragment>
</SettingsContainer>
