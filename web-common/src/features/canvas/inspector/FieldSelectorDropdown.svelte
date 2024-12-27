<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import {
    getDimensionDisplayName,
    getMeasureDisplayName,
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { MinusIcon, PlusIcon } from "lucide-svelte";
  import {
    useAllDimensionFromMetric,
    useAllSimpleMeasureFromMetric,
  } from "./selectors";

  export let metricName: string;
  export let label: string | undefined = undefined;
  export let id: string;
  export let selectedItem: string | undefined = undefined;
  export let type: "measure" | "dimension";
  export let searchableItems: string[] | undefined = undefined;
  export let onSelect: (item: string) => void = () => {};
  export let onMultiSelect: (items: string[]) => void = () => {};
  export let multi = false;
  export let selectedItems: string[] | undefined = undefined;

  $: allDimensions = useAllDimensionFromMetric($runtime.instanceId, metricName);
  $: allFilteredMeasures = useAllSimpleMeasureFromMetric(
    $runtime.instanceId,
    metricName,
  );

  $: items =
    type === "measure"
      ? ($allFilteredMeasures?.data?.map((m) => m.name as string) ?? [])
      : ($allDimensions?.data?.map((d) => d.name || (d.column as string)) ??
        []);

  $: displayMap =
    type === "measure"
      ? Object.fromEntries(
          $allFilteredMeasures?.data?.map((m) => [
            m.name,
            getMeasureDisplayName(m),
          ]) ?? [],
        )
      : Object.fromEntries(
          $allDimensions?.data?.map((d) => [
            d.name || d.column,
            getDimensionDisplayName(d),
          ]) ?? [],
        );

  let open = false;
  let searchValue = "";

  // This Set keeps track of the “in-progress” selection while dropdown is open.
  let selectedProxy = new Set(multi ? selectedItems : selectedItem);

  $: filteredItems = (
    searchableItems && searchValue ? searchableItems : items
  ).filter((item) => {
    const matchesSearch =
      displayMap[item]?.toLowerCase().includes(searchValue.toLowerCase()) ||
      item.toLowerCase().includes(searchValue.toLowerCase());
    // For single-select, remove the already selected item from the list
    if (!multi && selectedItem && item === selectedItem) return false;
    return matchesSearch;
  });
</script>

<div class="flex flex-col gap-y-2 pt-1">
  {#if !multi && label}
    <div class="flex items-center gap-x-2">
      <InputLabel small {label} {id} />
    </div>
  {/if}

  {#if multi}
    <DropdownMenu.Root
      bind:open
      typeahead={false}
      closeOnItemClick={true}
      onOpenChange={() => {
        // Reset our proxy whenever the menu opens (or closes)
        if (!open) {
          selectedProxy = new Set(selectedItems);
        }
      }}
    >
      <DropdownMenu.Trigger asChild let:builder>
        <div class="flex justify-between gap-x-2">
          {#if label}
            <InputLabel small {label} {id} />
          {/if}
          <button use:builder.action {...builder} class="text-sm px-2 h-6">
            <PlusIcon size="14px" />
          </button>
        </div>
      </DropdownMenu.Trigger>

      <DropdownMenu.Content class="p-0 w-[300px]">
        <div class="p-3 pb-1">
          <Search bind:value={searchValue} autofocus={false} />
        </div>

        <div class="max-h-64 overflow-y-auto">
          {#each filteredItems as item (item)}
            <DropdownMenu.CheckboxItem
              checked={selectedProxy.has(item)}
              class="pl-8 mx-1"
              on:click={() => {
                if (selectedProxy.has(item)) {
                  selectedProxy.delete(item);
                } else {
                  selectedProxy.add(item);
                }
                selectedProxy = selectedProxy;
                onMultiSelect(Array.from(selectedProxy));
              }}
            >
              <slot {item}>
                {displayMap[item] || item}
              </slot>
            </DropdownMenu.CheckboxItem>
          {:else}
            {#if searchValue}
              <div class="ui-copy-disabled text-center p-2 w-full">
                no results
              </div>
            {/if}
          {/each}
        </div>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {:else}
    <DropdownMenu.Root
      bind:open
      typeahead={false}
      closeOnItemClick={false}
      onOpenChange={() => {
        if (!open) {
          selectedProxy = new Set(selectedItem);
        }
      }}
    >
      <DropdownMenu.Trigger asChild let:builder>
        <Chip fullWidth caret {type} builders={[builder]}>
          <span class="font-bold truncate" slot="body">
            {#if selectedItem}
              {displayMap[selectedItem] || selectedItem}
            {:else}
              Select a {type} field
            {/if}
          </span>
        </Chip>
      </DropdownMenu.Trigger>

      <DropdownMenu.Content sameWidth class="p-0">
        <div class="p-3 pb-1">
          <Search bind:value={searchValue} autofocus={false} />
        </div>
        <div class="max-h-64 overflow-y-auto">
          {#each filteredItems as item (item)}
            <DropdownMenu.Item
              class="pl-8 mx-1"
              on:click={() => {
                onSelect(item);
                open = false;
              }}
            >
              <slot {item}>
                {displayMap[item] || item}
              </slot>
            </DropdownMenu.Item>
          {:else}
            {#if searchValue}
              <div class="ui-copy-disabled text-center p-2 w-full">
                no results
              </div>
            {/if}
          {/each}
        </div>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}

  {#if multi && selectedItems && selectedItems.length > 0}
    <div class="flex flex-col gap-1">
      {#each selectedItems as item}
        <div class="flex items-center justify-between gap-x-2">
          <div class="flex-1">
            <Chip fullWidth {type}>
              <span class="font-bold truncate" slot="body">
                {displayMap[item] || item}
              </span>
            </Chip>
          </div>
          <button
            class=" px-2 py-1 text-xs"
            on:click={() => {
              selectedProxy.delete(item);
              onMultiSelect(Array.from(selectedProxy));
            }}
          >
            <MinusIcon size="14px" />
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style lang="postcss">
  .open {
    @apply ring-2 ring-primary-100 border-primary-600;
  }
</style>
