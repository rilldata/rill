<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  let {
    organization,
    hasPayment,
    plan,
    billingPortalUrl,
  }: {
    organization: string;
    hasPayment: boolean;
    plan: V1BillingPlan;
    billingPortalUrl: string | undefined;
  } = $props();

  let open = $state(false);
</script>

<SettingsContainer title={plan?.displayName}>
  <div>
    <div>You're currently on a custom contract.</div>
    {#if billingPortalUrl}
      <div>
        <a
          href={billingPortalUrl}
          target="_blank"
          rel="noreferrer noopener"
          class="invoice-link">View Invoice</a
        >
      </div>
    {/if}
    <PlanQuotas {organization} />
  </div>
  {#snippet contact()}
    <span>To make changes to your contract,</span>
    <ContactUs variant="enterprise" />
  {/snippet}
  {#snippet action()}
    {#if hasPayment}
      <Button type="primary" onClick={() => (open = true)}>
        Start Team plan
      </Button>
    {/if}
  {/snippet}
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
