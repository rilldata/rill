<script lang="ts">
  export let billingPlan: string | null;

  $: showBranding = billingPlan === "free-plan" || billingPlan === "growth-plan";
</script>

{#if showBranding}
  <a
    href="https://www.rilldata.com"
    target="_blank"
    rel="noopener noreferrer"
    class="pointer-events-auto fixed bottom-3 right-3 z-50 flex items-center gap-1.5 rounded-full bg-white/90 px-3 py-1.5 text-xs font-medium text-gray-600 shadow-sm ring-1 ring-gray-200/50 backdrop-blur-sm transition-colors hover:text-gray-900 hover:shadow"
  >
    Made with <span class="font-semibold text-gray-900">Rill</span>
  </a>
{/if}
