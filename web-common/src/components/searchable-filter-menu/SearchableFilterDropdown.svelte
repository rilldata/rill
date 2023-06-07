<script lang="ts">
  import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import Check from "../icons/Check.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../menu";
  import { Search } from "../search";
  import Footer from "./Footer.svelte";
  import Button from "../button/Button.svelte";
  import { createEventDispatcher } from "svelte";
  import { matchSorter } from "match-sorter";

  const dispatch = createEventDispatcher();

  export let selectableItems: SearchableFilterSelectableItem[];
  export let selectedItems: boolean[];

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

  $: numSelectedNotShown =
    numSelected - (menuItems?.filter((item) => item.selected)?.length || 0);

  $: allToggleText = numSelected === 0 ? "Select all" : "Deselect all";

  $: allToggleEvt = numSelected === 0 ? "select-all" : "deselect-all";
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
  <Search bind:value={searchText} />
  <!-- apply a wrapped flex element to ensure proper bottom spacing between body and footer -->
  <div class="flex flex-col flex-1 overflow-auto w-full pb-1">
    {#each menuItems as { name, label, selected, index }}
      <MenuItem
        icon
        animateSelect={false}
        focusOnMount={false}
        on:select={() => {
          dispatch("item-clicked", { index, name });
        }}
      >
        <svelte:fragment slot="icon">
          {#if selected}
            <Check size="20px" color="#15141A" />
          {:else}
            <Spacer size="20px" />
          {/if}
        </svelte:fragment>
        <span class:ui-copy-disabled={!selected}>
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
</Menu>
