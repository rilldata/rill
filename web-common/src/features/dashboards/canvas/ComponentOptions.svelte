<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ArrowUp01, List, Table, Text } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import ChartOptions from "./chart/ChartOptions.svelte";

  const dispatch = createEventDispatcher();

  const chartTypes = [
    { id: "bar", title: "Bar", icon: BarChart },
    { id: "stacked-bar", title: "Stacked Bar", icon: StackedBar },
    { id: "line", title: "Line", icon: LineChart },
  ];

  const coreComponents = [
    { id: "kpi", title: "KPI", icon: ArrowUp01 },
    { id: "table", title: "Table", icon: Table },
    { id: "text", title: "Text", icon: Text },
    { id: "leaderboard", title: "Leaderboard", icon: List },
  ];
  let selectedChartType;

  function selectChartType(chartType) {
    selectedChartType = chartType.id;
    dispatch("select", chartType);
  }
</script>

<div class="section">
  <InputLabel
    label="Charts"
    id="chart-components"
    hint="Chose a chart component to add to your canvas"
  />
  <div class="chart-icons">
    {#each chartTypes as chart}
      <Tooltip distance={8} location="right">
        <Button
          square
          type="secondary"
          selected={selectedChartType === chart.id}
          on:click={() => selectChartType(chart)}
        >
          <svelte:component this={chart.icon} size="24px" />
        </Button>
        <TooltipContent slot="tooltip-content">
          {chart.title}
        </TooltipContent>
      </Tooltip>
    {/each}
  </div>
  {#if selectedChartType}
    <ChartOptions chartType={selectedChartType} />
  {/if}
</div>

<div class="section">
  <InputLabel
    label="Core components"
    id="core-components"
    hint="Chose a core component to add to your canvas"
  />
  <div class="core-icons">
    {#each coreComponents as component}
      <Tooltip distance={8} location="right">
        <Button
          square
          type="secondary"
          on:click={() => selectChartType(component)}
        >
          <svelte:component this={component.icon} size="24px" />
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
    @apply flex flex-col gap-y-2 p-4;
    @apply border-b border-slate-200;
  }

  .chart-icons,
  .core-icons {
    @apply flex gap-x-2;
  }
</style>
