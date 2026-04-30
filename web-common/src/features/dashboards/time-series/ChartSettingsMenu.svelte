<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import ChartTypeSelector from "@rilldata/web-common/features/dashboards/time-dimension-details/charts/ChartTypeSelector.svelte";
  import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";

  export let connectNulls: boolean;
  export let dynamicYAxisScale: boolean;
  export let showChartTypeSelector = true;
  export let chartType: TDDChart = TDDChart.DEFAULT;
  export let hasComparison = false;
  export let exploreName = "";
  export let onChartTypeChange: ((type: TDDChart) => void) | undefined =
    undefined;
  export let onDynamicYAxisScaleChange: ((value: boolean) => void) | undefined =
    undefined;

  let open = false;
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
      <div class="flex flex-col gap-y-2">
        <span>Always show as</span>
        <ChartTypeSelector
          {exploreName}
          {chartType}
          {hasComparison}
          {onChartTypeChange}
        />
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
  .popover-item {
    @apply flex flex-row items-center justify-between;
    @apply gap-x-2 h-6;
  }
</style>
