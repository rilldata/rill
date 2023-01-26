<script lang="ts">
  import { Body } from "@rilldata/web-common/components/data-graphic/elements";
  import SimpleDataGraphic from "@rilldata/web-common/components/data-graphic/elements/SimpleDataGraphic.svelte";
  import { WithTween } from "@rilldata/web-common/components/data-graphic/functional-components";
  import {
    Axis,
    Grid,
    PointLabel,
  } from "@rilldata/web-common/components/data-graphic/guides";
  import {
    Area,
    Line,
  } from "@rilldata/web-common/components/data-graphic/marks";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { previousValueStore } from "@rilldata/web-local/lib/store-utils";
  import { extent } from "d3-array";
  import { interpolateArray } from "d3-interpolate";
  import { cubicOut, linear } from "svelte/easing";
  import { writable } from "svelte/store";
  import { fade, fly } from "svelte/transition";
  import {
    humanizeDataType,
    NicelyFormattedTypes,
  } from "../../../../util/humanize-numbers";

  export let start;
  export let end;
  export let formatPreset: string;
  export let data;
  export let accessor: string;
  export let yMin = 0;

  $: formatPresetEnum =
    (formatPreset as NicelyFormattedTypes) || NicelyFormattedTypes.HUMANIZE;

  // the recycled mouseover event, in case anyone else has one set
  export let mouseover = undefined;
  // we use this is a key as well.
  export let timeGrain: string;
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

  $: allZeros = dataCopy.every((di) => di[accessor] === 0);
  $: dataInDomain = dataCopy.some((di) => di.ts >= start && di.ts <= end);

  $: [xMin, xMax] = extent(dataCopy, (d) => d.ts);
  $: [_, yMax] = extent(dataCopy, (d) => d[accessor]);

  let yms = writable(yMax);
  $: yms.set(yMax);

  $: timeRangeKey = xMin + xMax + dataCopy.length;

  let keyStore = writable(timeRangeKey);
  $: keyStore.set(timeRangeKey);
  let previousKeyStore = previousValueStore(keyStore);

  // get previous time grain so we can track whether we animate transitions with a scale-down
  // or a fade.
  let timeGrainStore = writable(timeGrain);
  $: timeGrainStore.set(timeGrain);
  let previousTimeGrain = previousValueStore(timeGrainStore);

  export function scaleVertical(
    node: Element,
    {
      delay = 0,
      duration = 400,
      easing = cubicOut,
      start = 0,
      opacity = 0,
      scaleDown = false,
    } = {}
  ) {
    const style = getComputedStyle(node);
    const target_opacity = +style.opacity;
    const transform = style.transform === "none" ? "" : style.transform;

    const sd = 1 - start;
    const od = target_opacity * (1 - opacity);

    return {
      delay,
      duration,
      easing,
      css: (_t, u) => {
        const yScale = scaleDown ? ` scaleY(${1 - sd * u})` : "";
        return `
    transform: ${transform} scaleY(${1 - sd * u}) ${yScale};
    transform-origin: 100% calc(100% - ${0}px);
    opacity: ${target_opacity - od * u}
  `;
      },
    };
  }

  const previousYMax = previousValueStore(yms);

  const scaleTweenDuration = 300;

  const lineTweenDuration = scaleTweenDuration;
  const lineTweenDelay = scaleTweenDuration * 1.3;

  /**
   * Plot animations
   * ======== Y Axis ========
   * There are two states to track:
   * new > old (taller) – usually when clearing filters
   * old > new (shorter) - usually when adding filters
   *  */

  // for now, just assume the y axis min value tween only functions in this one way.
  $: yMinTweenProps = {
    duration: longTimeSeries ? 0 : allZeros ? 100 : 500,
    delay: 200,
  };

  let lineTweenProps = { duration: 400, delay: 0 };

  // reactive variables for clarity
  $: newY = yMax;
  $: oldY = $previousYMax;
  $: isSmaller = newY < oldY;

  $: differentTimeRanges = $keyStore !== $previousKeyStore;
  $: diffTimeGrains = $previousTimeGrain !== $timeGrainStore;

  let xTweenProps = {
    duration: scaleTweenDuration * 2,
    delay: scaleTweenDuration,
  };
  let yMaxTweenProps = { duration: 400, delay: 0 };

  $: if (longTimeSeries) {
    /**
     *
     * case 1: long time series
     *
     */
    yMaxTweenProps = {
      duration: 0,
      delay: 0,
    };

    lineTweenProps = {
      duration: 0,
      // if new is larger than old, delay animation so the line does not
      // go off the page.
      delay: 0,
      interpolate: interpolateArray,
    };
  } else if (allZeros) {
    yMaxTweenProps = {
      duration: 100,
      delay: 0,
    };

    lineTweenProps = {
      duration: 0,
      // if new is larger than old, delay animation so the line does not
      // go off the page.
      delay: 0,
      interpolate: interpolateArray,
    };
  } else if (!isSmaller && !differentTimeRanges) {
    // We tween the yMax first, then the line.
    // this is to prevent the line from blowing past the plot extents
    // and being super weird.
    yMaxTweenProps = {
      duration: scaleTweenDuration,
      delay: 0,
      easing: linear,
    };
    lineTweenProps = {
      duration: lineTweenDuration,
      // if new is larger than old, delay animation so the line does not
      // go off the page.
      delay: lineTweenDelay,
      easing: cubicOut,
      interpolate: interpolateArray,
    };
  } else if (isSmaller && !differentTimeRanges) {
    // we can tween the yMax and the line at the same time, since there is no risk of clipping the area chart.
    yMaxTweenProps = {
      duration: scaleTweenDuration,
      delay: 0,
      easing: linear,
    };

    lineTweenProps = {
      duration: scaleTweenDuration,
      // if new is larger than old, delay animation so the line does not
      // go off the page.
      delay: 0,
      easing: cubicOut,
      interpolate: interpolateArray,
    };
  } else {
    lineTweenProps = {
      duration: lineTweenDuration,
      interpolate: interpolateArray,
    };
  }

  let timeout;
  /** Clear out the previousKeyStore and previousTimeGrain */
  $: setTimeout(() => {
    clearTimeout(timeout);
    previousKeyStore.set($keyStore);
    previousTimeGrain.set($timeGrainStore);
  }, scaleTweenDuration * 2);
