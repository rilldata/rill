<script lang="ts">
  import { fly } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithTween } from "$lib/components/data-graphic/functional-components";
  import { Axis, Grid, PointLabel } from "$lib/components/data-graphic/guides";
  import { Area, Line } from "$lib/components/data-graphic/marks";
  import { interpolateArray } from "d3-interpolate";
  import { Body } from "$lib/components/data-graphic/elements";
  import { guidGenerator } from "$lib/util/guid";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
  } from "$lib/util/humanize-numbers";
  export let start;
  export let end;
  export let formatPreset: NicelyFormattedTypes;
  export let data;
  export let accessor: string;
  export let mouseover = undefined;
  export let key: string;

  // bind and send up to parent to create global mouseover
  export let mouseoverValue = undefined;

  // workaround for formatting dates etc.
  //const xFormatter = interval.includes('day') ?

  $: longTimeSeries = data?.length > 1000;
  let longTimeSeriesKey;
  /**
   * Artificially generate a value for the key block.
   * For longer time series (let's say > 1000 pts) we
   * can default to a specialized animation where we mostly
   * just fly out the mark within the Body tag's clip path,
   * making it look like it is sinking into the ocean.
   * It's a nice effect.
   */
  $: if (data?.length > 1000) {
    longTimeSeriesKey = guidGenerator();
  } else {
    longTimeSeriesKey = undefined;
  }

  let hideCurrent = false;
</script>

{#if key && data?.length}
  <div transition:fly|local={{ duration: 500, y: 10 }}>
    <SimpleDataGraphic
      shareYScale={false}
      bind:mouseoverValue
      yMin={0}
      yMaxTweenProps={{ duration: longTimeSeries ? 0 : 600 }}
    >
      <Body>
        {#key key + longTimeSeriesKey}
          <!-- here, we switch hideCurrent before and after the transition, so
            in cases of the key updating, we can gracefully transition all kinds of
            interesting animations.
          -->
          <g
            out:fly|local={{ duration: 500, y: 475 }}
            style:opacity={hideCurrent && !longTimeSeries ? 0.125 : 1}
            style:transition="opacity 250ms"
            on:outrostart={() => {
              hideCurrent = true;
            }}
            on:outroend={() => {
              hideCurrent = false;
            }}
          >
            <WithTween
              value={data}
              let:output={tweenedData}
              tweenProps={{
                duration: longTimeSeries ? 0 : !hideCurrent ? 600 : 0,
                easing: cubicOut,
                interpolate: interpolateArray,
              }}
            >
              <Area data={tweenedData} yAccessor={accessor} xAccessor="ts" />
              <Line data={tweenedData} yAccessor={accessor} xAccessor="ts" />
            </WithTween>
          </g>
        {/key}
      </Body>
      <!--      <Axis-->
      <!--        side="right"-->
      <!--        format={(value) =>-->
      <!--          formatPreset === NicelyFormattedTypes.NONE-->
      <!--            ? `${value}`-->
      <!--            : humanizeDataType(value, formatPreset)}-->
      <!--      />-->
      <Grid />
      {#if mouseover}
        <PointLabel
          tweenProps={{ duration: 50 }}
          x={mouseover.ts}
          y={mouseover[accessor]}
          format={(value) =>
            formatPreset === NicelyFormattedTypes.NONE
              ? value
              : humanizeDataType(value, formatPreset)}
        />
      {/if}
    </SimpleDataGraphic>
  </div>
{/if}
