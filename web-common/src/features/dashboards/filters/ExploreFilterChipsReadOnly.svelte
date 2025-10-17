<script lang="ts">
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { useMetricsViewValidSpec } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let metricsViewName: string;
  export let filters: V1Expression | undefined;
  export let dimensionsWithInlistFilter: string[];
  export let dimensionThresholdFilters: DimensionThresholdFilter[];
  export let displayTimeRange: V1TimeRange | undefined = undefined;
  export let displayComparisonTimeRange: V1TimeRange | undefined = undefined;
  // `displayTimeRange` passed to this is usually a relative time range used just for display.
  // But we need resolved start and end based on current time in dimension filters to get query for accurate results.
  export let queryTimeStart: string | undefined = undefined;
  export let queryTimeEnd: string | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: metricsViewSpec = useMetricsViewValidSpec(instanceId, metricsViewName);

  $: dimensions = $metricsViewSpec.data?.dimensions ?? [];
  $: measures = $metricsViewSpec.data?.measures ?? [];
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
