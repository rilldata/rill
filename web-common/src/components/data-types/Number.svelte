<script>
  import { formatDataType } from "../../lib/formatters";
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let customStyle = "";
  export let type;
  export let value;
  export let truncate = false;
</script>

<Base
  {truncate}
  {isNull}
  classes="ui-copy-number font-normal {customStyle} {inTable &&
    'block text-right'}"
  {dark}
>
  <slot name="value">
    {formatDataType(value, type)}
  </slot>
</Base>
