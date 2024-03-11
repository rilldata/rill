<script lang="ts">
  /**
   * TimestampSpark.svelte
   * ---------------------
   * This simple component is a basic sparkline, meant to be used
   * in a table / model profile preview.
   * It optionally enables the user to determine a "window", which
   * is just a box emcompassing the zoomWindowXMin and zoomWindowXMax values.
   */
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithParentClientRect } from "@rilldata/web-common/components/data-graphic/functional-components";
  import {
    Area,
    Line,
  } from "@rilldata/web-common/components/data-graphic/marks";
  import { ScaleType } from "@rilldata/web-common/components/data-graphic/state";
  import type { TimeSeriesDatum } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";

  export let width = undefined;
  export let height = undefined;
  export let data: TimeSeriesDatum[];

  export let xAccessor: string | undefined = undefined;
  export let yAccessor: string | undefined = undefined;

  export let left = 0;
  export let right = 0;
  export let top = 12;
  export let bottom = 4;
</script>

{#if data?.length}
  <WithParentClientRect let:rect>
    <SimpleDataGraphic
      xType={ScaleType.DATE}
      yType={ScaleType.NUMBER}
      width={width || rect?.width || 400}
      height={height || rect?.height}
      {bottom}
      {top}
      {left}
      {right}
      bodyBuffer={0}
      marginBuffer={0}
      let:config
    >
      <Line {data} {xAccessor} {yAccessor} />
      <Area {data} {xAccessor} {yAccessor} />
      <line
        x1={config.plotLeft}
        x2={config.plotRight}
        y1={config.plotBottom}
        y2={config.plotBottom}
        stroke="black"
      />
    </SimpleDataGraphic>
  </WithParentClientRect>
{/if}
