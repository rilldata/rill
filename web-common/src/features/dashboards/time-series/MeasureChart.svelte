<script lang="ts">
  import { outline } from "@rilldata/web-common/components/data-graphic/actions/outline";
  import Body from "@rilldata/web-common/components/data-graphic/elements/Body.svelte";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import WithBisector from "@rilldata/web-common/components/data-graphic/functional-components/WithBisector.svelte";
  import {
    Axis,
    Grid,
  } from "@rilldata/web-common/components/data-graphic/guides";
  import { ChunkedLine } from "@rilldata/web-common/components/data-graphic/marks";
  import { NumberKind } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import { previousValueStore } from "@rilldata/web-local/lib/store-utils";
  import { extent } from "d3-array";
  import { cubicOut } from "svelte/easing";
  import { writable } from "svelte/store";
  import { fly } from "svelte/transition";
  import MeasureValueMouseover from "./MeasureValueMouseover.svelte";
  import { niceMeasureExtents } from "./utils";
  export let width: number = undefined;
  export let height: number = undefined;
  export let xMin: Date = undefined;
  export let xMax: Date = undefined;
  export let yMin: number = undefined;
  export let yMax: number = undefined;

  export let showComparison = false;
  export let data;
  export let xAccessor = "ts";
  export let yAccessor = "value";
  export let mouseoverValue;
  export let hovered = false;
  export let mouseoverFormat: (d: number | Date | string) => string = (v) =>
    v.toString();
  export let mouseoverTimeFormat: (d: number | Date | string) => string = (v) =>
    v.toString();
  export let numberKind: NumberKind = NumberKind.ANY;

  export let tweenProps = { duration: 400, easing: cubicOut };

  // control point for scrub functionality.
  export let scrubbing = false;
  export let scrubEnd = undefined;

  $: [xExtentMin, xExtentMax] = extent(data, (d) => d[xAccessor]);
  $: [yExtentMin, yExtentMax] = extent(data, (d) => d[yAccessor]);
  let comparisonExtents;

  /** if we are making a comparison, factor this into the extents calculation.*/
  $: if (showComparison) {
    comparisonExtents = extent(data, (d) => d[`comparison.${yAccessor}`]);
    yExtentMin = Math.min(yExtentMin, comparisonExtents[0]);
    yExtentMax = Math.max(yExtentMax, comparisonExtents[1]);
  }

  $: [internalYMin, internalYMax] = niceMeasureExtents(
    [
      yMin !== undefined ? yMin : yExtentMin,
      yMax !== undefined ? yMax : yExtentMax,
    ],
    6 / 5
  );

  $: internalXMin = xMin || xExtentMin;
  $: internalXMax = xMax || xExtentMax;
  // we delay the tween if previousYMax < yMax
  let yMaxStore = writable(yExtentMax);
  let previousYMax = previousValueStore(yMaxStore);

  $: yMaxStore.set(yExtentMax);
  const timeRangeKey = writable(`${xMin}-${xMax}`);

  const previousTimeRangeKey = previousValueStore(timeRangeKey);

  // FIXME: move this function to utils.ts
  /** reset the keys to trigger animations on time range changes */
  let syncTimeRangeKey;
  $: {
    timeRangeKey.set(`${xMin}-${xMax}`);
    if ($previousTimeRangeKey !== $timeRangeKey) {
      if (syncTimeRangeKey) clearTimeout(syncTimeRangeKey);
      syncTimeRangeKey = setTimeout(() => {
        previousTimeRangeKey.set($timeRangeKey);
      }, 400);
    }
  }

  $: delay =
    $previousTimeRangeKey === $timeRangeKey && $previousYMax < yExtentMax
      ? 100
      : 0;

  function alwaysBetween(min, max, value) {
    // note: must work with dates
    if (value < min) return min;
    if (value > max) return max;
    return value;
  }

  $: if (scrubbing) {
    scrubEnd = alwaysBetween(internalXMin, internalXMax, mouseoverValue);
  }

  let something = true;
  $: setTimeout(() => {
    something = !something;
  }, 1000);
</script>

<SimpleDataGraphic
  overflowHidden={false}
  yMin={internalYMin}
  yMax={internalYMax}
  shareYScale={false}
  yType="number"
  xType="date"
  {width}
  {height}
  top={4}
  left={12}
  right={50}
  bind:mouseoverValue
  bind:hovered
  let:config
  let:yScale
  yMinTweenProps={tweenProps}
  yMaxTweenProps={tweenProps}
  xMaxTweenProps={tweenProps}
  xMinTweenProps={tweenProps}
>
  <Axis side="right" format={mouseoverFormat} {numberKind} />
  <Grid />
  <Body>
    <!-- key on the time range itself to prevent weird tweening animations.
    We'll need to migrate this to a more robust solution once we've figured out
    the right way to "tile" together a time series with multiple pages of data.
    -->
    {#key $timeRangeKey}
      {#if showComparison}
        <g
          class="transition-opacity"
          class:opacity-80={mouseoverValue?.x}
          class:opacity-40={!mouseoverValue?.x}
        >
          <ChunkedLine
            area={false}
            lineColor={`hsl(217, 10%, 60%)`}
            delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
            duration={$timeRangeKey !== $previousTimeRangeKey ? 0 : 200}
            {data}
            {xAccessor}
            yAccessor="comparison.{yAccessor}"
            key={$timeRangeKey}
          />
        </g>
      {/if}
      <ChunkedLine
        delay={$timeRangeKey !== $previousTimeRangeKey ? 0 : delay}
        duration={$timeRangeKey !== $previousTimeRangeKey ? 0 : 200}
        {data}
        {xAccessor}
        {yAccessor}
        key={$timeRangeKey}
      />
    {/key}
    <line
      x1={config.plotLeft}
      x2={config.plotLeft + config.plotRight}
      y1={yScale(0)}
      y2={yScale(0)}
      class="stroke-blue-200"
    />
  </Body>
  {#if !scrubbing && mouseoverValue?.x}
    <WithBisector
      {data}
      callback={(d) => d[xAccessor]}
      value={mouseoverValue.x}
      let:point
    >
      <g transition:fly|local={{ duration: 100, x: -4 }}>
        <MeasureValueMouseover
          {point}
          {xAccessor}
          {yAccessor}
          {showComparison}
          {mouseoverFormat}
          {numberKind}
        />
      </g>

      <g transition:fly|local={{ duration: 100, x: -4 }}>
        <text
          use:outline
          class="fill-gray-600"
          x={config.plotLeft + config.bodyBuffer + 6}
          y={config.plotTop + 10 + config.bodyBuffer}
        >
          {mouseoverTimeFormat(point[xAccessor])}
        </text>
        {#if showComparison}
          <text
            use:outline
            class="fill-gray-400"
            x={config.plotLeft + config.bodyBuffer + 6}
            y={config.plotTop + 24 + config.bodyBuffer}
          >
            {mouseoverTimeFormat(point[`comparison.${xAccessor}`])} prev.
          </text>
        {/if}
      </g></WithBisector
    >
  {/if}
</SimpleDataGraphic>
