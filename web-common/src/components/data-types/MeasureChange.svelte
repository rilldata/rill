<script lang="ts">
  import type { NumberParts } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
  import Base from "./Base.svelte";
  import { isPercDiff } from "./type-utils";

  export let inTable = false;
  export let customStyle = "";
  export let value: string | number | undefined | null | NumberParts;
  export let color = "!text-fg-primary";

  let isNull = false;
  let isValueNegative = false;
  let isValuePositive = false;

  $: if (isPercDiff(value)) {
    isNull = true;
  }

  // Determine if the value is negative for coloring purposes
  $: {
    isValueNegative = false;
    isValuePositive = false;

    if (value !== null && value !== undefined) {
      if (typeof value === "number") {
        isValueNegative = value < 0;
        isValuePositive = value > 0;
      } else if (typeof value === "object" && "neg" in value) {
        const intPart = Number(value.int || 0);
        const fracPart = Number(`0.${value.frac || "0"}`);
        const absoluteValue = intPart + fracPart;

        isValueNegative = value.neg === "-";
        isValuePositive = !isValueNegative && absoluteValue > 0;
      } else if (typeof value === "string") {
        const numericValue = Number(value.replaceAll(",", ""));
        isValueNegative = value.startsWith("-") || numericValue < 0;
        isValuePositive = !Number.isNaN(numericValue) && numericValue > 0;
      }
    }
  }
</script>

<Base
  {isNull}
  {color}
  classes="ui-copy-number w-full font-normal {customStyle} {inTable
    ? 'text-right'
    : ''}"
>
  {#if isValueNegative}
    <span class="text-kpi-negative">
      {value}
    </span>
  {:else if isValuePositive}
    <span class="text-kpi-positive">{value}</span>
  {:else}
    <span class="text-fg-secondary">{value}</span>
  {/if}
</Base>
