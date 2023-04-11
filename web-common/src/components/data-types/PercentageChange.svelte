<script>
  import Base from "./Base.svelte";
  import { PERC_DIFF } from "./type-utils";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let value;
  export let type = "RILL_PERCENTAGE_CHANGE";
  export let tabularNumber = true;

  $: diffIsNegative = value?.neg === "-";
  $: intValue = value?.int ? value?.int : value?.int === 0 ? 0 : "";
</script>

<Base
  {isNull}
  classes="{tabularNumber
    ? 'ui-copy-number'
    : ''} font-normal {customStyle} {inTable && 'block text-right'}"
  {dark}
>
  <slot name="value">
    {#if value === PERC_DIFF.PREV_VALUE_NO_DATA || value === PERC_DIFF.PREV_VALUE_NULL}
      <span class="opacity-50 italic" style:font-size=".925em">no data</span>
    {:else if value !== undefined}
      <span class:text-red-500={diffIsNegative}>
        {value?.neg || ""}{intValue}<span class="opacity-50"
          >{value?.percent || ""}</span
        >
      </span>
    {/if}
  </slot>
</Base>
