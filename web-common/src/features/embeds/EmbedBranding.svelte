<script lang="ts">
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import { requiresEmbedBranding } from "@rilldata/web-common/features/embeds/embed-store.ts";

  export let billingPlan: string | null;

  $: showBranding = requiresEmbedBranding(billingPlan);
</script>

{#if showBranding}
  <a
    href="https://www.rilldata.com"
    target="_blank"
    rel="noopener noreferrer"
    class="grid place-content-center p-2"
  >
    <Rill />
  </a>
{/if}
