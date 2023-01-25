<script lang="ts">
  import type { RichFormatNumber } from "./number-to-string-formatters";

  export let containerWidth = 100;

  export let richNum: RichFormatNumber;
  export let alignSuffix = false;
  export let suffixPadding = 0;

  export let lowerCaseEForEng = false;
  export let alignDecimalPoints = false;
  export let showBars = false;
  export let zeroHandling: "noSpecial" | "exactZero" | "zeroDot" = "noSpecial";
  export let negativeColor = "#f5999977";
  export let positiveColor = "#ececec";

  export let showBaseline = false;
  export let baselineColor = "#eeeeee";

  // export let numFormattingWidthLookup: { [key: string]: number };

  $: int = richNum.splitStr.int;
  $: frac = richNum.splitStr.frac;
  $: suffix = richNum.splitStr.suffix;

  // $: intChars = richNum.spacing.maxWholeDigits;
  // $: fracChars = richNum.spacing.maxFracDigits;
  // $: suffixChars = richNum.spacing.maxSuffixChars;

  // FINALIZE CHARACTERS TO BE DISPLAYED
  let suffixFinal;
  $: {
    // console.log({ lowerCaseEForEng });
    suffixFinal = suffix;
    if (lowerCaseEForEng) suffixFinal = suffixFinal.replace("E", "e");
  }

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

  $: suffixPadFinal = richNum.maxPxWidth.suffix > 0 ? suffixPadding : 0;

  $: intPx = richNum.maxPxWidth.int;
  $: dotPx = richNum.maxPxWidth.dot;
  $: fracPx = richNum.maxPxWidth.frac;
  $: suffixPx = richNum.maxPxWidth.suffix + suffixPadFinal;

  // $: containerWidth = `${intPx + dotPx + fracPx + suffixPx}px`;

  $: fracAndSuffixWidth = `${dotPx + fracPx + suffixPx}px`;

  $: logProps = () => {
    console.log({ ...richNum, lowerCaseEForEng });
  };

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

<div class="number-and-bar-container">
  {#if showBars}
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
  {/if}
  {#if !alignDecimalPoints}
    <div
      on:click={logProps}
      class="number-container"
      style="width: {containerWidth}px;"
    >
      {int}{decimalPoint}{frac}{suffixFinal}
    </div>
  {:else}
    <div
      on:click={logProps}
      class="number-container"
      style="width: {containerWidth}px;"
    >
      <div class="number-whole" style="width: {intPx}px;">
        {int}
      </div>
      {#if alignSuffix}
        <div class="number-frac" style="width: {dotPx + fracPx}px;">
          {decimalPoint}{frac}
        </div>

        <div
          class="number-suff"
          style="width: {suffixPx}px; padding-left: {suffixPadFinal}px"
        >
          {suffixFinal}
        </div>
      {:else}
        <div class="number-frac-and-suff" style="width: {fracAndSuffixWidth};">
          {decimalPoint}{frac}{suffixFinal}
        </div>
      {/if}
    </div>
  {/if}
</div>

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

  div.baseline {
    position: relative;
    top: 0px;
    height: 100%;
    width: 1px;
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
