<script>
  import { formatDataType } from "../../lib/formatters";
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let value;
  export let type = "RILL_CHANGE";

  $: isNegative = typeof value === "string" && value?.startsWith("-");
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
