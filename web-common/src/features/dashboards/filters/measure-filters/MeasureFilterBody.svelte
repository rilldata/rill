<!-- @component
  renders the body content of a filter set chip:
  - a label for the current measure
  - a short hand notation of the filter criteria
-->
<script lang="ts">
  import type { MeasureFilterEntry } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import {
    AllMeasureFilterOperationOptions,
    AllMeasureFilterTypeOptions,
    MeasureFilterOperation,
    MeasureFilterType,
  } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";

  export let dimensionName: string;
  export let label: string | undefined;
  export let filter: MeasureFilterEntry | undefined;
  export let labelMaxWidth = "320px";
  export let comparisonLabel = "";

  let typeLabel: string | undefined;
  let shortLabel: string | undefined;
  $: if (filter) {
    const typeOption = AllMeasureFilterTypeOptions.find(
      (o) => o.value === filter?.type,
    );
    typeLabel = typeOption?.shortLabel;

    if (
      filter.type === MeasureFilterType.AbsoluteChange ||
      filter.type === MeasureFilterType.PercentChange
    ) {
      typeLabel += ` from ${comparisonLabel}`;
    }

    switch (filter.operation) {
      case MeasureFilterOperation.GreaterThan:
      case MeasureFilterOperation.GreaterThanOrEquals:
      case MeasureFilterOperation.LessThan:
      case MeasureFilterOperation.LessThanOrEquals:
      case MeasureFilterOperation.Equals:
      case MeasureFilterOperation.NotEquals:
        shortLabel =
          AllMeasureFilterOperationOptions.find(
            (o) => o.value === filter?.operation,
          )?.shortLabel +
          " " +
          filter.value1 +
          (filter.type === MeasureFilterType.PercentChange ? "%" : "");
        break;
      case MeasureFilterOperation.Between:
        shortLabel = `(${filter.value1},${filter.value2})`;
        break;
      case MeasureFilterOperation.NotBetween:
        shortLabel = `!(${filter.value1},${filter.value2})`;
        break;
    }
  }
</script>

<div class="flex gap-x-2">
  <div
    class="font-bold text-ellipsis overflow-hidden whitespace-nowrap"
    style:max-width={labelMaxWidth}
  >
    {label}
    {#if dimensionName}
      <!-- span needed to make sure the space before the `for` is not removed by prettier -->
      <span> for {dimensionName}</span>
    {/if}
    {#if typeLabel}
      <span>{typeLabel}</span>
    {/if}
  </div>
  <div class="flex flex-wrap flex-row items-baseline gap-y-1">
    {#if shortLabel}
      {shortLabel}
    {/if}
  </div>
</div>
