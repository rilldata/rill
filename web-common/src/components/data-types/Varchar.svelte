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

<Base {truncate} classes="{customStyle} {inTable && 'text-left'}" {dark}>
  <slot name="value">
    {#if isNull}
      <span class="opacity-60 italic">null</span>
    {:else}
      {formatDataType(value, type)}
    {/if}
  </slot>
</Base>
