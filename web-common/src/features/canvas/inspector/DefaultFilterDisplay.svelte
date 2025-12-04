<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getCanvasStore } from "../state-managers/state-managers";
  import DimensionFilterReadOnlyChip from "../../dashboards/filters/dimension-filters/DimensionFilterReadOnlyChip.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import MeasureFilterReadOnlyChip from "../../dashboards/filters/measure-filters/MeasureFilterReadOnlyChip.svelte";
  import TimeRangeReadOnly from "../../dashboards/filters/TimeRangeReadOnly.svelte";

  export let canvasName: string;

  $: ({ instanceId } = $runtime);

  $: ({
    canvasEntity: {
      clearDefaultFilters,
      filterManager: { _defaultUIFilters },
      timeControls: {
        interval: _interval,
        _defaultTimeRange,
        _defaultComparisonRange,
      },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: ({ dimensionFilters, measureFilters } = $_defaultUIFilters);

  $: interval = $_interval;

  $: defaultTimeRange = $_defaultTimeRange;

  $: defaultComparisonRange = $_defaultComparisonRange;
</script>

<div class="flex-col flex h-full">
  <div class="page-param">
    <p class="text-muted-foreground mb-4">
      The filters listed below are saved as your default view and will
      automatically apply each time you open this dashboard in Rill Cloud.
    </p>

    <div class="flex flex-col gap-y-2 gap-x-2 w-full flex-none">
      <div class="flex gap-x-2">
        {#if defaultTimeRange}
          <TimeRangeReadOnly
            timeRange={{ expression: defaultTimeRange }}
            comparisonTimeRange={defaultComparisonRange
              ? { expression: defaultComparisonRange }
              : undefined}
          />
        {/if}
      </div>

      {#each dimensionFilters as [id, filterData] (id)}
        {@const metricsViewNames = Array.from(filterData.dimensions.keys())}
        {@const dimension = filterData.dimensions.get(metricsViewNames[0])}

        {#if dimension && dimension.name}
          <DimensionFilterReadOnlyChip
            pinned={filterData.pinned}
            name={dimension.name}
            {metricsViewNames}
            label={dimension.displayName ||
              dimension.name ||
              dimension.column ||
              "Unnamed Dimension"}
            mode={filterData.mode}
            values={filterData.selectedValues ?? []}
            inputText={filterData.inputText}
            isInclude={filterData.isInclude === true}
            timeStart={interval?.start?.toISO()}
            timeEnd={interval?.end?.toISO()}
          />
        {/if}
      {/each}

      {#each measureFilters as [id, filterData] (id)}
        {@const metricsViewNames = Array.from(
          filterData?.measures?.keys() ?? [],
        )}
        {@const measure = filterData.measures?.get(metricsViewNames[0])}

        {#if measure && measure.name}
          <MeasureFilterReadOnlyChip
            pinned={filterData.pinned}
            dimensionName={filterData.dimensionName}
            label={filterData.label}
            filter={filterData.filter}
          />
        {/if}
      {/each}
    </div>
  </div>

  <div class="mt-auto border-t w-full px-5 py-3">
    <Button type="secondary" wide onClick={clearDefaultFilters}>
      <Trash />
      Clear default filters
    </Button>
  </div>
</div>

<style lang="postcss">
  .page-param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
