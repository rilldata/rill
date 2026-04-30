<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import AdaptiveChart from "@rilldata/web-common/components/icons/AdaptiveChart.svelte";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import {
    TDDChart,
    isAdaptiveChartType,
  } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import type { ComponentType, SvelteComponent } from "svelte";

  export let exploreName: string;
  export let chartType: TDDChart;
  export let hasComparison: boolean;
  export let onChartTypeChange: ((type: TDDChart) => void) | undefined =
    undefined;

  const comparisonCharts = [TDDChart.STACKED_AREA, TDDChart.STACKED_BAR];

  let tabEls: HTMLElement[] = [];
  let indicatorLeft = 0;
  let indicatorWidth = 0;
  let ready = false;

  $: activeIndex = chartTypeTabs.findIndex(
    (t) =>
      t.id === chartType ||
      (t.id === TDDChart.DEFAULT && isAdaptiveChartType(chartType)),
  );

  $: if (tabEls[activeIndex]) {
    const el = tabEls[activeIndex];
    indicatorLeft = el.offsetLeft;
    indicatorWidth = el.offsetWidth;
    ready = true;
  }

  const comparisonChartFallbacks: Record<TDDChart, TDDChart> = {
    [TDDChart.STACKED_AREA]: TDDChart.DEFAULT,
    [TDDChart.STACKED_BAR]: TDDChart.GROUPED_BAR,
    [TDDChart.DEFAULT]: TDDChart.DEFAULT,
    [TDDChart.LINE]: TDDChart.LINE,
    [TDDChart.GROUPED_BAR]: TDDChart.GROUPED_BAR,
  };

  const chartTypeTabs: {
    label: string;
    tooltip: string;
    id: TDDChart;
    Icon: ComponentType<SvelteComponent>;
  }[] = [
    {
      label: "Line",
      tooltip: "Line",
      id: TDDChart.LINE,
      Icon: LineChart,
    },
    {
      label: "Bar",
      tooltip: "Bar",
      id: TDDChart.GROUPED_BAR,
      Icon: BarChart,
    },
    {
      label: "Stacked area",
      tooltip: "Stacked area",
      id: TDDChart.STACKED_AREA,
      Icon: StackedArea,
    },
    {
      label: "Stacked bar",
      tooltip: "Stacked bar",
      id: TDDChart.STACKED_BAR,
      Icon: StackedBar,
    },
    {
      label: "Adaptive",
      tooltip:
        "Adaptive: Line chart by default. Switches to bar when there are few data points, and stacked bar when comparing dimension",
      id: TDDChart.DEFAULT,
      Icon: AdaptiveChart,
    },
  ];

  function handleChartTypeChange(type: TDDChart, isDisabled: boolean) {
    if (isDisabled) return;
    if (onChartTypeChange) {
      onChartTypeChange(type);
    } else {
      metricsExplorerStore.setTDDChartType(exploreName, type);
    }
  }

  // switch to non-comparison fallback if current selected chart is not available
  $: if (!hasComparison && comparisonCharts.includes(chartType)) {
    const fallback = comparisonChartFallbacks[chartType];
    if (onChartTypeChange) {
      onChartTypeChange(fallback);
    } else {
      metricsExplorerStore.setTDDChartType(exploreName, fallback);
    }
  }
</script>

<div class="chart-type-selector">
  {#if ready}
    <div
      class="indicator"
      style:left="{indicatorLeft}px"
      style:width="{indicatorWidth}px"
    ></div>
  {/if}
  {#each chartTypeTabs as { label, tooltip, id, Icon }, i (id)}
    {@const active =
      chartType === id ||
      (id === TDDChart.DEFAULT && isAdaptiveChartType(chartType))}
    {@const disabled = !hasComparison && comparisonCharts.includes(id)}
    <div bind:this={tabEls[i]} class="chart-type-tab" class:disabled>
      <IconButton
        {disabled}
        disableHover
        tooltipLocation="top"
        onclick={() => handleChartTypeChange(id, disabled)}
        ariaPressed={active}
      >
        <Icon
          primaryColor={disabled
            ? "var(--color-gray-300)"
            : active
              ? "var(--color-theme-600)"
              : "var(--color-gray-500)"}
          secondaryColor={disabled
            ? "var(--color-gray-200)"
            : active
              ? "var(--color-theme-300)"
              : "var(--color-gray-300)"}
          size="16px"
        />
        <svelte:fragment slot="tooltip-content">
          {disabled ? `Add comparison values to use ${label} chart` : tooltip}
        </svelte:fragment>
      </IconButton>
    </div>
  {/each}
</div>

<style lang="postcss">
  .chart-type-selector {
    @apply relative flex items-center gap-x-1;
    @apply bg-surface-muted rounded-lg p-[3px];
    min-width: 200px;
  }

  .indicator {
    @apply absolute bg-surface-overlay border border-border rounded-md;
    top: 3px;
    bottom: 3px;
    transition:
      left 150ms cubic-bezier(0.4, 0, 0.2, 1),
      width 150ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .chart-type-tab {
    @apply relative z-10 flex-1 flex items-center justify-center;
    @apply rounded-md;
  }

  .chart-type-tab.disabled {
    @apply opacity-80;
  }
</style>
