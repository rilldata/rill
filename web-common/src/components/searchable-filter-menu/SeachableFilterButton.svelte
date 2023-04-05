<!-- @component 
This component is intended for filtering and selecting over 
lists of items that are small enough to be handled comfortably
the client, for example selecting visible measures and dimensions
in the dashboard, where in the the worst existing cases in the
legacy dash, the number of measures is not more than a few dozen
and the number of dimensions does not exceed a few hundred.

This component takes props:
- `selectableItems`:string[], an array of item names to be shown in the menu.
- `selectedItems`:boolean[], a bit mask indicating which items are currently selected.
These arrays must be the same length or the 

This component emits events:
- `itemClicked`, which has a number `detail` field with the index of the item that was clicked.
- `selectAll`, with no `detail`
- `deselectAll`, with no `detail`
In both cases, it is up to the containing component to handle the toggling the selection state and updating the `selectedItems` prop as needed.

-->
<script lang="ts">
  import { fly } from "svelte/transition";
  import WithTogglableFloatingElement from "../floating-element/WithTogglableFloatingElement.svelte";

  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import SearchableFilterDropdown from "./SearchableFilterDropdown.svelte";
  import SelectButton from "../menu/triggers/SelectButton.svelte";

  export let selectableItems: string[];
  export let selectedItems: boolean[];
  export let tooltipText: string;
  export let label: string;

  $: {
    if (
      selectableItems?.length > 0 &&
      selectedItems?.length > 0 &&
      selectableItems?.length !== selectedItems?.length
    ) {
      throw new Error(
        "SearchableFilterButton component requires props `selectableItems` and `selectedItems` to be arrays of equal length"
      );
    }
  }
  let active = false;
  $: numAvailable = selectableItems?.length ?? 0;
  $: numShown = selectedItems?.filter((x) => x).length ?? 0;

  $: numShownString =
    numAvailable === numShown ? "All" : `${numShown} of ${numAvailable}`;
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
    <SelectButton type="secondary" compact on:click={toggleFloatingElement}
      ><strong>{numShownString} {label}</strong></SelectButton
    >
    <div slot="tooltip-content" transition:fly|local={{ duration: 300, y: 4 }}>
      <TooltipContent maxWidth="400px">
        {tooltipText}
      </TooltipContent>
    </div>
  </Tooltip>
  <SearchableFilterDropdown
    {selectedItems}
    {selectableItems}
    slot="floating-element"
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    on:apply
    on:search
    on:itemClicked
    on:deselectAll
    on:selectAll
  />
</WithTogglableFloatingElement>
