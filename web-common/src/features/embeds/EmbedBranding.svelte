<script lang="ts">
  import RillFilled from "@rilldata/web-common/components/icons/RillFilled.svelte";

  export let billingPlan: string | null;

  $: showBranding = billingPlan === "free-plan" || billingPlan === "growth-plan";
</script>

{#if showBranding}
  <a
    href="https://www.rilldata.com"
    target="_blank"
    rel="noopener noreferrer"
    class="grid place-content-center p-2"
  >
    <RillFilled size="16" />
  </a>
{/if}
