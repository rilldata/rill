<script lang="ts">
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import { useExploreStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";

  export let exploreName: string;

  $: exploreStore = useExploreStore(exploreName);

  let timeRange: V1TimeRange;
  $: timeRange = {
    isoDuration: $exploreStore.selectedTimeRange?.name,
    start: $exploreStore.selectedTimeRange?.start?.toISOString() ?? "",
    end: $exploreStore.selectedTimeRange?.end?.toISOString() ?? "",
  };
</script>

<FormSection
  description={"Inherited from underlying dashboard view."}
  padding=""
  title="Filters"
>
  <FilterChipsReadOnly
    dimensionThresholdFilters={$exploreStore.dimensionThresholdFilters}
    filters={$exploreStore.whereFilter}
    {exploreName}
    {timeRange}
  />
</FormSection>
