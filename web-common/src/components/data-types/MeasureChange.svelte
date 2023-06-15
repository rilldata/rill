<script>
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let value;

  $: isNegative =
    (typeof value === "string" && value?.startsWith("-")) || value < 0;
</script>

<Base
  {isNull}
  classes="ui-copy-number font-normal {customStyle} {inTable &&
    'block text-right'}"
  {dark}
>
  <slot name="value">
    {!isNegative ? "+" : ""}{value}
  </slot>
</Base>
