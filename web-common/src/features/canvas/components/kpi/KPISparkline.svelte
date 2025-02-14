<script lang="ts">
  import { SimpleDataGraphic } from "@rilldata/web-common/components/data-graphic/elements";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import WithRoundToTimegrain from "@rilldata/web-common/components/data-graphic/functional-components/WithRoundToTimegrain.svelte";
  import { ChunkedLine } from "@rilldata/web-common/components/data-graphic/marks";
  import {
    AreaMutedColorGradientLight,
    MainAreaColorGradientDark,
    MainLineColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
  import MeasureValueMouseover from "@rilldata/web-common/features/dashboards/time-series/MeasureValueMouseover.svelte";
  import { niceMeasureExtents } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { TimeRoundingStrategy } from "@rilldata/web-common/lib/time/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { extent } from "d3-array";

  export let sparkData: any[] = [];
  export let measureName: string;
  export let sparklineHeight: number;
  export let sparklineWidth: number;
  export let isSparkRight: boolean;
  export let timeGrain: V1TimeGrain;
  export let measureValueFormatter: (
    value: string | number | null,
  ) => string | null;
  export let numberKind: NumberKind;

  const focusedAreaGradient: [string, string] = [
    MainAreaColorGradientDark,
    AreaMutedColorGradientLight,
  ];

  let mouseoverValue: { x: Date; y: number } | undefined = undefined;
  let hovered = false;

  $: [yExtentMin, yExtentMax] = extent(sparkData, (d) => d[measureName]);
  $: [yMin, yMax] = niceMeasureExtents(
    [yExtentMin, yExtentMax],
    (isSparkRight ? 6 : 5) / 5,
  );

  $: [xMin, xMax] = extent(sparkData, (d) => d["ts_position"]);
  $: hoveredTime = mouseoverValue?.x instanceof Date && mouseoverValue?.x;
</script>

<div class={isSparkRight ? "h-full" : "w-full"}>
  <SimpleDataGraphic
    bind:hovered
    bind:mouseoverValue
    height={sparklineHeight}
    width={sparklineWidth}
    overflowHidden={false}
    top={5}
    bottom={0}
    right={0}
    left={16}
    {xMin}
    {xMax}
    {yMin}
    {yMax}
  >
    <ChunkedLine
      lineOpacity={0.75}
      areaEndOffset="95%"
      lineColor={MainLineColor}
      areaGradientColors={focusedAreaGradient}
      data={sparkData}
      xAccessor="ts"
      yAccessor={measureName}
    />
    {#if hoveredTime && mouseoverValue}
      <WithRoundToTimegrain
        strategy={TimeRoundingStrategy.PREVIOUS}
        value={hoveredTime}
        {timeGrain}
        let:roundedValue
      >
        <WithBisector
          data={sparkData}
          callback={(d) => d["ts"]}
          value={roundedValue}
          let:point
        >
          <MeasureValueMouseover
            {point}
            xAccessor="ts"
            yAccessor={measureName}
            showComparison={false}
            mouseoverFormat={measureValueFormatter}
            {numberKind}
          />
        </WithBisector>
      </WithRoundToTimegrain>
    {/if}
  </SimpleDataGraphic>
</div>
