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
  export let billingPortalUrl: string | undefined;

  let open = false;
</script>

<SettingsContainer title={plan?.displayName}>
  <div slot="body">
    <div>You're currently on a custom contract.</div>
    {#if billingPortalUrl}
      <a
        href={billingPortalUrl}
        target="_blank"
        rel="noreferrer noopener"
        class="invoice-link">View Invoice</a
      >
    {/if}
    <PlanQuotas {organization} />
  </div>
  <svelte:fragment slot="contact">
    <span>To make changes to your contract,</span>
    <ContactUs variant="enterprise" />
  </svelte:fragment>
  <svelte:fragment slot="action">
    {#if hasPayment}
      <Button type="primary" onClick={() => (open = true)}>
        Start Team plan
      </Button>
    {/if}
  </svelte:fragment>
</SettingsContainer>

<StartTeamPlanDialog bind:open {organization} type="base" />

<style lang="postcss">
  .invoice-link {
    @apply text-sm text-primary-500 no-underline mt-2 inline-block;
  }
  .invoice-link:hover {
    @apply text-primary-600 underline;
  }
</style>
