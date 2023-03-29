<script lang="ts">
  // import { createEventDispatcher } from "svelte";
  // import { Switch } from "../button";
  // import Cancel from "../icons/Cancel.svelte";
  import Check from "../icons/Check.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../menu";
  import { Search } from "../search";
  import Footer from "./Footer.svelte";
  import {
    getNumSelectedNotShown,
    SelectableItem,
    setItemsVisibleBySearchString,
  } from "./types-and-utils";

  export let selectableItems: SelectableItem[];

  export let selectableItemsOutput: SelectableItem[];
  // export let searchedValues: string[] = [];

  $: selectableItemsOutput = [...selectableItems];

  let searchText = "";

  $: {
    console.log("selectableItemsInner", selectableItems);
  }

  $: {
    selectableItems = setItemsVisibleBySearchString(
      selectableItems,
      searchText
    );
  }

  /** On instantiation, only take the exact current selectedValues, so that
   * when the user unchecks a menu item, it still persists in the FilterMenu
   * until the user closes.
   */
  // let candidateValues = selectableItems.;
  // let valuesToDisplay = [...candidateValues];

  // $: if (searchText) {
  //   valuesToDisplay = [...searchedValues];
  // } else valuesToDisplay = [...candidateValues];

  // $: numSelectedNotInSearch = selectedValues.filter(
  //   (v) => !valuesToDisplay.includes(v)
  // ).length;

  $: numSelectedNotShown = getNumSelectedNotShown(selectableItems);

  // function toggleValue(value) {
  //   dispatch("apply", value);

  //   if (!candidateValues.includes(value)) {
  //     candidateValues = [...candidateValues, value];
  //   }
  // }
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
    {#if selectableItems?.length}
      {#each selectableItems.filter((x) => x.visibleInMenu) as selectableItem}
        <MenuItem
          icon
          animateSelect={false}
          focusOnMount={false}
          on:select={() => {
            selectableItem.selected = !selectableItem.selected;
          }}
        >
          <svelte:fragment slot="icon">
            {#if selectableItem.selected}
              <Check size="20px" color="#15141A" />
            {:else}
              <Spacer size="20px" />
            {/if}
          </svelte:fragment>
          <span class:ui-copy-disabled={!selectableItem.selected}>
            {#if selectableItem.label?.length > 240}
              {selectableItem.labelselectableItem.label.slice(0, 240)}...
            {:else}
              {selectableItem.label}
            {/if}
          </span>
        </MenuItem>
      {/each}
    {:else}
      <div class="mt-5 ui-copy-disabled text-center">no results</div>
    {/if}
  </div>
  <Footer>
    {#if numSelectedNotShown}
      <div class="ui-label">
        {numSelectedNotShown} other value{numSelectedNotShown > 1 ? "s" : ""} selected
      </div>
    {/if}
  </Footer>
</Menu>
