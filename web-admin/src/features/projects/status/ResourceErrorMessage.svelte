<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let message: string;
</script>

{#if message}
  <Tooltip distance={8}>
    <span class="text-red-600 max-w-xs line-clamp-2">
      {message}
    </span>
    <TooltipContent slot="tooltip-content">
      {message}
    </TooltipContent>
  </Tooltip>
{:else}
  <span>-</span>
{/if}
