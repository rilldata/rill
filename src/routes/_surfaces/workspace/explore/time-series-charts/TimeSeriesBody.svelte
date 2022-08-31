<script lang="ts">
  import { Body } from "$lib/components/data-graphic/elements";
  import SimpleDataGraphic from "$lib/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithTween } from "$lib/components/data-graphic/functional-components";
  import { Axis, Grid, PointLabel } from "$lib/components/data-graphic/guides";
  import { Area, Line } from "$lib/components/data-graphic/marks";
  import { guidGenerator } from "$lib/util/guid";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
  } from "$lib/util/humanize-numbers";
  import { interpolateArray } from "d3-interpolate";
  import { cubicOut } from "svelte/easing";
  import { fly } from "svelte/transition";

  export let start;
  export let end;
  export let formatPreset: NicelyFormattedTypes;
  export let data;
  export let accessor: string;
  export let yMin = 0;

  // the recycled mouseover event, in case anyone else has one set
  export let mouseover = undefined;
  export let key: string;
  // bind and send up to parent to create global mouseover
  export let mouseoverValue = undefined;

  // workaround for formatting dates etc.
  //const xFormatter = interval.includes('day') ?

  // bug: currently `data` continuously refreshes for no apparent reason
  // hack: we use `dataCopy` so that continuous `data` updates don't lead to unneccessary rerenders
  let dataCopy;
  $: if (data !== dataCopy) dataCopy = data;

  $: longTimeSeries = dataCopy?.length > 1000;
  let longTimeSeriesKey;
  /**
   * Artificially generate a value for the key block.
   * For longer time series (let's say > 1000 pts) we
   * can default to a specialized animation where we mostly
   * just fly out the mark within the Body tag's clip path,
   * making it look like it is sinking into the ocean.
   * It's a nice effect.
   */
  $: if (dataCopy?.length > 1000) {
    longTimeSeriesKey = guidGenerator();
  } else {
    longTimeSeriesKey = undefined;
  }

  let hideCurrent = false;

  $: allZeros = dataCopy.every((di) => di[accessor] === 0);
  $: dataInDomain = dataCopy.some((di) => di.ts >= start && di.ts <= end);
</script>

{#if key && dataCopy?.length}
  <div transition:fly|local={{ duration: 500, y: 10 }}>
    <SimpleDataGraphic
      shareYScale={false}
      bind:mouseoverValue
      {yMin}
      yMinTweenProps={{ duration: longTimeSeries ? 0 : allZeros ? 100 : 300 }}
      yMaxTweenProps={{ duration: longTimeSeries ? 0 : allZeros ? 100 : 300 }}
      let:xScale
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
              value={dataCopy}
              let:output={tweenedData}
              tweenProps={{
                duration: longTimeSeries
                  ? 0
                  : !hideCurrent
                  ? allZeros
                    ? 0
                    : 300
                  : 0,
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
      <Axis
        side="right"
        format={(value) =>
          formatPreset === NicelyFormattedTypes.NONE
            ? `${value}`
            : humanizeDataType(value, formatPreset, {
                excludeDecimalZeros: true,
              })}
      />
      <Grid />
      {#if allZeros || (mouseover && !allZeros) || !dataInDomain}
        <PointLabel
          showMovingPoint={!allZeros && dataInDomain}
          tweenProps={{ duration: 50 }}
          x={dataInDomain ? mouseover?.ts : undefined}
          y={dataInDomain ? mouseover?.[accessor] : undefined}
          format={allZeros || !dataInDomain
            ? () => "no data for this time range"
            : (value) =>
                formatPreset === NicelyFormattedTypes.NONE
                  ? value
                  : humanizeDataType(value, formatPreset)}
        />
      {/if}
    </SimpleDataGraphic>
  </div>
{/if}
