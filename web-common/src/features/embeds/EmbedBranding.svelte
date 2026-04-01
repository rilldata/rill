<script lang="ts">
  import RillFilled from "@rilldata/web-common/components/icons/RillFilled.svelte";

  export let billingPlan: string | null;

  // TODO: restore plan check before merging: billingPlan === "free-plan" || billingPlan === "growth-plan"
  $: showBranding = true;
</script>

{#if showBranding}
  <a
    href="https://www.rilldata.com"
    target="_blank"
    rel="noopener noreferrer"
    class="pointer-events-auto fixed bottom-3 right-6 z-50 flex items-center gap-1.5 rounded-full bg-surface-overlay/90 px-3 py-1.5 text-xs font-medium text-fg-secondary shadow-sm ring-1 ring-border/50 backdrop-blur-sm transition-colors hover:text-fg-primary hover:shadow"
  >
    <RillFilled size="14" />
    Made with <span class="font-semibold text-fg-primary">Rill</span>
  </a>
{/if}
