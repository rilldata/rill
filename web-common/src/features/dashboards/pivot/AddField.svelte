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
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { appendChipsToZone, splitTagItems } from "./pivot-utils";

  export let zone: "rows" | "columns" | null = null;

  // Prefix used to identify tag rows in the dropdown's flat name space.
  // Tag names can collide with dimension/measure ids otherwise.
  const TAG_PREFIX = "__tag__:";

  const {
    selectors: {
      pivot: { dimensions, measures, rows, originalColumns },
      tags: { combinedTagIndex, dimensionTagIndex, measureTagIndex },
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

  // Tag rows count items that are routable for the current zone: rows can
  // only accept dimensions, so a tag of pure measures has nothing to add
  // there and is hidden.
  $: tagGroupItems = $combinedTagIndex.tags
    .map((t) => {
      const dimCount = $dimensionTagIndex.itemsByTag.get(t.name)?.length ?? 0;
      const measCount = $measureTagIndex.itemsByTag.get(t.name)?.length ?? 0;
      const usable = zone === "rows" ? dimCount : dimCount + measCount;
      return { tag: t, dimCount, measCount, usable };
    })
    .filter((t) => t.usable > 0)
    .map(({ tag, dimCount, measCount }) => ({
      name: `${TAG_PREFIX}${tag.name}`,
      label:
        dimCount > 0 && measCount > 0
          ? `${tag.name} (${dimCount} dim · ${measCount} meas)`
          : dimCount > 0
            ? `${tag.name} (${dimCount} dim)`
            : `${tag.name} (${measCount} meas)`,
    }));

  $: selectableGroups = [
    ...(tagGroupItems.length > 0
      ? [
          <SearchableFilterSelectableGroup>{
            name: m.pivot_tags(),
            items: tagGroupItems,
          },
        ]
      : []),
    ...(zone === "columns"
      ? [
          <SearchableFilterSelectableGroup>{
            name: m.dashboard_measures_label(),
            items: $measures?.map((m) => ({
              name: m.id,
              label: m.title,
            })),
          },
        ]
      : []),
    <SearchableFilterSelectableGroup>{
      name: m.dashboard_dimensions_label(),
      items: $dimensions?.map((d) => ({
        name: d.id,
        label: d.title,
      })),
    },
    <SearchableFilterSelectableGroup>{
      name: m.dashboard_time_label(),
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

  function handleSelectValue(name: string) {
    let toAdd: PivotChipData[];

    if (name.startsWith(TAG_PREFIX)) {
      const tagName = name.slice(TAG_PREFIX.length);
      const { dimensions: dims, measures: meas } = splitTagItems(
        tagName,
        $dimensionTagIndex,
        $measureTagIndex,
      );
      toAdd = zone === "rows" ? dims : [...dims, ...meas];
    } else {
      const selectedItem = allDimensionsMeasures.find(
        (item) => item.id === name,
      ) as PivotChipData | undefined;
      if (!selectedItem) return;
      toAdd = [selectedItem];
    }

    if (toAdd.length === 0) return;

    // appendChipsToZone dedups against both zones so a dimension never lands
    // in rows and columns at once. The dropdown sources already exclude
    // placed dimensions/measures, but time grains and tag bulk-adds can
    // include items already present elsewhere — this is the catch.
    if (zone === "rows") {
      const next = appendChipsToZone($rows, $originalColumns, toAdd);
      if (next.length !== $rows.length) {
        metricsExplorerStore.setPivotRows($exploreName, next);
      }
    } else {
      const next = appendChipsToZone($originalColumns, $rows, toAdd);
      if (next.length !== $originalColumns.length) {
        metricsExplorerStore.setPivotColumns($exploreName, next);
      }
    }
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <button
        {...props}
        class:active={open}
        aria-label={m.dashboard_add_filter_button_aria()}
      >
        <Add size="17px" />
      </button>
    {/snippet}
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
    @apply bg-input border;
  }

  button:hover {
    @apply bg-surface-hover;
  }

  button:active,
  .active {
    @apply bg-surface-active;
  }
</style>