</script>

{#if timeRangeKey && dataCopy?.length}
  <div transition:fly|local={{ duration: 500, y: 10 }}>
    <SimpleDataGraphic
      shareYScale={false}
      bind:mouseoverValue
      yMin={yMin > 0 ? 0 : yMin}
      {yMax}
      {yMinTweenProps}
      {yMaxTweenProps}
      xMinTweenProps={xTweenProps}
      xMaxTweenProps={xTweenProps}
    >
      <Body>
        {#key timeGrain}
          <!-- this key will trigger the scale changes.
            We typically only trigger scale changes when the date ranges change
          -->
          <g
            in:scaleVertical|local={{
              duration: scaleTweenDuration,
              delay: 0,
              //diffTimeGrains && !differentTimeRanges ? scaleTweenDuration : 0,
              start: 0,
              scaleDown: diffTimeGrains,
            }}
            out:scaleVertical|local={{
              duration: scaleTweenDuration,
              delay: 0,
              start: 0,
              scaleDown: diffTimeGrains,
            }}
            style:transition="opacity 250ms"
          >
            {#key timeRangeKey}
              <g transition:fade|local={{ duration: scaleTweenDuration }}>
                <WithTween
                  value={dataCopy}
                  let:output={tweenedData}
                  tweenProps={lineTweenProps}
                >
                  <Area
                    data={tweenedData}
                    yAccessor={accessor}
                    xAccessor="ts"
                  />
                  <Line
                    data={tweenedData}
                    yAccessor={accessor}
                    xAccessor="ts"
                  />
                </WithTween>
              </g>
            {/key}
          </g>
        {/key}
      </Body>
      <Axis
        side="right"
        format={(value) =>
          formatPreset === NicelyFormattedTypes.NONE
            ? `${value}`
            : humanizeDataType(value, formatPresetEnum, {
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
                  : humanizeDataType(value, formatPresetEnum)}
        />
      {/if}
    </SimpleDataGraphic>
  </div>
{/if}
