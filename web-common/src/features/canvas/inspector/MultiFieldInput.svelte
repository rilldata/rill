<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { PlusIcon } from "lucide-svelte";
  import ChipDragList from "./ChipDragList.svelte";
  import { useMetricFieldData } from "./selectors";
  import type { FieldType } from "./types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let canvasName: string;
  export let metricName: string;
  export let label: string;
  export let id: string;
  export let selectedItems: string[] = [];
  export let types: FieldType[];
  export let onMultiSelect: (items: string[]) => void = () => {};

  let open = false;
  let searchValue = "";
  // Local state for optimistic updates
  let localSelectedItems: string[] = selectedItems;

  $: ({ instanceId } = $runtime);

  $: ctx = getCanvasStore(canvasName, instanceId);
  $: fieldData = useMetricFieldData(ctx, metricName, types);
  $: selectableGroups = [
    ...(types.includes("measure")
      ? [
          <SearchableFilterSelectableGroup>{
            name: "measure",
            label: "MEASURES",
            items: $fieldData.items
              .filter((item) => $fieldData.displayMap[item]?.type === "measure")
              .map((item) => ({
                name: item,
                label: $fieldData.displayMap[item].label,
              })),
          },
        ]
      : []),
    ...(types.includes("time")
      ? [
          <SearchableFilterSelectableGroup>{
            name: "time",
            label: "TIME",
            items: $fieldData.items
              .filter((item) => $fieldData.displayMap[item]?.type === "time")
              .map((item) => ({
                name: item,
                label: $fieldData.displayMap[item].label,
              })),
          },
        ]
      : []),
    ...(types.includes("dimension")
      ? [
          <SearchableFilterSelectableGroup>{
            name: "dimension",
            label: "DIMENSIONS",
            items: $fieldData.items
              .filter(
                (item) => $fieldData.displayMap[item]?.type === "dimension",
              )
              .map((item) => ({
                name: item,
                label: $fieldData.displayMap[item].label,
              })),
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
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={false}>
    <DropdownMenu.Trigger asChild let:builder>
      <div class="flex justify-between gap-x-2">
        <InputLabel small {label} {id} />
        <button
          aria-label={`Add ${types.join(", ")} fields`}
          use:builder.action
          {...builder}
          class="text-sm px-2 h-6"
        >
          <PlusIcon size="14px" />
        </button>
      </div>
    </DropdownMenu.Trigger>

    <SearchableMenuContent
      {selectableGroups}
      selectedItems={selectableGroups.map((group) =>
        localSelectedItems?.filter(
          (item) => $fieldData.displayMap[item]?.type === group.name,
        ),
      )}
      allowMultiSelect={true}
      searchText={searchValue}
      allowSelectAll={false}
      onSelect={handleSelect}
    />
  </DropdownMenu.Root>

  {#if selectedItems?.length > 0}
    <div class="mt-2">
      <ChipDragList
        items={selectedItems}
        displayMap={$fieldData.displayMap}
        onUpdate={onMultiSelect}
      />
    </div>
  {/if}
</div>
