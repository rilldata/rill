<script lang="ts">
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { mapExprToMeasureFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
  import MeasureFilterBody from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterBody.svelte";
  import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
  import type {
    V1Expression,
    V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";

  export let filters: V1Expression | undefined;
  export let comparisonTimeRange: V1TimeRange | undefined;

  $: filtersLength = filters?.cond?.exprs?.length ?? 0;

  $: measureFilters = filters?.cond?.exprs?.map(mapExprToMeasureFilter) ?? [];

  $: comparisonLabel =
    TIME_COMPARISON[comparisonTimeRange?.isoOffset]?.label?.toLowerCase();
</script>

<div class="flex flex-col gap-y-3">
  <MetadataLabel>Criteria</MetadataLabel>
  <div class="flex flex-wrap gap-2">
    {#if filtersLength}
      {#each measureFilters as filter, index (index)}
        <div animate:flip={{ duration: 200 }}>
          <Chip type="measure" label={filter.measure} readOnly>
            <div class="mx-2" slot="body">
              <MeasureFilterBody
                dimensionName=""
                {filter}
                label={filter.measure}
                {comparisonLabel}
                labelMaxWidth=""
              />
            </div>
          </Chip>
        </div>
      {/each}
    {:else}
      <div
        in:fly|local={{ duration: 200, x: 8 }}
        class="ui-copy-disabled grid items-center"
        style:min-height="26px"
      >
        No criteria selected
      </div>
    {/if}
  </div>
</div>
