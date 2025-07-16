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

<Base {truncate} classes="{customStyle} {inTable && 'text-left'}" {color}>
  <slot name="value">
    {#if isNull}
      <span class=" italic">null</span>
    {:else}
      {formatDataType(value, type)}
    {/if}
  </slot>
</Base>
