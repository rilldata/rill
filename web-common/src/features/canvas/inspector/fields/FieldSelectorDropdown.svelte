<script lang="ts">
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import CreateEphemeralMeasureButton from "@rilldata/web-common/features/dashboards/ephemeral-measures/CreateEphemeralMeasureButton.svelte";
  import { ephemeralMeasuresFirst } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useMetricFieldData } from "../selectors";
  import type { FieldType } from "../types";

  export let canvasName: string;
  export let metricName: string;
  export let selectedItems: string[] = [];
  export let types: FieldType[];
  export let excludedValues: string[] | undefined = undefined;
  export let ephemeralMeasures: EphemeralMeasureDef[] | undefined = undefined;
  // When set, the menu footer offers a "Create adhoc measure" action.
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
  let selectionIntents = new Map<string, boolean>();

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
            items: ephemeralMeasuresFirst(
              $fieldData.filteredItems.filter(
                (item) => $fieldData.displayMap[item]?.type === "measure",
              ),
              (item) => item,
              (ephemeralMeasures ?? []).map((def) => def.name),
            ).map((item) => {
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

  $: if (!open && selectionIntents.size) selectionIntents = new Map();
  $: localSelectedItems = applySelectionIntents(
    selectedItems,
    selectionIntents,
  );

  function applySelectionIntents(
    items: string[],
    intents: Map<string, boolean>,
  ) {
    const selected = new Set(items);
    for (const [name, included] of intents) {
      if (included) selected.add(name);
      else selected.delete(name);
    }
    return Array.from(selected);
  }

  function handleSelect(name: string) {
    if (allowMultiSelect) {
      // Keep this session's toggles when an older reconciled spec is echoed back,
      // while still accepting external changes to other fields.
      selectionIntents = new Map(selectionIntents).set(
        name,
        !localSelectedItems.includes(name),
      );
      localSelectedItems = applySelectionIntents(
        selectedItems,
        selectionIntents,
      );
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
      <CreateEphemeralMeasureButton
        slot="action"
        count={ephemeralMeasures?.length ?? 0}
        onOpen={() => (open = false)}
        onCreate={onCreateEphemeral}
      />
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
