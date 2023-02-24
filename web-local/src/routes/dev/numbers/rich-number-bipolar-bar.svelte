<script lang="ts">
  import type { RichFormatNumber } from "./number-to-string-formatters";

  export let containerWidth = 100;

  export let richNum: RichFormatNumber;
  export let negativeColor = "#f5999977";
  export let positiveColor = "#ececec";

  export let showBaseline = false;
  export let baselineColor = "#eeeeee";
  export let barBackgroundColor = "#ffffff";

  export let absoluteValExtentsIfPosAndNeg = true;
  export let absoluteValExtentsAlways = false;

  export let reflectNegativeBars = false;

  let symmetricExtents: boolean;
  let absExtent: number;
  let validMin: number;
  let validMax: number;
  let barLeft: number;
  let barRight: number;

  $: {
    if (reflectNegativeBars) {
      // the min for the range is 0
      validMin = 0;
      // since reflecting, max is either true max or negative of min
      validMax = Math.max(-richNum.range.min, richNum.range.max);

      barLeft = 0;
      barRight = Math.abs(richNum.number);
    } else {
      symmetricExtents =
        absoluteValExtentsAlways ||
        (absoluteValExtentsIfPosAndNeg &&
          richNum.range.min < 0 &&
          richNum.range.max > 0);

      absExtent = symmetricExtents
        ? Math.max(-richNum.range.min, richNum.range.max)
        : 0;

      // if all the value are positive, the min for the range is 0
      validMin = Math.min(richNum.range.min, -absExtent);

      // if all the values are negative, the max for the range is 0
      validMax = Math.max(richNum.range.max, absExtent);

      barLeft = richNum.number < 0 ? richNum.number : 0;
      barRight = richNum.number < 0 ? 0 : richNum.number;
    }
  }

  // $: symmetricExtents =
  //   absoluteValExtentsAlways ||
  //   (absoluteValExtentsIfPosAndNeg &&
  //     richNum.range.min < 0 &&
  //     richNum.range.max > 0);

  // $: absExtent = symmetricExtents
  //   ? Math.max(-richNum.range.min, richNum.range.max)
  //   : 0;

  // // if all the value are positive, the min for the range is 0
  // $: validMin = Math.min(richNum.range.min, -absExtent);

  // // if all the values are negative, the max for the range is 0
  // $: validMax = Math.max(richNum.range.max, absExtent);

  // $: barLeft = richNum.number < 0 ? richNum.number : 0;
  // $: barRight = richNum.number < 0 ? 0 : richNum.number;

  const pctWithinExtents = (x, min, max) => 100 * ((x - min) / (max - min));

  $: barLeftPct = pctWithinExtents(barLeft, validMin, validMax);
  $: barWidthPct = pctWithinExtents(barRight, validMin, validMax) - barLeftPct;
  $: barColor = richNum.number < 0 ? negativeColor : positiveColor;

  $: baselineLeftPct = pctWithinExtents(0, validMin, validMax);
</script>

<div
  class="bar-container"
  style="width: {containerWidth}px; background-color:{barBackgroundColor}"
>
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
    height: 100%;
    top: 0px;
    right: 0px;
  }
</style>
