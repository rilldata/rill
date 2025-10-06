<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import { matchSorter } from "match-sorter";
  import Button from "../button/Button.svelte";
  import { Search } from "../search";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { builderActions, getAttrs } from "bits-ui";

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
        items.filter(({ id }) => selectedItems[groupIndex]?.includes(id)).length
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
        items: matchSorter(group.items, searchText, {
          keys: ["labels.0", "id"],
        }),
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
    <Search bind:value={searchText} label="Search list" />
  </div>

  <div class="flex flex-col flex-1 overflow-y-auto w-full h-fit pb-1">
    {#each filteredGroups as { name, label, items }, index (name)}
      <DropdownMenu.Group class="px-1">
        {#if filteredGroups.length > 1}
          <DropdownMenu.Label>
            {label ?? name}
          </DropdownMenu.Label>
        {/if}
        {#each items as { id, labels, tooltip } (id)}
          {@const selected = selectedItems[index]?.includes(name)}

          <svelte:component
            this={allowMultiSelect || showSelection
              ? DropdownMenu.CheckboxItem
              : DropdownMenu.Item}
            {...allowMultiSelect || showSelection
              ? { checked: selected, showXForSelected }
              : {}}
            class="cursor-pointer"
            role="menuitem"
            disabled={requireSelection && singleSelection && selected}
            aria-disabled={requireSelection && singleSelection && selected}
            on:click={() => {
              if (requireSelection && singleSelection && selected) return;

              onSelect(id);
            }}
          >
            <Tooltip.Root portal="body" openDelay={750}>
              <Tooltip.Trigger asChild let:builder>
                <span
                  use:builderActions={{ builders: [builder] }}
                  {...getAttrs([builder])}
                  class="line-clamp-1 truncate w-full"
                  class:ui-copy-disabled={fadeUnselected &&
                    !selected &&
                    allowMultiSelect}
                >
                  {#if labels.length > 1}
                    {id}
                    <span class="ui-copy-disabled font-normal">
                      {` (${labels.join(", ")})`}</span
                    >
                  {:else}
                    {@const label = Array.from(labels)[0]}
                    {#if label.length > 240}
                      {label.slice(0, 240)}...
                    {:else}
                      {label}
                    {/if}
                  {/if}
                </span>
              </Tooltip.Trigger>

              {#if tooltip}
                <Tooltip.Content
                  side="right"
                  sideOffset={20}
                  class="bg-neutral-50 border text-neutral-700 text-[13px] pr-4 font-medium shadow-md"
                >
                  <h3>Metrics Views</h3>

                  <ul class="">
                    {#each tooltip?.split(", ") || [] as line, i (i)}
                      <li class="flex flex-row items-center gap-x-1.5">
                        <!-- <span class="size-1 rounded-full bg-neutral-600"></span> -->
                        <a href="/files/metrics/{line}.yaml">{line}</a>
                      </li>
                    {/each}
                  </ul>
                </Tooltip.Content>
              {/if}
            </Tooltip.Root>
          </svelte:component>
        {:else}
          <div
            data-testid="searchable-menu-no-results"
            class="ui-copy-disabled text-center p-2 w-full"
          >
            no results
          </div>
        {/each}

        {#if index !== filteredGroups.length - 1}
          <DropdownMenu.Separator />
        {/if}
      </DropdownMenu.Group>
    {/each}
  </div>

  {#if allowSelectAll && allowMultiSelect}
    <footer>
      <Button onClick={onToggleSelectAll} type="plain">
        {#if allSelected}
          Deselect all
        {:else}
          Select all
        {/if}
      </Button>

      <slot name="action" />
      {#if numSelectedNotShown && showHiddenSelectionsCount}
        <div class="ui-label">
          {numSelectedNotShown} other value{numSelectedNotShown > 1 ? "s" : ""} selected
        </div>
      {/if}
    </footer>
  {/if}
</DropdownMenu.Content>

<style lang="postcss">
  footer {
    height: 42px;
    @apply border-t border-slate-300;
    @apply bg-slate-100;
    @apply flex flex-row flex-none items-center justify-end;
    @apply gap-x-2 p-2 px-3.5;
  }

  footer:is(.dark) {
    @apply bg-gray-800;
    @apply border-gray-700;
  }
</style>
