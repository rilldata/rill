<script lang="ts">
  import Base from "./Base.svelte";
  import { PERC_DIFF } from "./type-utils";
  import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let value: undefined | NumberParts | PERC_DIFF;
  export let tabularNumber = true;

  let diffIsNegative = false;
  let intValue: string;
  let negSign = "";
  let approxSign = "";
  let suffix = "";

  $: isNoData = Object.values(PERC_DIFF).includes(value) || value === undefined;

  $: if (!isNoData) {
    // in this case, we have a NumberParts object.
    // We have a couple cases to consider:
    // * If the NumberParts object has approxZero===true,
    // we want to show e.g. "~0" WITHOUT a negative sign,
    // even if the NumberParts object has a negative sign,
    // we do want to show the number in red to indicate a
    // small negative change.

    intValue = value?.int;
    diffIsNegative = value?.neg === "-";
    negSign = diffIsNegative && !value?.approxZero ? "-" : "";
    approxSign = value?.approxZero ? "~" : "";
    suffix = value?.suffix ?? "";

    // This formatter should only ever recieve a NumberParts object
    // with a value.percent === "%" field. If that invariant fails,
    // we don't want to crash, but we'll emit a warning.
    if (value?.percent !== "%") {
      console.warn(
        `PercentageChange component expects a NumberParts object with a percent sign, received ${JSON.stringify(
          value
        )} instead.`
      );
    }
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
    {:else if value !== undefined}
      <span class:text-red-500={diffIsNegative}>
        {approxSign}{negSign}{intValue}{suffix}<span class="opacity-50">%</span>
      </span>
    {/if}
  </slot>
</Base>
