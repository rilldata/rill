<script lang="ts">
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";

  let { managed = false }: { managed?: boolean } = $props();

  function handleContactSales() {
    window.Pylon("show");
  }
</script>

<PlanContainer
  badge={managed ? "Managed" : "Enterprise"}
  description="Custom contract{managed ? ' · Fully managed' : ''}"
>
  {#snippet action()}
    <button class="contact-us-btn" onclick={handleContactSales}>
      Contact us
    </button>
  {/snippet}

  <p class="text-sm text-fg-tertiary mt-4 pb-4">
    Fully managed slots, dedicated CSM, white-label capabilities, and custom
    SLAs. Contact your CSM for contract details or changes.
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
