<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";

  export let organization: string;
  export let plan: V1BillingPlan;
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
</SettingsContainer>
