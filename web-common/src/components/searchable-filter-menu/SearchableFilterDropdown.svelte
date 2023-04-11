<script lang="ts">
  import Check from "../icons/Check.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../menu";
  import { Search } from "../search";
  import Footer from "./Footer.svelte";
  import Button from "../button/Button.svelte";
  import { createEventDispatcher } from "svelte";
  import { matchSorter } from "match-sorter";

  const dispatch = createEventDispatcher();

  export let selectableItems: string[];
  export let selectedItems: boolean[];

  interface MenuItemData {
    label: string;
    selected: boolean;
    index: number;
  }

  export const setItemsVisibleBySearchString = (
    items: string[],
    selected: boolean[],
    searchText: string
  ): MenuItemData[] => {
    let menuEntries = items.map((item, i) => ({
      label: item,
      selected: selected[i],
      index: i,
    }));
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

  $: allToggleEvt = numSelected === 0 ? "selectAll" : "deselectAll";
  $: dispatchAllToggleEvt = () => {
    dispatch(allToggleEvt);
  };
</script>

<Menu
  paddingTop={1}
  paddingBottom={0}
  rounded={false}
  focusOnMount={false}
  maxWidth="480px"
  minHeight="150px"
  maxHeight="400px"
  on:escape
  on:click-outside
>
  <!-- the min-height is set to have about 3 entries in it -->
  <Search bind:value={searchText} />
  <!-- apply a wrapped flex element to ensure proper bottom spacing between body and footer -->
  <div class="flex flex-col flex-1 overflow-auto w-full pb-1">
    {#each menuItems as { label, selected, index }}
      <MenuItem
        icon
        animateSelect={false}
        focusOnMount={false}
        on:select={() => {
          dispatch("itemClicked", index);
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
      <Button type="text" compact on:click={dispatchAllToggleEvt}
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
