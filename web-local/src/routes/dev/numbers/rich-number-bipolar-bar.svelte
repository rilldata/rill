<script lang="ts">
  import type { RichFormatNumber } from "./number-to-string-formatters";

  export let containerWidth = 100;

  export let richNum: RichFormatNumber;
  export let negativeColor = "#f5999977";
  export let positiveColor = "#ececec";

  export let showBaseline = false;
  export let baselineColor = "#eeeeee";

  // if all the value are positive, the min for the range is 0
  $: validMin = Math.min(richNum.range.min, 0);

  // if all the values are negative, the max for the range is 0
  $: validMax = Math.max(richNum.range.max, 0);

  $: barLeft = richNum.number < 0 ? richNum.number : 0;
  $: barRight = richNum.number < 0 ? 0 : richNum.number;

  const pctWithinExtents = (x, min, max) => 100 * ((x - min) / (max - min));

  $: barLeftPct = pctWithinExtents(barLeft, validMin, validMax);
  $: barWidthPct = pctWithinExtents(barRight, validMin, validMax) - barLeftPct;
  $: barColor = richNum.number < 0 ? negativeColor : positiveColor;

  $: baselineLeftPct = pctWithinExtents(0, validMin, validMax);
</script>

<div class="bar-container" style="width: {containerWidth}px;">
  <div
    class="number-bar"
    style="left:{barLeftPct}%; width: {barWidthPct}%; background-color:{barColor};"
  />
</div>
{#if showBaseline}
  <div class="bar-container" style="width: {containerWidth}px;">
    <div
      class="baseline"
      style="left:{baselineLeftPct}%; background-color:{baselineColor};"
    />
  </div>
{/if}

<style>
  div.number-bar {
    position: relative;
    top: 0px;
    height: 100%;
    /* z-index: 5; */
  }

  div.baseline {
    position: relative;
    top: 0px;
    height: 100%;
    width: 1px;
    /* z-index: 5; */
  }

  div.bar-container {
    display: block;
    position: absolute;
    width: 100%;
    height: 100%;
    /* background-color: rgba(255, 0, 225, 0.29); */
    top: 0px;
    left: 0px;
  }
</style>
