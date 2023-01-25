<script lang="ts">
  import type { RichFormatNumber } from "./number-to-string-formatters";
  import RichNumberBipolarBar from "./rich-number-bipolar-bar.svelte";

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
  export let barBackgroundColor = "#ffffff";

  export let showBaseline = false;
  export let baselineColor = "#eeeeee";

  $: int = richNum.splitStr.int;
  $: frac = richNum.splitStr.frac;
  $: suffix = richNum.splitStr.suffix;

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
</script>

<div class="number-and-bar-container">
  {#if showBars}
    <RichNumberBipolarBar
      {richNum}
      {containerWidth}
      {positiveColor}
      {negativeColor}
      {showBaseline}
      {baselineColor}
      {barBackgroundColor}
    />
  {/if}
  <div
    on:click={logProps}
    class="number-container"
    style="width: {containerWidth}px;"
  >
    {#if !alignDecimalPoints}
      {int}{decimalPoint}{frac}{suffixFinal}
    {:else}
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
    {/if}
  </div>
</div>

<style>
  div.number-and-bar-container {
    position: relative;
  }

  div.number-container {
    display: flex;
    flex-direction: row;
    justify-content: flex-end;
    flex-wrap: nowrap;
    white-space: nowrap;
    overflow: hidden;
    position: relative;
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
