<script lang="ts">
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import TDDHeader from "./TDDHeader.svelte";
  import TddNew from "./TDDNew.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";

  export let metricViewName;

  const timeSeriesDataStore = useTimeSeriesDataStore(getStateManagers());
  $: console.log($timeSeriesDataStore);

  $: dashboardStore = useDashboardStore(metricViewName);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension;
</script>

<TDDHeader {dimensionName} {metricViewName} />

<TddNew />
