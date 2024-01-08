<!-- @component 
This component is intended for filtering and selecting over 
lists of items that are small enough to be handled comfortably
the clientInformation. The canonical example is
selecting visible measures and dimensions
in the dashboard, where in the the worst existing cases from the
legacy dash, the number items does not exceed a few hundred.

This component takes props:
- `selectableItems`:string[], an array of item names to be shown in the menu.
- `selectedItems`:boolean[], a bit mask indicating which items are currently selected.
These arrays must be the same length or the the component will
throw an error.

This component emits events:
- `item-clicked`. This event has has a `detail` field containing an object of type `{ index: number; label: string}` with the index and the label of the item that was clicked.
- `select-all`, with no `detail`
- `deselect-all`, with no `detail`
In both cases, it is up to the containing component to handle these
events, toggling the selection state and passing in new component
props as needed.

-->
<script lang="ts">
  import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";
  import WithTogglableFloatingElement from "../floating-element/WithTogglableFloatingElement.svelte";

  import { IconSpaceFixer } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { defaultChipColors } from "@rilldata/web-common/components/chip/chip-types";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import SearchableFilterDropdown from "./SearchableFilterDropdown.svelte";

  export let selectableItems: SearchableFilterSelectableItem[];
  export let selectedItems: boolean[];
  export let tooltipText: string;
  export let label: string;

  const dispatch = createEventDispatcher();

  $: {
    if (
      selectableItems?.length > 0 &&
      selectedItems?.length > 0 &&
      selectableItems?.length !== selectedItems?.length
    ) {
      throw new Error(
        "SearchableFilterButton component requires props `selectableItems` and `selectedItems` to be arrays of equal length",
      );
    }
  }
</script>

<WithTogglableFloatingElement
  alignment="start"
  distance={8}
  let:active
  let:toggleFloatingElement
>
  <Tooltip
    activeDelay={60}
    alignment="start"
    distance={8}
    location="bottom"
    suppress={active}
  >
    <!-- TODO: Switch to Measure colors once Theming supports it -->
    <Chip
      {...defaultChipColors}
      {active}
      extraRounded={false}
      {label}
      on:click={toggleFloatingElement}
      outline={true}
    >
      <div class="flex gap-x-2" slot="body">
        <div
          class="font-bold text-ellipsis overflow-hidden whitespace-nowrap ml-2"
        >
          {label}
        </div>

        <div class="flex items-center">
          <IconSpaceFixer pullRight>
            <div class="transition-transform" class:-rotate-180={active}>
              <CaretDownIcon size="14px" />
            </div>
          </IconSpaceFixer>
        </div>
      </div>
    </Chip>
    <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
      <TooltipContent maxWidth="400px">
        {tooltipText}
      </TooltipContent>
    </div>
  </Tooltip>
  <SearchableFilterDropdown
    allowMultiSelect={false}
    let:toggleFloatingElement
    on:apply
    on:click-outside={toggleFloatingElement}
    on:deselect-all
    on:escape={toggleFloatingElement}
    on:item-clicked={(e) => {
      toggleFloatingElement();
      dispatch("item-clicked", e.detail.name);
    }}
    on:search
    on:select-all
    selectableGroups={[{ name: "", items: selectableItems }]}
    selectedItems={[selectedItems]}
    slot="floating-element"
  />
</WithTogglableFloatingElement>
