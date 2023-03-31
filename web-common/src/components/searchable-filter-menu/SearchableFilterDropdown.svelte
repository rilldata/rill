<script lang="ts">
  import Check from "../icons/Check.svelte";
  import Spacer from "../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../menu";
  import { Search } from "../search";
  import Footer from "./Footer.svelte";
  import Button from "../button/Button.svelte";
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let selectableItems: string[];
  export let selectedItems: boolean[];

  export const getNumSelectedNotShown = (
    selectedItems: boolean[],
    visibleInSearch: boolean[]
  ): number =>
    selectedItems?.filter((s, i) => s && !visibleInSearch[i])?.length || 0;

  export const setItemsVisibleBySearchString = (
    items: string[],
    searchText: string
  ): boolean[] => {
    return items?.map((x) => x.includes(searchText.trim()));
  };

  let searchText = "";

  let visibleInSearch: boolean[];
  $: visibleInSearch = setItemsVisibleBySearchString(
    selectableItems,
    searchText
  );

  const deselectAll = () => {
    dispatch("deselectAll");
    // selectedItems = selectedItems.map((_) => false);
  };

  $: numSelectedNotShown = getNumSelectedNotShown(
    selectedItems,
    visibleInSearch
  );

  $: dispatch("selectedItemsChanged", selectedItems);

  $: menuItems = selectableItems
    .map((item, i) => ({
      label: item,
      visible: visibleInSearch[i],
      selected: selectedItems[i],
      index: i,
    }))
    .filter((item) => item.visible);

  // const updateSelectedItemsByIndex = (index) => {
  //   selectedItems = selectedItems.map((x, i) => (i === index ? !x : x));
  // };
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
    {#each menuItems as { label, visible, selected, index } (label + index)}
      <MenuItem
        icon
        animateSelect={false}
        focusOnMount={false}
        on:select={() => {
          // selectedItems[index] = !selectedItems[index];
          // updateSelectedItemsByIndex(index);
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
      <Button type="text" compact on:click={deselectAll}>Deselect all</Button>
    </span>
    {#if numSelectedNotShown}
      <div class="ui-label">
        {numSelectedNotShown} other value{numSelectedNotShown > 1 ? "s" : ""} selected
      </div>
    {/if}
  </Footer>
</Menu>
