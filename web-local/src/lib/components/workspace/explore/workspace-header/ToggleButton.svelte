<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let active = false;
  export let disabled = false;

  export let tooltip = undefined;
  export let tooltipLocation = "bottom";
  export let tooltipAlignment = "center";
  export let tooltipDistance = 8;
</script>

<Tooltip
  suppress={!tooltip?.length}
  location={tooltipLocation}
  alignment={tooltipAlignment}
  distance={tooltipDistance}
>
  <button
    on:click
    {disabled}
    class:cursor-disabled={disabled}
    class:bg-gray-300={active}
    class:text-gray-800={active}
    class:bg-transparent={!active && !disabled}
    class:bg-white={disabled}
    class:text-gray-500={!active && !disabled}
    class:text-gray-400={disabled}
    class:italic={disabled}
    class="rounded px-3 py-1"
  >
    <slot />
  </button>
  <TooltipContent slot="tooltip-content">
    {tooltip}
  </TooltipContent>
</Tooltip>
