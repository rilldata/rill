<script lang="ts">
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";

  export let billingPlan: string | null;

  $: showBranding =
    billingPlan === "free-plan" || billingPlan === "growth-plan";
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
