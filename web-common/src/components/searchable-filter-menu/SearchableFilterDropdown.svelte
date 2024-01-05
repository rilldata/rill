<script lang="ts">
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import Divider from "@rilldata/web-common/components/menu/core/Divider.svelte";
  import { getMenuGroups } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import Check from "../icons/Check.svelte";
  import { Menu, MenuItem } from "../menu";
  import { Search } from "../search";
  import Footer from "./Footer.svelte";
  import Button from "../button/Button.svelte";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  export let selectableGroups: SearchableFilterSelectableGroup[];
  export let selectedItems: boolean[][];
  export let allowMultiSelect = true;
  export let showSelection = true;

  let searchText = "";

  $: menuGroups = getMenuGroups(selectableGroups, selectedItems, searchText);

  $: numSelected = selectedItems.reduce(
    (sel, items) => sel + items.filter((i) => i).length,
    0
  );

  $: singleSelection = numSelected === 1;

  $: numSelectedNotShown =
    numSelected -
    menuGroups.reduce(
      (sel, mg) => sel + mg.items.filter((i) => i.selected).length,
      0
    );

  $: selectableCount = selectableGroups.reduce(
    (sel, g) => sel + g.items.length,
    0
  );
  $: searchResultCount = menuGroups.reduce(
    (sel, mg) => sel + mg.items.length,
    0
  );

  $: allToggleText =
    numSelected === selectableCount ? "Deselect all" : "Select all";

  $: allToggleEvt =
    numSelected === selectableCount ? "deselect-all" : "select-all";

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
    {#if searchResultCount > 0}
      {#each menuGroups as { name, items, showDivider }}
        {#if items.length}
          {#if showDivider}
            <Divider />
          {/if}
          {#if name}
            <span class="gap-x-3 px-3 pb-1 text-gray-500 font-semibold"
              >{name}</span
            >
          {/if}
          {#each items as { name, label, selected, index }}
            <MenuItem
              icon={showSelection}
              animateSelect={false}
              focusOnMount={false}
              on:hover={() => {
                dispatch("hover", { index, name });
              }}
              on:focus={() => {
                dispatch("focus", { index, name });
              }}
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
                {:else}
                  <Spacer size="20px" />
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
          {/each}
        {/if}
      {/each}
    {:else}
      <div class="mt-5 ui-copy-disabled text-center">no results</div>
    {/if}
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
