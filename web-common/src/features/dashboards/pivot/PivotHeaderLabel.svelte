<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let label: string;
  export let description: string | undefined = undefined;
</script>

<Tooltip location="top" distance={8} activeDelay={200}>
  <p class="truncate">{label}</p>
  <TooltipContent slot="tooltip-content" maxWidth="280px">
    <div class="pointer-events-none items-baseline" aria-label="tooltip-name">
      {label}
    </div>
    {#if description}
      <div
        class="text-fg-inverse/70 pointer-events-none pt-1"
        style:max-width="280px"
        aria-label="tooltip-name-description"
      >
        {description}
      </div>
    {/if}
  </TooltipContent>
</Tooltip>
