<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import { matchSorter } from "match-sorter";
  import { PencilIcon } from "lucide-svelte";
  import Button from "../button/Button.svelte";
  import { Search } from "../search";

  const voidFn = () => {};

  export let selectableGroups: SearchableFilterSelectableGroup[];
  export let selectedItems: string[][];
  export let allowMultiSelect = false;
  export let requireSelection = false;
  export let fadeUnselected = false;
  export let showXForSelected = false;
  export let showSelection = false;
  export let allowSelectAll = true;
  export let searchText = "";
  export let showHiddenSelectionsCount = false;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let onSelect: (name: string) => void;
  export let onToggleSelectAll: () => void = voidFn;
  // When set, items flagged `ephemeral` get an edit button invoking this.
  export let onEditItem: ((name: string) => void) | undefined = undefined;

  $: allSelected = selectableGroups.every((g, i) => {
    return (
      selectedItems[i]?.length && g.items.length === selectedItems[i].length
    );
  });

  $: selectedCount = selectedItems.flat().length;

  $: singleSelection = selectedCount === 1;

  $: numSelectedShown = selectableGroups.reduce(
    (sel, { items }, groupIndex) => {
      return (
        sel +
        items.filter(({ name }) => selectedItems[groupIndex]?.includes(name))
          .length
      );
    },
    0,
  );

  $: numSelectedNotShown = selectedCount - numSelectedShown;

  $: filteredGroups = filterGroups(selectableGroups, searchText);

  function filterGroups(
    selectableGroups: SearchableFilterSelectableGroup[],

    searchText: string,
  ) {
    if (!searchText) return selectableGroups;

    return selectableGroups.map((group) => {
      return {
        ...group,
        items: matchSorter(group.items, searchText, { keys: ["label"] }),
      };
    });
  }
</script>

<DropdownMenu.Content
  align="start"
  {side}
  class="flex flex-col max-h-96 w-72 overflow-hidden p-0"
>
  <div class="px-3 pt-3 pb-1">
    <Search
      bind:value={searchText}
      label={m.common_search_list()}
      showBorderOnFocus={false}
    />
  </div>

  <div class="flex flex-col flex-1 overflow-y-auto w-full h-fit pb-1">
    {#each filteredGroups as { name, label, items }, index (name)}
      <DropdownMenu.Group class="px-1">
        {#if filteredGroups.length > 1}
          <DropdownMenu.Label>
            {label ?? name}
          </DropdownMenu.Label>
        {/if}
        {#each items as { name, label, description, ephemeral } (name)}
          {@const selected = selectedItems[index]?.includes(name)}

          {#if allowMultiSelect || showSelection}
            <DropdownMenu.CheckboxItem
              checked={selected}
              {showXForSelected}
              closeOnSelect={!allowMultiSelect}
              class="text-xs cursor-pointer"
              disabled={requireSelection && singleSelection && selected}
              aria-disabled={requireSelection && singleSelection && selected}
              onclick={() => {
                if (requireSelection && singleSelection && selected) return;

                // Deferred so the checkbox's own toggle handler (which runs
                // after this chained handler) cannot overwrite the state
                // derived from the selection update. See the item state
                // getting stuck when the update flushes synchronously.
                queueMicrotask(() => onSelect(name));
              }}
            >
              <span
                class="inline-flex items-center gap-x-1"
                class:text-fg-disabled={fadeUnselected &&
                  !selected &&
                  allowMultiSelect}
                title={description}
              >
                {#if label.length > 240}
                  {label.slice(0, 240)}...
                {:else}
                  {label}
                {/if}
                {#if ephemeral}
                  <span class="text-[10px] font-semibold italic">ƒx</span>
                {/if}
              </span>
              {#if ephemeral && onEditItem}
                <button
                  class="ml-auto flex-none text-fg-secondary hover:text-fg-primary"
                  type="button"
                  aria-label={m.dashboard_pivot_ephemeral_edit_title()}
                  onpointerdown={(e) => e.stopPropagation()}
                  onpointerup={(e) => e.stopPropagation()}
                  onclick={(e) => {
                    e.stopPropagation();
                    onEditItem?.(name);
                  }}
                >
                  <PencilIcon size="12px" />
                </button>
              {/if}
            </DropdownMenu.CheckboxItem>
          {:else}
            <DropdownMenu.Item
              class="text-xs cursor-pointer"
              disabled={requireSelection && singleSelection && selected}
              aria-disabled={requireSelection && singleSelection && selected}
              onclick={() => {
                if (requireSelection && singleSelection && selected) return;

                onSelect(name);
              }}
            >
              <span
                class="inline-flex items-center gap-x-1"
                title={description}
              >
                {#if label.length > 240}
                  {label.slice(0, 240)}...
                {:else}
                  {label}
                {/if}
                {#if ephemeral}
                  <span class="text-[10px] font-semibold italic">ƒx</span>
                {/if}
              </span>
              {#if ephemeral && onEditItem}
                <button
                  class="ml-auto flex-none text-fg-secondary hover:text-fg-primary"
                  type="button"
                  aria-label={m.dashboard_pivot_ephemeral_edit_title()}
                  onpointerdown={(e) => e.stopPropagation()}
                  onpointerup={(e) => e.stopPropagation()}
                  onclick={(e) => {
                    e.stopPropagation();
                    onEditItem?.(name);
                  }}
                >
                  <PencilIcon size="12px" />
                </button>
              {/if}
            </DropdownMenu.Item>
          {/if}
        {:else}
          <div
            data-testid="searchable-menu-no-results"
            class="text-fg-disabled text-center p-2 w-full"
          >
            {m.common_no_results()}
          </div>
        {/each}

        {#if index !== filteredGroups.length - 1}
          <DropdownMenu.Separator />
        {/if}
      </DropdownMenu.Group>
    {/each}
  </div>

  {#if (allowSelectAll && allowMultiSelect) || $$slots.action}
    <footer>
      {#if allowSelectAll && allowMultiSelect}
        <Button onClick={onToggleSelectAll} type="tertiary">
          {#if allSelected}
            {m.common_deselect_all()}
          {:else}
            {m.common_select_all()}
          {/if}
        </Button>
      {/if}

      <slot name="action" />
      {#if allowSelectAll && allowMultiSelect && numSelectedNotShown && showHiddenSelectionsCount}
        <div class="ui-label">
          {m.common_other_values_selected({ count: numSelectedNotShown })}
        </div>
      {/if}
    </footer>
  {/if}
</DropdownMenu.Content>

<style lang="postcss">
  footer {
    height: 42px;
    @apply border-t border-gray-300;
    @apply bg-gray-100;
    @apply flex flex-row flex-none items-center justify-end;
    @apply gap-x-2 p-2 px-3.5;
  }
</style>
