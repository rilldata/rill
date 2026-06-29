<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let { managed = false }: { managed?: boolean } = $props();

  function handleContactSales() {
    window.Pylon("show");
  }
</script>

<PlanContainer
  badge={managed ? m.billing_plan_badge_managed() : m.billing_plan_badge_enterprise()}
  description={managed ? m.billing_enterprise_desc_managed() : m.billing_enterprise_desc()}
>
  {#snippet action()}
    <button class="contact-us-btn" onclick={handleContactSales}>
      {m.billing_contact_us_cta()}
    </button>
  {/snippet}

  <p class="text-sm text-fg-tertiary mt-4 pb-4">
    {m.billing_enterprise_details()}
  </p>
</PlanContainer>

<style lang="postcss">
  .contact-us-btn {
    @apply text-sm font-medium text-primary-600 border border-primary-500 px-4 py-2 cursor-pointer bg-white rounded-none;
    height: 36px;
  }

  .contact-us-btn:hover {
    @apply bg-primary-50;
  }
</style>
