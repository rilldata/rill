<script lang="ts">
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import TDDHeader from "./TDDHeader.svelte";
  import TddNew from "./TDDNew.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeDimensionDataStore } from "./time-dimension-data-store";

  export let metricViewName;

  const timeDimensionDataStore = useTimeDimensionDataStore(getStateManagers());
  $: dashboardStore = useDashboardStore(metricViewName);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension;

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;
</script>

<TDDHeader {dimensionName} {metricViewName} />

{#if $timeDimensionDataStore?.data}
  <TddNew
    {dimensionName}
    {metricViewName}
    {excludeMode}
    comparing={$timeDimensionDataStore.comparing}
    timeFormatter={$timeDimensionDataStore.timeFormatter}
    data={$timeDimensionDataStore.data}
  />
{/if}
