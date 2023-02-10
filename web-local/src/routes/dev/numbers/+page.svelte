<script lang="ts">
  import TableExampleWidget from "./table-example-widget.svelte";
  import StringMeasurer from "./string-measurer.svelte";
  import { runTestsmallestPrecisionMagnitude } from "./smallest-precision-magnitude";
  runTestsmallestPrecisionMagnitude();
  import { runTest_formatNumWithOrderOfMag2 } from "./format-with-order-of-magnitude";
  runTest_formatNumWithOrderOfMag2();
</script>

<h1 class="pb-4">Tabular / columnar number formatting</h1>

<TableExampleWidget defaultFormatterIndex={2} />

<TableExampleWidget
  alignDecimalPoints={false}
  alignSuffixes={false}
  showMagSuffixForZero={true}
  zeroHandling="noSpecial"
/>

<h1 class="pb-4 pt-10">string widths</h1>
Worst case if we always want to use multiple of 3 exponents
<StringMeasurer string="-$123e-303" />
<StringMeasurer string="-123e-303%" />

Worst case if we allow non-multiple of 3 exponents for infinitesimals requiring
a 3 digit exponent
<StringMeasurer string="-$1e-301" />
<StringMeasurer string="-1e-301%" />

Worst case if we require multiple of 3 exponents for infinitesimals requiring a
2 digit exponent
<StringMeasurer string="-$123e-99" />
<StringMeasurer string="-123e-99%" />

<style>
  h1 {
    font-size: large;
  }
</style>
