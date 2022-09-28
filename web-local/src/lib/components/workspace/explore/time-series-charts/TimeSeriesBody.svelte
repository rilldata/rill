<script lang="ts">
  import { extent } from "d3-array";
  import { interpolateArray } from "d3-interpolate";
  import { cubicOut } from "svelte/easing";
  import { derived, get, writable } from "svelte/store";
  import { fly } from "svelte/transition";
  import { guidGenerator } from "../../../../util/guid";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
  } from "../../../../util/humanize-numbers";
  import { Body } from "../../../data-graphic/elements";
  import SimpleDataGraphic from "../../../data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithTween } from "../../../data-graphic/functional-components";
  import { Axis, Grid, PointLabel } from "../../../data-graphic/guides";
  import { Area, Line } from "../../../data-graphic/marks";

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

  $: [_, yMax] = extent(dataCopy, (d) => d[accessor]);

  let yms = writable(yMax);
  $: yms.set(yMax);
  let prevValue = yMax;

  function previousStoreValue(anotherStore) {
    let previousValue = get(anotherStore);
    return derived(anotherStore, ($currentValue, set) => {
      set(previousValue);
      previousValue = $currentValue;
    });
  }

  const previousYMax = previousStoreValue(yms);
  // if $prev < yMax, do something
  // if $prev > yMax, do something else

  // need to control xMin, xMax, yMin, yMax.
</script>

{#if key && dataCopy?.length}
  <div transition:fly|local={{ duration: 500, y: 10 }}>
    <SimpleDataGraphic
      shareYScale={false}
      bind:mouseoverValue
      {yMin}
      {yMax}
      yMinTweenProps={{
        duration: longTimeSeries ? 0 : allZeros ? 100 : 300,
        delay: 200,
      }}
      yMaxTweenProps={{
        duration: longTimeSeries ? 0 : allZeros ? 100 : 300,
        delay: 200,
      }}
      let:xScale
    >
      <Body>
        {#key key + longTimeSeriesKey + $previousYMax > yMax}
          <!-- here, we switch hideCurrent before and after the transition, so
            in cases of the key updating, we can gracefully transition all kinds of
            interesting animations.
          -->
          <text x={30} y={30}>{yMax} - {$previousYMax}</text>
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
                delay: $previousYMax < yMax ? 500 : 0,
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
