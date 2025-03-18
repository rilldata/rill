<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    getDimensionDisplayName,
    getMeasureDisplayName,
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";
  import { PlusIcon } from "lucide-svelte";
  import { useMetricFieldData, type FieldType } from "./selectors";

  export let metricName: string;
  export let label: string;
  export let id: string;
  export let selectedItems: string[] = [];
  export let type: FieldType[];
  export let onMultiSelect: (items: string[]) => void = () => {};

  let open = false;
  let searchValue = "";
  // Local state for optimistic updates
  let localSelectedItems: string[] = selectedItems;

  const ctx = getCanvasStateManagers();

  const metricViewSpec =
    ctx.canvasEntity.spec.getMetricsViewFromName(metricName);

  $: fieldData = useMetricFieldData(ctx, metricName, type);

  $: selectableGroups = [
    ...(type.includes("measure")
      ? [
          <SearchableFilterSelectableGroup>{
            name: "MEASURES",
            items:
              $metricViewSpec?.measures?.map((m) => ({
                name: m.name as string,
                label: getMeasureDisplayName(m),
              })) ?? [],
          },
        ]
      : []),
    ...(type.includes("dimension")
      ? [
          <SearchableFilterSelectableGroup>{
            name: "DIMENSIONS",
            items:
              $metricViewSpec?.dimensions?.map((d) => ({
                name: (d.name || d.column) as string,
                label: getDimensionDisplayName(d),
              })) ?? [],
          },
        ]
      : []),
  ];

  $: {
    localSelectedItems = selectedItems;
  }

  function handleSelect(name: string) {
    const selectedProxy = new Set(localSelectedItems);
    if (selectedProxy.has(name)) {
      selectedProxy.delete(name);
    } else {
      selectedProxy.add(name);
    }

    localSelectedItems = Array.from(selectedProxy);
    onMultiSelect(localSelectedItems);
  }

  function handleRemove(item: string) {
    const selectedProxy = new Set(localSelectedItems);
    selectedProxy.delete(item);
    localSelectedItems = Array.from(selectedProxy);
    onMultiSelect(localSelectedItems);
  }
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={false}>
    <DropdownMenu.Trigger asChild let:builder>
      <div class="flex justify-between gap-x-2">
        <InputLabel small {label} {id} />
        <button use:builder.action {...builder} class="text-sm px-2 h-6">
          <PlusIcon size="14px" />
        </button>
      </div>
    </DropdownMenu.Trigger>

    <SearchableMenuContent
      {selectableGroups}
      selectedItems={[localSelectedItems]}
      allowMultiSelect={true}
      requireSelection
      searchText={searchValue}
      onSelect={handleSelect}
    />
  </DropdownMenu.Root>

  {#if selectedItems?.length > 0}
    <div class="flex flex-col gap-1 mt-2">
      {#each selectedItems as item}
        <Chip
          removable
          fullWidth
          type={$fieldData.displayMap[item]?.type ?? "dimension"}
          on:remove={() => handleRemove(item)}
        >
          <span class="font-bold truncate" slot="body">
            {$fieldData.displayMap[item]?.label || item}
          </span>
        </Chip>
      {/each}
    </div>
  {/if}
</div>
