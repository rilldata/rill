<script lang="ts">
  // import { createEventDispatcher } from "svelte";
  // import { Switch } from "../button";
  // import Cancel from "../icons/Cancel.svelte";
  import Check from "../icons/Check.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../menu";
  import { Search } from "../search";
  import Footer from "./Footer.svelte";
  import Button from "../button/Button.svelte";
  import {
    getNumSelectedNotShown,
    // SelectableItem,
    setItemsVisibleBySearchString,
  } from "./types-and-utils";

  export let selectableItems: string[];
  export let selectedItems: boolean[];
  // let visibleInSearch = selectableItems.map((_) => true);

  // $: selectedItems = selectableItems.map((x) => x.selected);

  let searchText = "";

  $: {
    console.log("selectedItems Inner", selectedItems);
  }

  $: visibleInSearch = setItemsVisibleBySearchString(
    selectableItems,
    searchText
  );

  const deselectAll = () => {
    selectedItems = selectedItems.map((_) => false);
    // selectableItems = selectableItems.map((x) => ({ ...x, selected: false }));
  };

  $: numSelectedNotShown = getNumSelectedNotShown(
    selectableItems,
    visibleInSearch
  );
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
    {#each selectableItems.filter((_, i) => visibleInSearch[i]) as selectableItem, i}
      <MenuItem
        icon
        animateSelect={false}
        focusOnMount={false}
        on:select={() => {
          selectedItems[i] = !selectedItems[i];
        }}
      >
        <svelte:fragment slot="icon">
          {#if selectedItems[i]}
            <Check size="20px" color="#15141A" />
          {:else}
            <Spacer size="20px" />
          {/if}
        </svelte:fragment>
        <span class:ui-copy-disabled={!selectedItems[i]}>
          {#if selectableItem.length > 240}
            {selectableItem.slice(0, 240)}...
          {:else}
            {selectableItem}
          {/if}
        </span>
      </MenuItem>
    {:else}
      <div class="mt-5 ui-copy-disabled text-center">no results</div>
    {/each}
  </div>
  <Footer>
    <span class="ui-copy">
      <Button type="text" compact on:click={deselectAll}>Deselect all</Button>
    </span>
    {#if numSelectedNotShown}
      <div class="ui-label">
        {numSelectedNotShown} other value{numSelectedNotShown > 1 ? "s" : ""} selected
      </div>
    {/if}
  </Footer>
</Menu>
