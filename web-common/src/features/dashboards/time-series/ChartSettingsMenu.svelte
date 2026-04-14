<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
  import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
  import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";
  import type { ComponentType, SvelteComponent } from "svelte";

  export let connectNulls: boolean;
  export let dynamicYAxisScale: boolean;
  export let showChartTypeSelector = true;
  export let chartType: TDDChart = TDDChart.DEFAULT;
  export let hasComparison = false;
  export let onChartTypeChange: ((type: TDDChart) => void) | undefined =
    undefined;
  export let onDynamicYAxisScaleChange: ((value: boolean) => void) | undefined =
    undefined;

  let open = false;

  const comparisonCharts = [TDDChart.STACKED_AREA, TDDChart.STACKED_BAR];

  const chartTypeTabs: {
    label: string;
    id: TDDChart;
    Icon: ComponentType<SvelteComponent>;
  }[] = [
    { label: "Line", id: TDDChart.DEFAULT, Icon: LineChart },
    { label: "Bar", id: TDDChart.GROUPED_BAR, Icon: BarChart },
    { label: "Stacked area", id: TDDChart.STACKED_AREA, Icon: StackedArea },
    { label: "Stacked bar", id: TDDChart.STACKED_BAR, Icon: StackedBar },
  ];
</script>

<Popover bind:open>
  <PopoverTrigger>
    <IconButton rounded active={open}>
      <MoreHorizontal size="16px" />
    </IconButton>
  </PopoverTrigger>
  <PopoverContent
    align="start"
    side="bottom"
    class="flex flex-col gap-y-2 w-[260px] px-3.5 py-2.5"
  >
    {#if showChartTypeSelector}
      <div class="flex flex-row items-center justify-between gap-x-2">
        <span>Always show as</span>
        <div class="chart-type-selector">
          {#each chartTypeTabs as { label, id, Icon }, i (id)}
            {#if i > 0}
              <div class="chart-type-divider"></div>
            {/if}
            {@const active = chartType === id}
            {@const disabled = !hasComparison && comparisonCharts.includes(id)}
            <IconButton
              {disabled}
              disableHover
              tooltipLocation="top"
              onclick={() => {
                if (!disabled) onChartTypeChange?.(id);
              }}
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
                {disabled
                  ? `Add comparison values to use ${label} chart`
                  : label}
              </svelte:fragment>
            </IconButton>
          {/each}
        </div>
      </div>
    {/if}
    <div class="popover-item">
      <span>Connect sparse data</span>
      <Switch
        small
        checked={connectNulls}
        onCheckedChange={() => (connectNulls = !connectNulls)}
      />
    </div>
    <div class="popover-item">
      <span>Dynamic Y-axis scale</span>
      <Switch
        small
        checked={dynamicYAxisScale}
        onCheckedChange={() => {
          dynamicYAxisScale = !dynamicYAxisScale;
          onDynamicYAxisScaleChange?.(dynamicYAxisScale);
        }}
      />
    </div>
  </PopoverContent>
</Popover>

<style lang="postcss">
  .chart-type-selector {
    @apply flex items-center gap-x-0.5;
    @apply bg-surface-muted rounded p-0.5 h-6;
    @apply border border-slate-200;
  }

  .chart-type-divider {
    @apply w-px h-4 bg-slate-300;
  }

  .popover-item {
    @apply flex flex-row items-center justify-between;
    @apply gap-x-2 h-6;
  }
</style>
