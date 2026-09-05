<script lang="ts">
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { PlusIcon } from "lucide-svelte";
  import { useMetricFieldData } from "../selectors";
  import type { FieldType } from "../types";

  export let canvasName: string;
  export let metricName: string;
  export let selectedItems: string[] = [];
  export let types: FieldType[];
  export let excludedValues: string[] | undefined = undefined;
  export let ephemeralMeasures: EphemeralMeasureDef[] | undefined = undefined;
  // When set, the menu footer offers a "Create calculated measure" action.
  export let onCreateEphemeral: (() => void) | undefined = undefined;
  // When set, ephemeral items get an edit button invoking this.
  export let onEditEphemeral: ((name: string) => void) | undefined = undefined;
  export let onMultiSelect: (items: string[]) => void = () => {};
  export let open = false;
  export let searchValue = "";
  export let allowMultiSelect = true;
  export let allowSelectAll = false;

  const client = useRuntimeClient();

  // Local state for optimistic updates
  let localSelectedItems: string[] = selectedItems;

  $: ctx = getCanvasStore(canvasName, client.instanceId);
  $: fieldData = useMetricFieldData(
    ctx,
    metricName,
    types,
    undefined,
    searchValue,
    excludedValues,
    ephemeralMeasures,
  );
  $: selectableGroups = [
    ...(types.includes("measure")
      ? [
          <SearchableFilterSelectableGroup>{
            name: "measure",
            label: "MEASURES",
            items: $fieldData.filteredItems
              .filter((item) => $fieldData.displayMap[item]?.type === "measure")
              .map((item) => {
                const def = ephemeralMeasures?.find((d) => d.name === item);
                return {
                  name: item,
                  label: $fieldData.displayMap[item].label,
                  ...(def
                    ? { description: def.expression, ephemeral: true }
                    : {}),
                };
              }),
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

<DropdownMenu.Root bind:open>
  <slot name="trigger" {open} />

  {#if onCreateEphemeral}
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
      onEditItem={onEditEphemeral
        ? (name) => {
            open = false;
            onEditEphemeral?.(name);
          }
        : undefined}
    >
      <button
        slot="action"
        class="create-ephemeral-measure"
        type="button"
        on:click={() => {
          open = false;
          onCreateEphemeral?.();
        }}
      >
        <PlusIcon size="14px" />
        {m.dashboard_pivot_ephemeral_create()}
      </button>
    </SearchableMenuContent>
  {:else}
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
      onEditItem={onEditEphemeral
        ? (name) => {
            open = false;
            onEditEphemeral?.(name);
          }
        : undefined}
    />
  {/if}
</DropdownMenu.Root>

<style lang="postcss">
  .create-ephemeral-measure {
    @apply flex w-full items-center gap-x-1.5 rounded-none border-none bg-transparent;
    @apply h-7 px-2 text-xs text-fg-primary;
  }

  .create-ephemeral-measure:hover {
    @apply bg-surface-hover;
  }
</style>
