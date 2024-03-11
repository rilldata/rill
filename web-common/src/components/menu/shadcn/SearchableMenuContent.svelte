<script lang="ts">
  import { getMenuGroups } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import Search from "../../search/Search.svelte";
  import Button from "../../button/Button.svelte";
  import { createEventDispatcher } from "svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  const dispatch = createEventDispatcher();

  export let selectableGroups: SearchableFilterSelectableGroup[];
  export let selectedItems: boolean[][];
  export let allowMultiSelect = true;

  let searchText = "";

  $: menuGroups = getMenuGroups(selectableGroups, selectedItems, searchText);

  $: numSelected = selectedItems.reduce(
    (sel, items) => sel + items.filter((i) => i).length,
    0,
  );

  $: singleSelection = numSelected === 1;

  $: numSelectedNotShown =
    numSelected -
    menuGroups.reduce(
      (sel, mg) => sel + mg.items.filter((i) => i.selected).length,
      0,
    );

  $: selectableCount = selectableGroups.reduce(
    (sel, g) => sel + g.items.length,
    0,
  );

  $: searchResultCount = menuGroups.reduce(
    (sel, mg) => sel + mg.items.length,
    0,
  );

  $: allToggleText =
    numSelected === selectableCount ? "Deselect all" : "Select all";

  $: allToggleEvt =
    numSelected === selectableCount ? "deselect-all" : "select-all";

  $: dispatchAllToggleEvt = () => {
    dispatch(allToggleEvt);
  };
</script>

<DropdownMenu.Content
  align="start"
  class="max-w-96 max-h-80 min-w-60 p-0 overflow-hidden flex flex-col"
>
  <div class="px-3 pt-3">
    <Search bind:value={searchText} showBorderOnFocus={false} />
  </div>

  <div class="overflow-y-scroll overflow-x-hidden size-full py-1">
    {#if searchResultCount > 0}
      {#each menuGroups as { name, items, showDivider }}
        {#if items.length}
          {#if showDivider}
            <DropdownMenu.Separator />
          {/if}
          {#if name}
            <span class="gap-x-3 px-3 pb-1 text-gray-500 font-semibold">
              {name}
            </span>
          {/if}

          <DropdownMenu.Group class="px-1">
            {#each items as { name, label, selected, index }}
              <svelte:component
                this={allowMultiSelect
                  ? DropdownMenu.CheckboxItem
                  : DropdownMenu.Item}
                {...allowMultiSelect ? { checked: selected } : {}}
                class="text-xs cursor-pointer"
                role="menuitem"
                bind:checked={selected}
                disabled={singleSelection && selected}
                on:click={() => {
                  if (singleSelection && selected) return;
                  dispatch("item-clicked", { index, name });
                }}
              >
                <span class:ui-copy-disabled={!selected && allowMultiSelect}>
                  {#if label.length > 240}
                    {label.slice(0, 240)}...
                  {:else}
                    {label}
                  {/if}
                </span>
              </svelte:component>
            {/each}
          </DropdownMenu.Group>
        {/if}
      {/each}
    {:else}
      <div class="ui-copy-disabled text-center p-2 w-full">no results</div>
    {/if}
  </div>

  {#if allowMultiSelect}
    <DropdownMenu.Group
      class="flex items-center justify-between pl-2 pr-4 py-1 border-t bg-gray-50 dark:bg-gray-600 border-gray-200 dark:border-gray-500"
    >
      <Button compact on:click={dispatchAllToggleEvt} type="text">
        {allToggleText}
      </Button>
      {#if numSelectedNotShown}
        <div class="ui-label">
          {numSelectedNotShown} other value{numSelectedNotShown > 1 ? "s" : ""} selected
        </div>
      {/if}
    </DropdownMenu.Group>
  {/if}
</DropdownMenu.Content>
