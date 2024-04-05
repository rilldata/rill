<script lang="ts">
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import FormSection from "@rilldata/web-common/components/forms/FormSection.svelte";
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";

  export let metricsViewName: string;

  $: dashboardStore = useDashboardStore(metricsViewName);

  let timeRange: V1TimeRange;
  $: timeRange = {
    isoDuration: $dashboardStore.selectedTimeRange?.name,
    start: $dashboardStore.selectedTimeRange?.start?.toISOString() ?? "",
    end: $dashboardStore.selectedTimeRange?.end?.toISOString() ?? "",
  };
</script>

<FormSection
  description={"Inherited from underlying dashboard view."}
  padding=""
  title="Filters"
>
  <FilterChipsReadOnly
    comparisonTimeRange={undefined}
    dimensionThresholdFilters={$dashboardStore.dimensionThresholdFilters}
    filters={$dashboardStore.whereFilter}
    {metricsViewName}
    {timeRange}
  />
</FormSection>
