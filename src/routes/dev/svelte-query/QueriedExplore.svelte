<script lang="ts">
  import { useQuery } from "@sveltestack/svelte-query";
  import { fetchWrapper } from "$lib/util/fetchWrapper";
  import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import QueriedExploreMeasure from "./QueriedExploreMeasure.svelte";

  const metricsDefinitions = useQuery<Array<MetricsDefinitionEntity>>(
    "metrics",
    () => fetchWrapper("metrics", "GET")
  );
  let metricsDefId: string;

  const dimensionsFetcher = ({ queryKey }) => {
    const [, metricsId] = queryKey;
    return fetchWrapper(`dimensions/?metricsDefId=${metricsId}`, "GET");
  };
  const dimensions = useQuery<Array<DimensionDefinitionEntity>>(
    ["dimensions", metricsDefId],
    dimensionsFetcher
  );
  $: dimensions.setOptions(["dimensions", metricsDefId], dimensionsFetcher);
  let dimensionId = "";
</script>

<div>
  <select
    on:change={(evt) => {
      metricsDefId = evt.target.value;
      dimensionId = "";
    }}
  >
    {#each $metricsDefinitions.data as metric (metric.id)}
      <option value={metric.id}>{metric.metricDefLabel}</option>
    {/each}
  </select>
</div>
<QueriedExploreMeasure {metricsDefId} />
<div>
  Dimension: <select
    value={dimensionId}
    on:change={(evt) => (dimensionId = evt.target.value)}
  >
    {#each $dimensions.data as dimension (dimension.id)}
      <option value={dimension.id}>{dimension.dimensionColumn}</option>
    {/each}
  </select>
</div>
