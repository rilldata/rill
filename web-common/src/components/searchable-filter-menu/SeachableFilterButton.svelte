<!-- @component 
This component is intended for filtering and selecting over 
lists of items that are small enough to be handled comfortably
the client, for example selecting visible measures and dimensions
in the dashboard, where in the the worst existing cases in the
legacy dash, the number of measures is not more than a few dozen
and the number of dimensions does not exceed a few hundered.


-->
<script lang="ts">
  import { fly } from "svelte/transition";
  import WithTogglableFloatingElement from "../floating-element/WithTogglableFloatingElement.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import SearchableFilterDropdown from "./SearchableFilterDropdown.svelte";
  import type { SelectableItem } from "./types-and-utils";

  export let selectableItems: SelectableItem[];
  export let selectableItemsOutput: SelectableItem[];
  // export let selectedValues: boolean[];
  export let tooltipText: string;
  export let excludeMode = false;

  let active = false;
</script>

<WithTogglableFloatingElement
  let:toggleFloatingElement
  bind:active
  distance={8}
  alignment="start"
>
  <Tooltip
    location="bottom"
    alignment="start"
    distance={8}
    activeDelay={60}
    suppress={active}
  >
    <div on:click={toggleFloatingElement}>fake button</div>
    <div slot="tooltip-content" transition:fly|local={{ duration: 300, y: 4 }}>
      <TooltipContent maxWidth="400px">
        {tooltipText}
      </TooltipContent>
    </div>
  </Tooltip>
  <SearchableFilterDropdown
    bind:selectableItemsOutput
    {selectableItems}
    slot="floating-element"
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    on:apply
    on:search
  />
</WithTogglableFloatingElement>
