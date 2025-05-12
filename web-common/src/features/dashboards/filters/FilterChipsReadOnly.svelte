<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import TimeRangeReadOnly from "@rilldata/web-common/features/dashboards/filters/TimeRangeReadOnly.svelte";
  import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { getDimensionFilters } from "../state-managers/selectors/dimension-filters";
  import { getMeasureFilters } from "../state-managers/selectors/measure-filters";
  import DimensionFilterReadOnlyChip from "./dimension-filters/DimensionFilterReadOnlyChip.svelte";
  import MeasureFilterReadOnlyChip from "./measure-filters/MeasureFilterReadOnlyChip.svelte";

  export let metricsViewName: string;
  export let dimensions: MetricsViewSpecDimension[];
  export let measures: MetricsViewSpecMeasure[];
  export let filters: V1Expression | undefined;
  export let dimensionsWithInlistFilter: string[];
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let displayTimeRange: V1TimeRange | undefined;
  export let displayComparisonTimeRange: V1TimeRange | undefined = undefined;
  // `displayTimeRange` passed to this is usually a relative time range used just for display.
  // But we need resolved start and end based on current time in dimension filters to get query for accurate results.
  export let queryTimeStart: string | undefined = undefined;
  export let queryTimeEnd: string | undefined = undefined;

  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => dimension.name as string,
  );
  $: dimensionFilters = getDimensionFilters(
    dimensionIdMap,
    filters,
    dimensionsWithInlistFilter,
  );

  $: measureIdMap = getMapFromArray(
    measures,
    (measure) => measure.name as string,
  );
  $: measureFilters = getMeasureFilters(
    measureIdMap,
    dimensionThresholdFilters,
  );
</script>

<div
  class="relative flex flex-row flex-wrap gap-x-2 gap-y-2 items-center"
  aria-label="Readonly Filter Chips"
>
  {#if displayTimeRange}
    <TimeRangeReadOnly
      timeRange={displayTimeRange}
      comparisonTimeRange={displayComparisonTimeRange}
    />
  {/if}
  {#if dimensionFilters.length > 0}
    {#each dimensionFilters as { name, label, mode, selectedValues, inputText, isInclude } (name)}
      {@const dimension = dimensions.find((d) => d.name === name)}
      <div animate:flip={{ duration: 200 }}>
        {#if dimension?.column || dimension?.expression}
          <DimensionFilterReadOnlyChip
            {name}
            metricsViewNames={[metricsViewName]}
            {mode}
            label={label || name}
            values={selectedValues}
            {inputText}
            {isInclude}
            timeStart={queryTimeStart}
            timeEnd={queryTimeEnd}
          />
        {/if}
      </div>
    {/each}
  {/if}
  {#if measureFilters.length > 0}
    {#each measureFilters as { name, label, dimensionName, filter } (name)}
      <div animate:flip={{ duration: 200 }}>
        <MeasureFilterReadOnlyChip
          label={label || name}
          {dimensionName}
          {filter}
        />
      </div>
    {/each}
  {/if}
</div>
