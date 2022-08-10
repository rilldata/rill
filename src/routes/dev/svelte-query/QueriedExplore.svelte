<script lang="ts">
  import { useQuery } from "@sveltestack/svelte-query";
  import { fetchWrapper } from "$lib/util/fetchWrapper";
  import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";

  const metricsDefinitions = useQuery<Array<MetricsDefinitionEntity>>(
    "metrics",
    () => fetchWrapper("metrics", "GET")
  );
  let metricsDefId: string;

  const measuresFetcher = async ({ queryKey }) => {
    const [, metricsId] = queryKey;
    return await fetchWrapper(`measures/?metricsDefId=${metricsId}`, "GET");
  };
  const measures = useQuery<Array<MeasureDefinitionEntity>>(
    ["measures", metricsDefId],
    measuresFetcher
  );
  $: measures.setOptions(["measures", metricsDefId], measuresFetcher);
  let measureId = "";

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
      measureId = "";
      dimensionId = "";
    }}
  >
    {#each $metricsDefinitions.data as metric (metric.id)}
      <option value={metric.id}>{metric.metricDefLabel}</option>
    {/each}
  </select>
</div>
<div>
  Measure: <select
    value={measureId}
    on:change={(evt) => (measureId = evt.target.value)}
  >
    {#each $measures.data as measure (measure.id)}
      <option value={measure.id}>{measure.expression}</option>
    {/each}
  </select>
</div>
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
