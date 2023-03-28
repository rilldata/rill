<script>
  import Base from "./Base.svelte";
  import { PERCENTAGE } from "./type-utils";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let value;
  export let type = "RILL_PERCENTAGE_CHANGE";

  $: diffIsNegative = value?.neg === "-";
</script>

<Base
  {isNull}
  classes="ui-copy-number font-normal {customStyle} {inTable &&
    'block text-right'}"
  {dark}
>
  <slot name="value">
    {#if value === PERCENTAGE.NO_DATA}
      <span class="opacity-50 italic">no data</span>
    {:else if value !== undefined}
      <span class:text-red-500={diffIsNegative}>
        {value?.neg || ""}{value?.int || ""}<span class="opacity-50"
          >{value?.percent || ""}</span
        >
      </span>
    {/if}
  </slot>
</Base>
