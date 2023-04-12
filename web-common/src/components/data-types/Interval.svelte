<script>
  import { formatDataType } from "../../lib/formatters";
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let type;
  export let value;
</script>

<Base
  {isNull}
  classes="font-semibold truncate {inTable && 'block text-right'}"
  {dark}
>
  <slot name="value">
    {formatDataType(value, type)}
  </slot>
</Base>
