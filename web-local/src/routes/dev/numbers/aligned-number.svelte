<script lang="ts">
  import type { RichFormatNumber } from "./number-to-string-formatters";

  export let richNum: RichFormatNumber;
  export let alignSuffix = false;
  export let lowerCaseEForEng = false;
  export let alignDecimalPoints = false;
  export let showBars = false;
  export let zeroHandling: "noSpecial" | "exactZero" | "zeroDot" = "noSpecial";
  export let negativeColor = "#f5999977";
  export let positiveColor = "#ececec";

  export let showBaseline = false;
  export let baselineColor = "#eeeeee";

  $: whole = richNum.splitStr.int;
  $: frac = richNum.splitStr.frac;
  $: suffix = richNum.splitStr.suffix;

  $: wholeChars = richNum.spacing.maxWholeDigits;
  $: fracChars = richNum.spacing.maxFracDigits;
  $: suffixChars = richNum.spacing.maxSuffixChars;

  // IMPORTANT: add a bit of width to the containers for the decimal point
  const DECIMAL_POINT_WIDTH = 0.6;
  // IMPORTANT: add a bit of width to the int part to ensure decimal alignment
  const EXTRA_INT_WIDTH = 0.5;
  // a bit of padding between the suffix and the number
  const SUFFIX_PADDING = 0.2;
  let suffixFinal;
  $: {
    // console.log({ lowerCaseEForEng });
    suffixFinal = suffix;
    if (lowerCaseEForEng) suffixFinal = suffixFinal.replace("E", "e");
  }

  $: containerWidth = `calc(${wholeChars + EXTRA_INT_WIDTH}ch + ${
    fracChars + DECIMAL_POINT_WIDTH
  }ch + ${suffixChars + SUFFIX_PADDING}em)`;

  $: fracAndSuffixWidth = alignSuffix
    ? ""
    : `calc( ${fracChars + DECIMAL_POINT_WIDTH}ch + ${
        suffixChars + SUFFIX_PADDING
      }em)`;

  let decimalPoint: "" | ".";

  $: {
    decimalPoint = frac !== "" ? "." : "";
    if (richNum.number === 0) {
      if (zeroHandling === "exactZero") {
        decimalPoint = "";
        frac = "";
      } else if (zeroHandling === "zeroDot") {
        decimalPoint = ".";
        frac = "";
      }
    }
  }

  $: logProps = () => {
    console.log({ ...richNum, lowerCaseEForEng });
  };

  // if any number is negative, include negative bars;
  // otherwise all bars range from [0, max]
  // $: rangeWidth = richNum.range.max - Math.min(richNum.range.min, 0);

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
</script>

{#if !alignDecimalPoints}
  <div class="number-and-bar-container">
    <div
      on:click={logProps}
      class="number-container"
      style="width: {containerWidth};"
    >
      {whole}{decimalPoint}{frac}{suffixFinal}
    </div>
  </div>
{:else}
  <div class="number-and-bar-container">
    {#if showBars}
      <div class="bar-container" style="width: {containerWidth};">
        <div
          class="number-bar"
          style="left:{barLeftPct}%; width: {barWidthPct}%; background-color:{barColor};"
        />
      </div>
    {/if}

    <div
      on:click={logProps}
      class="number-container"
      style="width: {containerWidth};"
    >
      <div
        class="number-whole"
        style="width: {wholeChars + EXTRA_INT_WIDTH}ch;"
      >
        {whole}
      </div>
      {#if alignSuffix}
        <div
          class="number-frac"
          style="width: {fracChars + DECIMAL_POINT_WIDTH}ch;"
        >
          {decimalPoint}{frac}
        </div>

        <div
          class="number-suff"
          style="width: {suffixChars +
            SUFFIX_PADDING}em; padding-left: {SUFFIX_PADDING}em;"
        >
          {suffixFinal}
        </div>
      {:else}
        <div class="number-frac-and-suff" style="width: {fracAndSuffixWidth};">
          {decimalPoint}{frac}{suffixFinal}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  div.number-and-bar-container {
    position: relative;
  }

  div.number-bar {
    position: relative;
    top: 0px;
    height: 100%;
    /* z-index: 5; */
  }

  div.number-container {
    display: flex;
    flex-direction: row;
    justify-content: flex-end;
    flex-wrap: nowrap;
    white-space: nowrap;
    overflow: hidden;
    position: relative;
    /* z-index: 10; */
    /* outline: 1px solid black; */
  }

  div.number-whole {
    text-align: right;
  }
  div.number-frac {
    text-align: left;
  }

  div.number-suff {
    text-align: left;
  }

  div.number-frac-and-suff {
    text-align: left;
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
