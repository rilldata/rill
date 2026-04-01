<script lang="ts">
  import type { V1BillingPlan, V1Subscription } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import { getNextBillingCycleDate } from "@rilldata/web-admin/features/billing/plans/selectors";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";

  export let organization: string;
  export let subscription: V1Subscription;
  export let plan: V1BillingPlan;
</script>

<SettingsContainer title={plan?.displayName ?? "Growth Plan"}>
  <div slot="body">
    Next billing cycle will start on
    <b>{getNextBillingCycleDate(subscription.currentBillingCycleEndDate)}</b>.
    <a
      href="https://www.rilldata.com/pricing"
      target="_blank"
      rel="noreferrer noopener">See pricing details -></a
    >
    <PlanQuotas {organization} />
  </div>
  <svelte:fragment slot="contact">
    <span>For any questions,</span>
    <ContactUs />
  </svelte:fragment>
</SettingsContainer>
