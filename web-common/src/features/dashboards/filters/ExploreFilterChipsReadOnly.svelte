<script lang="ts">
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { allDimensions } from "@rilldata/web-common/features/dashboards/state-managers/selectors/dimensions";
  import { allMeasures } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measures";
  import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let exploreName: string;
  export let filters: V1Expression | undefined;
  export let dimensionsWithInlistFilter: string[];
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let displayTimeRange: V1TimeRange | undefined;
  export let displayComparisonTimeRange: V1TimeRange | undefined = undefined;
  // `displayTimeRange` passed to this is usually a relative time range used just for display.
  // But we need resolved start and end based on current time in dimension filters to get query for accurate results.
  export let queryTimeStart: string | undefined = undefined;
  export let queryTimeEnd: string | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: validExploreSpecs = useExploreValidSpec(instanceId, exploreName);
  $: metricsViewName = $validExploreSpecs.data?.explore?.metricsView ?? "";

  // Get dimension filters
  $: dimensions = allDimensions({
    validExplore: $validExploreSpecs.data?.explore,
    validMetricsView: $validExploreSpecs.data?.metricsView,
  });

  // Get measure filters
  $: measures = allMeasures({
    validExplore: $validExploreSpecs.data?.explore,
    validMetricsView: $validExploreSpecs.data?.metricsView,
  });
</script>

<FilterChipsReadOnly
  {metricsViewName}
  {dimensions}
  {measures}
  {dimensionThresholdFilters}
  {dimensionsWithInlistFilter}
  {filters}
  {displayComparisonTimeRange}
  {displayTimeRange}
  {queryTimeStart}
  {queryTimeEnd}
/>
