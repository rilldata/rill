<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";

  export let connectNulls: boolean;
  export let forceLineChart: boolean;
  export let dynamicYAxisScale: boolean;
  export let showForceLineChart = true;
  export let onForceLineChartChange: ((value: boolean) => void) | undefined =
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
    class="flex flex-col gap-y-2 w-[220px] px-3.5 py-2.5"
  >
    <div class="flex flex-row items-center justify-between gap-x-2">
      <span>Connect sparse data</span>
      <Switch
        small
        checked={connectNulls}
        onCheckedChange={() => (connectNulls = !connectNulls)}
      />
    </div>
    {#if showForceLineChart}
      <div class="flex flex-row items-center justify-between gap-x-2">
        <span>Always show as line chart</span>
        <Switch
          small
          checked={forceLineChart}
          onCheckedChange={() => {
            forceLineChart = !forceLineChart;
            onForceLineChartChange?.(forceLineChart);
          }}
        />
      </div>
    {/if}
    <div class="flex flex-row items-center justify-between gap-x-2">
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
