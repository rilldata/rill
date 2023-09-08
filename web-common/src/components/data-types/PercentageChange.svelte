<script lang="ts">
  import Base from "./Base.svelte";
  import { PERC_DIFF } from "./type-utils";
  import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let value: NumberParts | PERC_DIFF.PREV_VALUE_NO_DATA;
  export let tabularNumber = true;

  $: diffIsNegative = value?.neg === "-";
  let intValue: number | string;
  $: if (typeof value === "string" || value === PERC_DIFF.PREV_VALUE_NO_DATA) {
    // NO-OP
  } else {
    intValue = value?.int;
  }
  intValue = value?.int ? value?.int : value?.int === 0 ? 0 : "";
</script>

<Base
  {isNull}
  classes="{tabularNumber
    ? 'ui-copy-number'
    : ''} font-normal {customStyle} {inTable && 'block text-right'}"
  {dark}
>
  <slot name="value">
    {#if value === PERC_DIFF.PREV_VALUE_NO_DATA}
      <span class="opacity-50 italic" style:font-size=".925em">no data</span>
    {:else if value !== undefined}
      <span class:text-red-500={diffIsNegative}>
        {value?.neg || ""}{intValue}{value.suffix}<span class="opacity-50"
          >{value?.percent || ""}</span
        >
      </span>
    {/if}
  </slot>
</Base>
