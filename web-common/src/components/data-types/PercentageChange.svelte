<script lang="ts">
  import Base from "./Base.svelte";
  import { PERC_DIFF } from "./type-utils";
  import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let value: null | NumberParts | PERC_DIFF;
  export let tabularNumber = true;

  let diffIsNegative = false;
  let intValue: string;
  let negSign = "";
  let approxSign = "";
  let suffix = "";

  const isPercDiff = (token: unknown): token is PERC_DIFF[keyof PERC_DIFF] =>
    Object.values(PERC_DIFF).includes(token as PERC_DIFF);

  $: isNoData = isPercDiff(value) || value === null;

  $: if (!isNoData && !isPercDiff(value) && value !== null) {
    // in this case, we have a NumberParts object.
    // We have a couple cases to consider:
    // * If the NumberParts object has approxZero===true,
    // we want to show e.g. "~0%" WITHOUT a negative sign
    // * However, in this case we show the number in red to indicate a
    // small negative change.
    //
    // Otherwise, we format the number as usual.
    intValue = value.int;
    diffIsNegative = value?.neg === "-";
    negSign = diffIsNegative && !value?.approxZero ? "-" : "";
    approxSign = value?.approxZero ? "~" : "";
    suffix = value?.suffix ?? "";
  }
</script>

<Base
  {isNull}
  classes="{tabularNumber
    ? 'ui-copy-number'
    : ''} font-normal {customStyle} {inTable && 'block text-right'}"
  {dark}
>
  <slot name="value">
    {#if isNoData}
      <span class="opacity-50 italic" style:font-size=".925em">no data</span>
    {:else if value !== null}
      <span class:text-red-500={diffIsNegative}>
        {approxSign}{negSign}{intValue}{suffix}<span class="opacity-50">%</span>
      </span>
    {/if}
  </slot>
</Base>
