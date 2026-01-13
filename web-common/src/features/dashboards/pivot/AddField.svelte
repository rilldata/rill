<script lang="ts" context="module">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  import { metricsExplorerStore } from "../stores/dashboard-stores";

  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    getAllowedTimeGrains,
    isGrainBigger,
  } from "@rilldata/web-common/lib/time/grains";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";

  import { PivotChipType } from "./types";
  import type { PivotChipData } from "./types";
</script>

<script lang="ts">
  export let zone: "rows" | "columns" | null = null;

  const {
    selectors: {
      pivot: { dimensions, measures },
    },
    exploreName,
  } = getStateManagers();
  const timeControlsStore = useTimeControlStore(getStateManagers());

  let open = false;

  $: allTimeGrains = getAllowedTimeGrains(
    new Date($timeControlsStore.timeStart!),
    new Date($timeControlsStore.timeEnd!),
  ).map((tgo) => {
    return {
      id: tgo.grain,
      title: tgo.label,
      type: PivotChipType.Time,
    };
  });

  $: timeGrainOptions = allTimeGrains.filter(
    (tgo) =>
      $timeControlsStore.minTimeGrain === undefined ||
      $timeControlsStore.minTimeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED ||
      !isGrainBigger($timeControlsStore.minTimeGrain, tgo.id),
  );

  $: selectableGroups = [
    ...(zone === "columns"
      ? [
          <SearchableFilterSelectableGroup>{
            name: "MEASURES",
            items: $measures?.map((m) => ({
              name: m.id,
              label: m.title,
            })),
          },
        ]
      : []),
    <SearchableFilterSelectableGroup>{
      name: "DIMENSIONS",
      items: $dimensions?.map((d) => ({
        name: d.id,
        label: d.title,
      })),
    },
    <SearchableFilterSelectableGroup>{
      name: "TIME",
      items: timeGrainOptions.map((tgo) => ({
        name: tgo.id,
        label: tgo.title,
        type: PivotChipType.Time,
      })),
    },
  ];

  $: allDimensionsMeasures = [
    ...$dimensions,
    ...$measures,
    ...timeGrainOptions,
  ];

  function handleSelectValue(name) {
    const selectedItem = allDimensionsMeasures.find(
      (item) => item.id === name,
    ) as PivotChipData;

    metricsExplorerStore.addPivotField(
      $exploreName,
      selectedItem,
      zone === "rows",
    );
  }
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      class:active={open}
      use:builder.action
      {...builder}
      aria-label="Add filter button"
    >
      <Add size="17px" />
    </button>
  </DropdownMenu.Trigger>

  <SearchableMenuContent
    allowMultiSelect={false}
    onSelect={(name) => {
      handleSelectValue(name);
    }}
    {selectableGroups}
    selectedItems={[]}
  />
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;
    @apply bg-surface;
  }

  button.addBorder {
    @apply border border-dashed border-gray-300;
  }

  button:hover {
    @apply bg-gray-100;
  }

  button:active,
  .active {
    @apply bg-gray-200;
  }
</style>
