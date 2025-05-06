<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";

  export let exploreName: string;
  export let chartType: TDDChart;
  export let hasComparison: boolean;

  const comparisonCharts = [TDDChart.STACKED_AREA, TDDChart.STACKED_BAR];

  const chartTypeTabs = [
    {
      label: "Line",
      id: TDDChart.DEFAULT,
      Icon: LineChart,
    },
    {
      label: "Bar",
      id: TDDChart.GROUPED_BAR,
      Icon: BarChart,
    },
    {
      label: "Stacked area",
      id: TDDChart.STACKED_AREA,
      Icon: StackedArea,
    },
    {
      label: "Stacked bar",
      id: TDDChart.STACKED_BAR,
      Icon: StackedBar,
    },
  ];

  function handleChartTypeChange(type: TDDChart, isDisabled: boolean) {
    if (isDisabled) return;
    metricsExplorerStore.setTDDChartType(exploreName, type);
  }

  // switch to default if current selected chart is not available
  $: if (!hasComparison && comparisonCharts.includes(chartType)) {
    metricsExplorerStore.setTDDChartType(exploreName, TDDChart.DEFAULT);
  }
</script>

<div class="chart-type-selector">
  {#each chartTypeTabs as { label, id, Icon } (label)}
    {@const active = chartType === id}
    {@const disabled = !hasComparison && comparisonCharts.includes(id)}
    <div class:bg-theme-100={active} class="chart-icon-wrapper">
      <IconButton
        {disabled}
        disableHover
        tooltipLocation="top"
        on:click={() => handleChartTypeChange(id, disabled)}
      >
        <Icon
          primaryColor={disabled
            ? "var(--color-gray-400)"
            : "var(--color-theme-700)"}
          secondaryColor={disabled
            ? "var(--color-gray-300)"
            : "var(--color-theme-300)"}
          size="20px"
        />
        <svelte:fragment slot="tooltip-content">
          {disabled ? `Add comparison values to use ${label} chart` : label}
        </svelte:fragment>
      </IconButton>
    </div>
  {/each}
</div>

<style lang="postcss">
  .chart-type-selector {
    @apply flex ml-auto overflow-hidden mr-4;
    @apply border border-theme-300 divide-x divide-theme-300 rounded-sm;
  }
  .chart-icon-wrapper {
    @apply p-1;
  }

  .chart-icon-wrapper:hover {
    @apply bg-theme-100;
  }
</style>
