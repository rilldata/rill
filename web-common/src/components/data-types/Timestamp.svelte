<script lang="ts">
  import { formatDataType } from "../../lib/formatters";
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export const textAlign = "text-right";
  export let customStyle = "";
  export let type: string;
  export let value;
  export let truncate: boolean;
  export let color = "";
</script>

<Base
  {truncate}
  {isNull}
  classes="ui-copy-number font-semibold {customStyle} text-foreground {inTable &&
    'text-left'}"
  {color}
>
  <slot name="value">
    {formatDataType(value, type)}
  </slot>
</Base>
