<script lang="ts">
  import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import Check from "../icons/Check.svelte";
  import { Menu, MenuItem } from "../menu";
  import { Search } from "../search";
  import Footer from "./Footer.svelte";
  import Button from "../button/Button.svelte";
  import { createEventDispatcher } from "svelte";
  import { matchSorter } from "match-sorter";

  const dispatch = createEventDispatcher();

  export let selectableItems: SearchableFilterSelectableItem[];
  export let selectedItems: boolean[];
  export let allowMultiSelect = true;

  interface MenuItemData {
    name: string;
    label: string;
    selected: boolean;
    index: number;
  }

  export const setItemsVisibleBySearchString = (
    items: SearchableFilterSelectableItem[],
    selected: boolean[],
    searchText: string
  ): MenuItemData[] => {
    let menuEntries = items.map((item, i) => ({
      name: item.name,
      label: item.label,
      selected: selected[i],
      index: i,
    }));
    // if there is no search text, return menuEntries right away,
    // otherwise matchSorter sorts the mentu entries
    if (!searchText) return menuEntries;
    return matchSorter(menuEntries, searchText, { keys: ["label"] });
  };

  let searchText = "";

  $: menuItems = setItemsVisibleBySearchString(
    selectableItems,
    selectedItems,
    searchText
  );

  $: numSelected = selectedItems?.filter((s) => s)?.length || 0;

  $: singleSelection = numSelected === 1;

  $: numSelectedNotShown =
    numSelected - (menuItems?.filter((item) => item.selected)?.length || 0);

  $: allToggleText =
    numSelected === selectableItems?.length ? "Deselect all" : "Select all";

  $: allToggleEvt =
    numSelected === selectableItems?.length ? "deselect-all" : "select-all";
  $: dispatchAllToggleEvt = () => {
    dispatch(allToggleEvt);
  };
</script>

<Menu
  focusOnMount={false}
  maxHeight="400px"
  maxWidth="480px"
  minHeight="150px"
  on:click-outside
  on:escape
  paddingBottom={0}
  paddingTop={1}
  rounded={false}
>
  <!-- the min-height is set to have about 3 entries in it -->
  <div class="px-3 py-2">
    <Search bind:value={searchText} showBorderOnFocus={false} />
  </div>
  <!-- apply a wrapped flex element to ensure proper bottom spacing between body and footer -->
  <div class="flex flex-col flex-1 overflow-auto w-full pb-1">
    {#each menuItems as { name, label, selected, index }}
      <MenuItem
        icon
        animateSelect={false}
        focusOnMount={false}
        on:select={() => {
          if (singleSelection && selected) return;
          dispatch("item-clicked", { index, name });
        }}
      >
        <svelte:fragment slot="icon">
          {#if selected}
            <Check
              size="20px"
              color={allowMultiSelect && singleSelection
                ? "#9CA3AF"
                : "#15141A"}
            />
          {/if}
        </svelte:fragment>
        <span class:ui-copy-disabled={!selected && allowMultiSelect}>
          {#if label.length > 240}
            {label.slice(0, 240)}...
          {:else}
            {label}
          {/if}
        </span>
      </MenuItem>
    {:else}
      <div class="mt-5 ui-copy-disabled text-center">no results</div>
    {/each}
  </div>
  {#if allowMultiSelect}
    <Footer>
      <span class="ui-copy">
        <Button compact on:click={dispatchAllToggleEvt} type="text"
          >{allToggleText}</Button
        >
      </span>
      {#if numSelectedNotShown}
        <div class="ui-label">
          {numSelectedNotShown} other value{numSelectedNotShown > 1 ? "s" : ""} selected
        </div>
      {/if}
    </Footer>
  {/if}
</Menu>
