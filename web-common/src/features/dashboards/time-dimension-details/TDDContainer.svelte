<script lang="ts">
  import { useDashboardStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import TDDHeader from "./TDDHeader.svelte";
  import TddNew from "./TDDNew.svelte";

  export let metricViewName;

  $: dashboardStore = useDashboardStore(metricViewName);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension;
</script>

<TDDHeader {dimensionName} {metricViewName} />
<TddNew />
