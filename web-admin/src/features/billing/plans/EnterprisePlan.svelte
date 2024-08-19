<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;
  export let plan: V1BillingPlan;
</script>

<PlanContainer {organization} {plan}>
  <svelte:fragment slot="description">
    You’re currently on a custom contract.
  </svelte:fragment>
  <svelte:fragment slot="contact">
    <span>To make changes to your contract,</span>
    <Button type="link" compact forcedStyle="padding-left:2px !important;">
      contact support
    </Button>
  </svelte:fragment>
</PlanContainer>
