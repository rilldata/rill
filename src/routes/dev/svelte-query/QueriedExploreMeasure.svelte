<script lang="ts">
  import { fetchWrapper } from "$lib/util/fetchWrapper";
  import { useQuery } from "@sveltestack/svelte-query";
  import { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

  export let metricsDefId: string;

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
</script>

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
