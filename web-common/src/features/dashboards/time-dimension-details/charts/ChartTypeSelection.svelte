<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";

  export let metricViewName: string;

  function handleChartTypeChange(type: TDDChart) {
    metricsExplorerStore.setTDDChartType(metricViewName, type);
  }
</script>

<div class="flex gap-x-2 ml-auto">
  <Button
    type="text"
    compact
    on:click={() => handleChartTypeChange(TDDChart.DEFAULT)}>Default</Button
  >
  <Button
    type="text"
    compact
    on:click={() => handleChartTypeChange(TDDChart.GROUPED_BAR)}>Bar</Button
  >
  <Button
    type="text"
    compact
    on:click={() => handleChartTypeChange(TDDChart.STACKED_BAR)}
    >Stacked Bar</Button
  >
  <Button
    type="text"
    compact
    on:click={() => handleChartTypeChange(TDDChart.STACKED_AREA)}
    >Stacked Area</Button
  >
</div>
