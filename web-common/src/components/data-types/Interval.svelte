<script lang="ts">
  import { formatDuckdbIntervalLossless } from "@rilldata/web-common/lib/number-formatting/strategies/intervals";
  import Base from "./Base.svelte";
  export let isNull = false;
  export let inTable = false;
  export let value;
  export let truncate = false;
  export let color = "";
</script>

<Base
  {truncate}
  {isNull}
  classes="font-semibold truncate w-full {inTable && 'block text-right'}"
  {color}
>
  <slot name="value">
    {value ? formatDuckdbIntervalLossless(value) : ""}
  </slot>
</Base>
