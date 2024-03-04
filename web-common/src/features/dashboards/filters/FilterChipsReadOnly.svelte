<!-- @component
The main feature-set component for dashboard filters
 -->
<script lang="ts">
  import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { flip } from "svelte/animate";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useDashboard } from "../selectors";
  import { getDimensionFilters } from "../state-managers/selectors/dimension-filters";
  import { getMeasureFilters } from "../state-managers/selectors/measure-filters";
  import DimensionFilterReadOnlyChip from "./dimension-filters/DimensionFilterReadOnlyChip.svelte";
  import MeasureFilterReadOnlyChip from "./measure-filters/MeasureFilterReadOnlyChip.svelte";

  export let metricsViewName: string;
  export let filters: V1Expression | undefined;

  $: dashboard = useDashboard($runtime.instanceId, metricsViewName);

  // Get dimension filters
  $: dimensions =
    $dashboard.data?.metricsView?.state?.validSpec?.dimensions ?? [];
  $: dimensionIdMap = getMapFromArray(
    dimensions,
    (dimension) => dimension.name as string,
  );
  $: dimensionFilters = getDimensionFilters(dimensionIdMap, filters);

  // Get measure filters
  $: measures = $dashboard.data?.metricsView?.state?.validSpec?.measures ?? [];
  $: measureIdMap = getMapFromArray(
    measures,
    (measure) => measure.name as string,
  );
  $: measureFilters = getMeasureFilters(measureIdMap, filters);
</script>

<div class="relative flex flex-row flex-wrap gap-x-2 gap-y-2 items-center">
  {#if dimensionFilters.length > 0}
    {#each dimensionFilters as { name, label, selectedValues } (name)}
      {@const dimension = dimensions.find((d) => d.name === name)}
      <div animate:flip={{ duration: 200 }}>
        {#if dimension?.column}
          <DimensionFilterReadOnlyChip
            label={label ?? name}
            values={selectedValues}
          />
        {/if}
      </div>
    {/each}
  {/if}
  {#if measureFilters.length > 0}
    {#each measureFilters as { name, label, dimensionName, expr } (name)}
      <div animate:flip={{ duration: 200 }}>
        <MeasureFilterReadOnlyChip
          label={label ?? name}
          {dimensionName}
          {expr}
        />
      </div>
    {/each}
  {/if}
</div>
