<script lang="ts">
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import TDDHeader from "./TDDHeader.svelte";
  import TddNew from "./TDDNew.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeDimensionDataStore } from "./time-dimension-data-store";

  export let metricViewName;

  const timeDimensionDataStore = useTimeDimensionDataStore(getStateManagers());
  $: console.log($timeDimensionDataStore);

  $: dashboardStore = useDashboardStore(metricViewName);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension;
</script>

<TDDHeader {dimensionName} {metricViewName} />

{#if $timeDimensionDataStore?.data}
  <TddNew data={$timeDimensionDataStore.data} />
{/if}
