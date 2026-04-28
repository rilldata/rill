<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let frequency: string;
</script>

{#if frequency}
  <Tooltip distance={8}>
    <span class="text-fg-primary truncate block">{frequency}</span>
    <TooltipContent slot="tooltip-content">{frequency}</TooltipContent>
  </Tooltip>
{:else}
  <span class="text-fg-tertiary">—</span>
{/if}
