<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";

  export let organization: string;
  export let plan: V1BillingPlan;
</script>

<SettingsContainer title={plan?.displayName ?? "Free"}>
  <div slot="body">
    <div>
      You're on the Free plan.
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
</SettingsContainer>
