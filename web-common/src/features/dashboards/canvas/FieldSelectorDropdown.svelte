<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import {
    getDimensionDisplayName,
    getMeasureDisplayName,
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import { getStateManagers } from "../state-managers/state-managers";

  export let label: string;
  export let id: string;
  export let selectedItem: string | undefined;
  export let type: "measure" | "dimension";
  export let searchableItems: string[] | undefined = undefined;
  export let onSelect: (item: string) => void;

  const {
    selectors: {
      dimensions: { allDimensions },
      measures: { filteredSimpleMeasures },
    },
  } = getStateManagers();

  $: items =
    type === "measure"
      ? ($filteredSimpleMeasures()?.map((m) => m.name as string) ?? [])
      : ($allDimensions?.map((d) => d.name || (d.column as string)) ?? []);

  $: displayMap =
    type === "measure"
      ? Object.fromEntries(
          $filteredSimpleMeasures()?.map((m) => [
            m.name,
            getMeasureDisplayName(m),
          ]) ?? [],
        )
      : Object.fromEntries(
          $allDimensions?.map((d) => [
            d.name || d.column,
            getDimensionDisplayName(d),
          ]) ?? [],
        );

  let open = false;
  let searchValue = "";
  let selectedProxy = selectedItem ? new Set([selectedItem]) : new Set();

  $: filteredItems = (
    searchableItems && searchValue ? searchableItems : items
  ).filter((item) => {
    return (
      (!selectedItem || item !== selectedItem) &&
      (displayMap[item]?.toLowerCase().includes(searchValue.toLowerCase()) ||
        item.toLowerCase().includes(searchValue.toLowerCase()))
    );
  });
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <InputLabel
    {label}
    {id}
    hint="Selection of a {type} from the underlying metrics view for inclusion on the dashboard"
  />
  <DropdownMenu.Root
    bind:open
    typeahead={false}
    closeOnItemClick={false}
    onOpenChange={() => {
      if (!open) {
        selectedProxy = selectedItem ? new Set([selectedItem]) : new Set();
      }
    }}
  >
    <DropdownMenu.Trigger asChild let:builder>
      <button
        use:builder.action
        {...builder}
        class:open
        class="flex px-3 gap-x-2 h-8 max-w-full items-center text-sm border-gray-300 border rounded-[2px] break-all overflow-hidden"
      >
        {#if selectedItem}
          {displayMap[selectedItem] || selectedItem}
        {:else}
          Select a {type} field
        {/if}

        <CaretDownIcon
          size="12px"
          className="!fill-gray-600 ml-auto flex-none"
        />
      </button>
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
</div>

<style lang="postcss">
  .open {
    @apply ring-2 ring-primary-100 border-primary-600;
  }
</style>
