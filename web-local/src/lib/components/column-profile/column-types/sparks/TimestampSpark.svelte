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

  //export let data;

  export let width = undefined;
  export let height = undefined;
  export let data: unknown[];
  // export let curve = "curveLinear";
  // export let area = false;
  // export let color = "hsl(217, 10%, 50%)";
  // export let tweenIn = false;

  // // the color of the zoom window
  // export let zoomWindowColor = "hsla(217, 90%, 60%, .2)";
  // // the color of the zoom window boundaries
  // export let zoomWindowBoundaryColor = "rgb(100,100,100)";
  // export let zoomWindowXMin: Date = undefined;
  // export let zoomWindowXMax: Date = undefined;

  export let xAccessor: string = undefined;
  export let yAccessor: string = undefined;

  // rowsize for table

  export let left = 0;
  export let right = 0;
  export let top = 12;
  export let bottom = 4;

  // export let buffer = 4;
  // export let leftBuffer = buffer;
  // export let rightBuffer = buffer;
  // export let topBuffer = buffer;
  // export let bottomBuffer = buffer;

  // export let objectName: string;
  // export let columnName: string;

  // $: sparkQuery = useRuntimeServiceGenerateTimeSeries(
  //   $runtimeStore?.instanceId,
  //   // FIXME: convert pixel back to number once the API
  //   {
  //     tableName: objectName,
  //     timestampColumnName: columnName,
  //     pixels: 92,
  //   }
  // );
  // let data = [];
  // $: data = convertTimestampPreview(
  //   $sparkQuery?.data?.rollup?.spark?.map((di) => {
  //     let next = { ...di };
  //     next[yAccessor] = next.records[yAccessor];
  //     return next;
  //   }) || []
  // );
</script>

{#if data?.length}
  <WithParentClientRect let:rect>
    <SimpleDataGraphic
      xType="date"
      yType="number"
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
