<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";

  export let exploreName: string;

  $: exploreState = useExploreState(exploreName);

  let timeRange: V1TimeRange;
  $: timeRange = {
    isoDuration: $exploreState.selectedTimeRange?.name,
    start: $exploreState.selectedTimeRange?.start?.toISOString() ?? "",
    end: $exploreState.selectedTimeRange?.end?.toISOString() ?? "",
  };
</script>

<FormSection
  description={"Inherited from underlying dashboard view."}
  padding=""
  title="Filters"
>
  <FilterChipsReadOnly
    dimensionThresholdFilters={$exploreState.dimensionThresholdFilters}
    filters={$exploreState.whereFilter}
    {exploreName}
    {timeRange}
  />
</FormSection>
