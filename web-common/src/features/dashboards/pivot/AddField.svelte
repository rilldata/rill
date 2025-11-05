<script lang="ts" context="module">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
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
  import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";

  export let measures: PivotChipData[];
  export let dimensions: PivotChipData[];
  export let timeControlsForPillActions: Pick<
    TimeControlState,
    "timeStart" | "timeEnd" | "minTimeGrain"
  >;
  export let zone: "rows" | "columns" | null = null;
  export let addField: (value: PivotChipData, rows: boolean) => void;

  let open = false;

  $: allTimeGrains = getAllowedTimeGrains(
    new Date(timeControlsForPillActions.timeStart!),
    new Date(timeControlsForPillActions.timeEnd!),
  ).map((tgo) => {
    return {
      id: tgo.grain,
      title: tgo.label,
      type: PivotChipType.Time,
    };
  });

  $: timeGrainOptions = allTimeGrains.filter(
    (tgo) =>
      timeControlsForPillActions.minTimeGrain === undefined ||
      timeControlsForPillActions.minTimeGrain ===
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED ||
      !isGrainBigger(timeControlsForPillActions.minTimeGrain, tgo.id),
  );

  $: selectableGroups = [
    ...(zone === "columns"
      ? [
          <SearchableFilterSelectableGroup>{
            name: "MEASURES",
            items: measures?.map((m) => ({
              name: m.id,
              label: m.title,
            })),
          },
        ]
      : []),
    <SearchableFilterSelectableGroup>{
      name: "DIMENSIONS",
      items: dimensions?.map((d) => ({
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

  $: allDimensionsMeasures = [...dimensions, ...measures, ...timeGrainOptions];

  function handleSelectValue(name) {
    const selectedItem = allDimensionsMeasures.find(
      (item) => item.id === name,
    ) as PivotChipData;

    addField(selectedItem, zone === "rows");
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
    @apply border border-dashed border-slate-300;
  }

  button:hover {
    @apply bg-slate-100;
  }

  button:active,
  .active {
    @apply bg-slate-200;
  }
</style>
