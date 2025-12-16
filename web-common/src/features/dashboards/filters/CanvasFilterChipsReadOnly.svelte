<script lang="ts">
  import TimeRangeReadOnly from "./TimeRangeReadOnly.svelte";
  import DimensionFilterReadOnlyChip from "./dimension-filters/DimensionFilterReadOnlyChip.svelte";
  import MeasureFilterReadOnlyChip from "./measure-filters/MeasureFilterReadOnlyChip.svelte";
  import type { UIFilters } from "../../canvas/stores/filter-manager";

  export let uiFilters: UIFilters;
  export let timeRangeString: string | undefined = undefined;
  export let comparisonRange: string | undefined = undefined;
  export let timeStart: string | undefined = undefined;
  export let timeEnd: string | undefined = undefined;
  export let col = true;

  $: ({ dimensionFilters, measureFilters } = uiFilters);
</script>

<div
  class:flex-col={col}
  class="flex gap-y-2 gap-x-2 w-full flex-none"
  aria-label="Readonly Filter Chips"
>
  <div class="flex gap-x-2">
    {#if timeRangeString}
      <TimeRangeReadOnly
        timeRange={{ expression: timeRangeString }}
        comparisonTimeRange={comparisonRange
          ? { expression: comparisonRange }
          : undefined}
      />
    {/if}
  </div>

  {#each dimensionFilters as [id, filterData] (id)}
    {@const metricsViewNames = Array.from(filterData.dimensions.keys())}
    {@const dimension = filterData.dimensions.get(metricsViewNames[0])}

    {#if dimension && dimension.name}
      <DimensionFilterReadOnlyChip
        pinned={filterData.pinned}
        name={dimension.name}
        {metricsViewNames}
        label={dimension.displayName ||
          dimension.name ||
          dimension.column ||
          "Unnamed Dimension"}
        mode={filterData.mode}
        values={filterData.selectedValues ?? []}
        inputText={filterData.inputText}
        isInclude={filterData.isInclude === true}
        {timeStart}
        {timeEnd}
      />
    {/if}
  {/each}

  {#each measureFilters as [id, filterData] (id)}
    {@const metricsViewNames = Array.from(filterData?.measures?.keys() ?? [])}
    {@const measure = filterData.measures?.get(metricsViewNames[0])}

    {#if measure && measure.name}
      <MeasureFilterReadOnlyChip
        pinned={filterData.pinned}
        dimensionName={filterData.dimensionName}
        label={filterData.label}
        filter={filterData.filter}
      />
    {/if}
  {/each}
</div>
