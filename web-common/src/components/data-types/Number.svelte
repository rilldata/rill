<script lang="ts">
  import { formatDataType } from "../../lib/formatters";
  import Base from "./Base.svelte";

  export let isNull = false;
  export let inTable = false;
  export let customStyle = "";
  export let type: string;
  export let value;
  export let truncate = false;
  export let color = "";
</script>

<Base
  {truncate}
  {isNull}
  classes="ui-copy-number font-normal {inTable
    ? 'block text-right'
    : ''} {customStyle}"
  {color}
>
  <slot name="value">
    {formatDataType(value, type)}
  </slot>
</Base>
