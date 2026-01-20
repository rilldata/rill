<script lang="ts">
  import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import Base from "./Base.svelte";
  import { isPercDiff } from "./type-utils";

  export let inTable = false;
  export let customStyle = "";
  export let value: string | number | undefined | null | NumberParts;

  let isNull = false;
  let isValueNegative = false;

  $: if (isPercDiff(value)) {
    isNull = true;
  }

  // Determine if the value is negative for coloring purposes
  $: if (value !== null && value !== undefined) {
    if (typeof value === "number") {
      isValueNegative = value < 0;
    } else if (typeof value === "object" && "neg" in value) {
      // For NumberParts, check if it has a negative sign
      isValueNegative = value.neg === "-";
    } else if (typeof value === "string") {
      // For strings, check if it starts with a minus sign
      isValueNegative = value.startsWith("-");
    }
  }
</script>

<Base
  {isNull}
  classes="ui-copy-number w-full font-normal {customStyle} {inTable
    ? 'text-right'
    : ''}"
>
  {#if isValueNegative}
    <span class="text-red-500">
      {value}
    </span>
  {:else}
    <span class="text-fg-secondary">{value}</span>
  {/if}
</Base>
