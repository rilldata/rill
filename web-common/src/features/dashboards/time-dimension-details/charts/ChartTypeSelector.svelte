<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import type { ComponentType, SvelteComponent } from "svelte";

  export let exploreName: string;
  export let chartType: TDDChart;
  export let hasComparison: boolean;
  export let onChartTypeChange: ((type: TDDChart) => void) | undefined =
    undefined;

  const comparisonCharts = [TDDChart.STACKED_AREA, TDDChart.STACKED_BAR];

  let buttonEls: HTMLElement[] = [];
  let indicatorLeft = 0;
  let indicatorWidth = 0;
  let ready = false;

  $: activeIndex = chartTypeTabs.findIndex((t) => t.id === chartType);

  $: if (buttonEls[activeIndex]) {
    const el = buttonEls[activeIndex];
    indicatorLeft = el.offsetLeft;
    indicatorWidth = el.offsetWidth;
    ready = true;
  }

  const comparisonChartFallbacks: Record<TDDChart, TDDChart> = {
    [TDDChart.STACKED_AREA]: TDDChart.DEFAULT,
    [TDDChart.STACKED_BAR]: TDDChart.GROUPED_BAR,
    [TDDChart.DEFAULT]: TDDChart.DEFAULT,
    [TDDChart.GROUPED_BAR]: TDDChart.GROUPED_BAR,
  };

  const chartTypeTabs: {
    label: string;
    id: TDDChart;
    Icon: ComponentType<SvelteComponent>;
  }[] = [
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
  {#each chartTypeTabs as { label, id, Icon }, i (id)}
    {#if i > 0}
      <div class="chart-type-divider"></div>
    {/if}
    {@const active = chartType === id}
    {@const disabled = !hasComparison && comparisonCharts.includes(id)}
    <div bind:this={buttonEls[i]} class="chart-icon-wrapper" class:disabled>
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
          size="18px"
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
    @apply relative flex items-center gap-x-1;
    @apply bg-surface-muted rounded p-0.5;
    @apply border border-slate-200;
  }

  .indicator {
    @apply absolute rounded bg-white;
    @apply top-0.5 bottom-0.5;
    box-shadow:
      0 1px 2px rgb(0 0 0 / 0.08),
      0 0 0 1px rgb(0 0 0 / 0.04);
    transition:
      left 150ms cubic-bezier(0.4, 0, 0.2, 1),
      width 150ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .chart-icon-wrapper {
    @apply relative z-10;
  }

  .chart-icon-wrapper.disabled {
    @apply opacity-80;
  }

  .chart-type-divider {
    @apply w-px h-4 bg-slate-300 mx-0.5;
  }
</style>
