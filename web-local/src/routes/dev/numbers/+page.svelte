<script lang="ts">
  import TableExampleWidget from "./table-example-widget.svelte";
  import SpanMeasurer from "./span-measurer.svelte";
  import { runTestsmallestPrecisionMagnitude } from "./smallest-precision-magnitude";
  runTestsmallestPrecisionMagnitude();
  import { runTest_formatNumWithOrderOfMag2 } from "./format-with-order-of-magnitude";
  import { onMount } from "svelte";
  runTest_formatNumWithOrderOfMag2();

  let fontsReady = false;
  // document.fonts.onloadingdone = () => {
  //   console.log("onloadingdone", { fontsReady });
  //   fontsReady = true;
  // };

  onMount(() => {
    fontsReady = fontsReady || document.fonts.check("12px Inter");
    document.fonts.onloadingdone = () => {
      fontsReady = true;
      console.log("onloadingdone", { fontsReady });
    };

    setTimeout(() => {
      fontsReady = fontsReady || document.fonts.check("12px Inter");
      console.log({ fontsReady });
    }, 1300);
  });
  $: console.log({ fontsReady });
</script>

<h1 class="pb-4">Tabular / columnar number formatting</h1>
{#if fontsReady}
  <TableExampleWidget defaultFormatterIndex={2} />

  <div style="height: 80px;" />

  <TableExampleWidget
    alignDecimalPoints={false}
    alignSuffixes={false}
    showMagSuffixForZero={true}
    zeroHandling="noSpecial"
  />

  <h1 class="pb-4 pt-10">string widths</h1>
  <div class="ui-copy-number">
    Worst case if we always want to use multiple of 3 exponents
    <SpanMeasurer>-$123e-303</SpanMeasurer>
    <SpanMeasurer>-123e-303%</SpanMeasurer>

    Worst case if we allow non-multiple of 3 exponents for infinitesimals
    requiring a 3 digit exponent
    <SpanMeasurer>-$1e-301</SpanMeasurer>
    <SpanMeasurer>-1e-301%</SpanMeasurer>

    Worst case if we require multiple of 3 exponents for infinitesimals
    requiring a 2 digit exponent
    <SpanMeasurer>-$123e-99</SpanMeasurer>
    <SpanMeasurer>-123e-99%</SpanMeasurer>

    <br />
    Does wrapping each character in a span matter for calculating widths?
    <br />
    (no spans) <SpanMeasurer>498.897e-15</SpanMeasurer>

    (with spans) <SpanMeasurer>
      {#each "498.897e-15".split("") as char}
        <span>{char}</span>
      {/each}
    </SpanMeasurer>

    (with spans) <SpanMeasurer>
      {#each "111.111e-11".split("") as char}
        <span>{char}</span>
      {/each}
    </SpanMeasurer>

    (with spans) <SpanMeasurer>
      {#each "000.000e-00".split("") as char}
        <span>{char}</span>
      {/each}
    </SpanMeasurer>
  </div>

  <style>
    h1 {
      font-size: large;
    }
  </style>
{/if}
