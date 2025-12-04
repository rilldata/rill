<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useMetricFieldData } from "../selectors";
  import type { FieldType } from "../types";

  export let canvasName: string;
  export let metricName: string;
  export let selectedItems: string[] = [];
  export let types: FieldType[];
  export let excludedValues: string[] | undefined = undefined;
  export let onMultiSelect: (items: string[]) => void = () => {};
  export let open = false;
  export let searchValue = "";
  export let allowMultiSelect = true;
  export let allowSelectAll = false;

  // Local state for optimistic updates
  let localSelectedItems: string[] = selectedItems;

  $: ({ instanceId } = $runtime);

  $: ctx = getCanvasStore(canvasName, instanceId);
  $: fieldData = useMetricFieldData(
    ctx,
    metricName,
    types,
    undefined,
    searchValue,
    excludedValues,
  );
  $: selectableGroups = [
    ...(types.includes("measure")
      ? [
          <SearchableFilterSelectableGroup>{
            name: "measure",
            label: "MEASURES",
            items: $fieldData.filteredItems
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
            items: $fieldData.filteredItems
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
            items: $fieldData.filteredItems
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
    if (allowMultiSelect) {
      const selectedProxy = new Set(localSelectedItems);
      if (selectedProxy.has(name)) {
        selectedProxy.delete(name);
      } else {
        selectedProxy.add(name);
      }
      localSelectedItems = Array.from(selectedProxy);
      onMultiSelect(localSelectedItems);
    } else {
      localSelectedItems = [name];
      onMultiSelect(localSelectedItems);
      open = false;
    }
  }
</script>

<DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={false}>
  <slot name="trigger" {open} />

  <SearchableMenuContent
    {selectableGroups}
    selectedItems={selectableGroups.map((group) =>
      localSelectedItems?.filter(
        (item) => $fieldData.displayMap[item]?.type === group.name,
      ),
    )}
    {allowMultiSelect}
    searchText={searchValue}
    {allowSelectAll}
    onSelect={handleSelect}
  />
</DropdownMenu.Root>
