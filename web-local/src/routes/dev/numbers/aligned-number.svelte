<script lang="ts">
  import type { RichFormatNumber } from "./number-to-string-formatters";

  export let richNum: RichFormatNumber;
  export let alignSuffix = false;
  export let lowerCaseEForEng = false;
  export let alignDecimalPoints = false;
  export let zeroHandling: "noSpecial" | "exactZero" | "zeroDot" = "noSpecial";

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
</script>

{#if !alignDecimalPoints}
  <div on:click={logProps} class="number-container">
    {whole}{decimalPoint}{frac}{suffixFinal}
  </div>
{:else}
  <div
    on:click={logProps}
    class="number-container"
    style="width: {containerWidth};"
  >
    <div class="number-whole" style="width: {wholeChars + EXTRA_INT_WIDTH}ch;">
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
{/if}

<style>
  div.number-container {
    display: flex;
    flex-direction: row;
    justify-content: flex-end;
    flex-wrap: nowrap;
    white-space: nowrap;
    overflow: hidden;
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
</style>
