<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ArrowUp01, List, Table, Text } from "lucide-svelte";
  import ChartOptions from "./chart/ChartOptions.svelte";

  export let selectedComponent;
  let selectedChartType;

  const chartTypes = [
    { id: "bar", title: "Bar", icon: BarChart },
    { id: "stacked-bar", title: "Stacked Bar", icon: StackedBar },
    { id: "line", title: "Line", icon: LineChart },
  ];

  const genericComponents = [
    { id: "kpi", title: "KPI", icon: ArrowUp01 },
    { id: "table", title: "Table", icon: Table },
    { id: "text", title: "Text", icon: Text },
    { id: "leaderboard", title: "Leaderboard", icon: List },
  ];

  function selectChartType(chartType) {
    selectedChartType = chartType.id;
    selectedComponent = chartType;
  }
</script>

<div class="section">
  <h3>Charts</h3>
  <div class="chart-icons">
    {#each chartTypes as chart}
      <Tooltip distance={8} location="right">
        <Button
          square
          type="secondary"
          selected={selectedChartType === chart.id}
          on:click={() => selectChartType(chart)}
        >
          <svelte:component this={chart.icon} size="18px" />
        </Button>
        <TooltipContent slot="tooltip-content">
          {chart.title}
        </TooltipContent>
      </Tooltip>
    {/each}
  </div>
  {#if selectedChartType}
    <ChartOptions />
  {/if}
</div>

<div class="section">
  <h3>Other Components</h3>
  <div class="generic-icons">
    {#each genericComponents as component}
      <Tooltip distance={8} location="right">
        <Button
          square
          type="secondary"
          on:click={() => selectChartType(component)}
        >
          <svelte:component this={component.icon} size="18px" />
        </Button>
        <TooltipContent slot="tooltip-content">
          {component.title}
        </TooltipContent>
      </Tooltip>
    {/each}
  </div>
</div>

<style lang="postcss">
  .section {
    @apply p-4;
    @apply border-b border-slate-200;
  }

  .section h3 {
    @apply pb-2 font-semibold;
  }

  .chart-icons,
  .generic-icons {
    @apply flex gap-x-2;
  }
</style>
