<script lang="ts">
  import TDDHeader from "@rilldata/web-common/features/dashboards/time-dimension-details/TDDHeader.svelte";
  import PivotTable from "./PivotTable.svelte";

  export let metricViewName;
</script>

<TDDHeader {metricViewName} dimensionName="" comparing="time" />

<PivotTable />
