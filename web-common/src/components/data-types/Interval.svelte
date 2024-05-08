<script>
  import { formatDuckdbIntervalLossless } from "@rilldata/web-common/lib/number-formatting/strategies/intervals";
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export let dark = false;
  export let value;
  export let truncate = false;
</script>

<Base
  {truncate}
  {isNull}
  classes="font-semibold truncate w-full {inTable && 'block text-right'}"
  {dark}
>
  <slot name="value">
    {value ? formatDuckdbIntervalLossless(value) : ""}
  </slot>
</Base>
