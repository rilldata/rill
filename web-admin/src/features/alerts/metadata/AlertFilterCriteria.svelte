<script lang="ts">
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import MeasureFilterReadOnlyChip from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterReadOnlyChip.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { getMeasureFilterForDimension } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { flip } from "svelte/animate";
  import { fly } from "svelte/transition";

  export let metricsViewName: string;
  export let filters: V1Expression | undefined;

  $: filtersLength = filters?.cond?.exprs?.length ?? 0;

  $: dashboard = useDashboard($runtime.instanceId, metricsViewName);
  $: measures = $dashboard.data?.metricsView?.state?.validSpec?.measures ?? [];
  $: measureIdMap = getMapFromArray(measures, (measure) => measure.name);
  $: measureFilters = getMeasureFilterForDimension(measureIdMap, filters);
</script>

<div class="flex flex-col gap-y-3">
  <MetadataLabel>Criteria</MetadataLabel>
  <div class="flex flex-wrap gap-2">
    {#if filtersLength}
      {#each measureFilters as { name, label, dimensionName, expr } (name)}
        <div animate:flip={{ duration: 200 }}>
          <MeasureFilterReadOnlyChip {label} {dimensionName} {expr} />
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
