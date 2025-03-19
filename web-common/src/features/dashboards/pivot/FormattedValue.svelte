<script lang="ts">
  import { createMeasureValueFormatter } from "@rilldata/web-common/lib/number-formatting/format-measure-value";
  //   import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";

  export let rawValue: string | number | null | undefined | unknown;
  export let formattedValue: string | undefined;
  export let type:
    | "comparison_percent"
    | "comparison_delta"
    | "measure"
    | undefined = undefined;
  export let columnId: string | undefined = undefined;
  export let rowId: string | undefined = undefined;
  export let formatter: ReturnType<typeof createMeasureValueFormatter> = (
    value,
  ) => value?.toString();
  export let showRightBorder = false;
  //   export let meta;
  //   export let isComparisonType = false;

  //   $: isNull = rawValue === undefined || rawValue === null;

  //   $: if (!formatter) console.log(meta, rawValue);

  //   $: formattedValue = !formatter
  //     ? rawValue
  //     : type === "comparison_percent"
  //       ? formatMeasurePercentageDifference(rawValue)
  //       : formatter?.(rawValue);

  //   function formatMeasurePercentageDifference(value) {
  //     if (value === null || value === undefined) return "-";
  //     if (value < 0.005 && value > -0.005) return "0%";
  //     return `${(value * 100).toFixed(2)}%`;
  //   }
</script>

<td
  class="ui-copy-number truncate text-gray-800"
  data-rowid={rowId}
  data-columnid={columnId}
  data-value={rawValue}
  class:!text-gray-400={rawValue === undefined || rawValue == null}
  class:border-r={showRightBorder}
  class:!text-red-500={rawValue < 0}
>
  {formattedValue ?? rawValue ?? "-"}
</td>

<style lang="postcss">
  td {
    @apply m-0 size-full p-1 px-2;
    @apply text-right;
    @apply whitespace-nowrap text-xs;
    height: var(--row-height);
  }

  td:last-of-type {
    @apply border-r-0;
  }

  :global(.with-row-dimension tr) > td:first-of-type {
    @apply sticky left-0 z-10;
    @apply bg-white;
  }

  :global(tbody > tr:nth-of-type(2)) > td:first-of-type {
    @apply font-semibold;
  }

  :global(tbody tr:hover) td:first-of-type {
    @apply bg-slate-100;
  }
</style>
