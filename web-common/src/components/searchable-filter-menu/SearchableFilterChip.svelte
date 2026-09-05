<script lang="ts">
  import { fly } from "svelte/transition";
  import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import { Chip } from "@rilldata/web-common/components/chip";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import SearchableMenuContent from "./SearchableMenuContent.svelte";

  export let selectableItems: SearchableFilterSelectableItem[];
  export let selectedItems: string[];
  export let tooltipText: string;
  export let label: string;
  // Marks the chip's current selection as an ephemeral measure.
  export let fx = false;
  export let onSelect: (name: string) => void;
  // When set, ephemeral items in the menu get an edit button that closes the
  // menu and invokes this with the item name.
  export let onEditItem: ((name: string) => void) | undefined = undefined;

  let open = false;
  let searchText = "";
</script>

<DropdownMenu.Root
  bind:open
  onOpenChange={(open) => {
    if (!open) {
      searchText = "";
    }
  }}
>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Tooltip
        activeDelay={60}
        alignment="start"
        distance={8}
        location="bottom"
        suppress={open}
      >
        <Chip {...props} theme type="measure" active={open} {label}>
          <div slot="body" class="flex items-center gap-x-1 font-bold">
            <span class="truncate">{label}</span>
            {#if fx}
              <span class="flex-none text-[10px] font-semibold italic">ƒx</span>
            {/if}
          </div>
        </Chip>
        <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
          <TooltipContent maxWidth="400px">
            {tooltipText}
          </TooltipContent>
        </div>
      </Tooltip>
    {/snippet}
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    bind:searchText
    showSelection
    {onSelect}
    selectedItems={[selectedItems]}
    selectableGroups={[{ name: "", items: selectableItems }]}
    onEditItem={onEditItem
      ? (name) => {
          open = false;
          onEditItem?.(name);
        }
      : undefined}
  />
</DropdownMenu.Root>
