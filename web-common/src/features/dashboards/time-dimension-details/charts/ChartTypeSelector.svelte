<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";

  export let metricViewName: string;
  export let chartType: TDDChart;
  export let isDimensional: boolean;

  const dimensionalCharts = [TDDChart.STACKED_AREA, TDDChart.STACKED_BAR];

  function handleChartTypeChange(type: TDDChart, isDisabled: boolean) {
    if (isDisabled) return;
    metricsExplorerStore.setTDDChartType(metricViewName, type);
  }
  const chartTypeTabs = [
    {
      label: "Line chart",
      id: TDDChart.DEFAULT,
      Icon: LineChart,
    },
    {
      label: "Bar  chart",
      id: TDDChart.GROUPED_BAR,
      Icon: BarChart,
    },
    {
      label: "Stacked Bar chart",
      id: TDDChart.STACKED_BAR,
      Icon: StackedBar,
    },
    {
      label: "Stacked Area chart",
      id: TDDChart.STACKED_AREA,
      Icon: StackedArea,
    },
  ];
</script>

<div class="chart-type-selector">
  {#each chartTypeTabs as { label, id, Icon } (label)}
    {@const active = chartType === id}
    {@const disabled = !isDimensional && dimensionalCharts.includes(id)}
    <div class:bg-primary-100={active} class="chart-icon-wrapper">
      <IconButton
        {disabled}
        disableHover
        tooltipLocation="top"
        on:click={() => handleChartTypeChange(id, disabled)}
      >
        <Icon
          primaryColor={disabled ? "#64748b" : "var(--color-primary-700)"}
          secondaryColor={disabled ? " #cbd5e1 " : "var(--color-primary-300)"}
          size="20px"
        />
        <svelte:fragment slot="tooltip-content">
          {label}
        </svelte:fragment>
      </IconButton>
    </div>
  {/each}
</div>

<style lang="postcss">
  .chart-type-selector {
    @apply flex ml-auto overflow-hidden;
    @apply border border-primary-300 divide-x divide-primary-300 rounded-sm;
  }
  .chart-icon-wrapper {
    @apply p-1;
  }

  .chart-icon-wrapper:hover {
    @apply bg-primary-100;
  }
</style>
