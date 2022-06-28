<script lang="ts">
  import TimestampIcon from "$lib/components/icons/TimestampType.svelte";
  import { store } from "$lib/redux-store/store-root";
  import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { getContext } from "svelte";
  import type { DerivedModelStore } from "$lib/application-state-stores/model-stores";
  import type { ProfileColumn } from "$lib/types";
  import { fetchManyDimensionsApi } from "$lib/redux-store/dimension-definition/dimension-definition-apis";
  import { fetchManyMeasuresApi } from "$lib/redux-store/measure-definition/measure-definition-apis";
  import { updateMetricsDefsApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";

  export let metricsDefId;

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  $: timeColumnSelectedValue =
    $selectedMetricsDef?.timeDimension || "__DEFAULT_VALUE__";

  $: if (metricsDefId) {
    store.dispatch(fetchManyMeasuresApi({ metricsDefId }));
    store.dispatch(fetchManyDimensionsApi({ metricsDefId }));
  }
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let derivedModelColumns: Array<ProfileColumn>;
  $: if ($selectedMetricsDef?.sourceModelId && $derivedModelStore?.entities) {
    derivedModelColumns = $derivedModelStore?.entities.find(
      (model) => model.id === $selectedMetricsDef.sourceModelId
    ).profile;
  } else {
    derivedModelColumns = [];
  }

  function updateMetricsDefinitionHandler(
    metricsDef: Partial<MetricsDefinitionEntity>
  ) {
    store.dispatch(
      updateMetricsDefsApi({
        id: metricsDefId,
        changes: metricsDef,
      })
    );
  }
</script>

<div style:height="40px" class="flex items-center pl-1 mb-2">
  <div class="flex items-center gap-x-2 pr-5">
    <TimestampIcon size="16px" /> timestamp
  </div>
  <div>
    {#if $selectedMetricsDef?.sourceModelId === undefined}
      <em>select a model before selecting a timestamp</em>
    {:else if derivedModelColumns.length === 0}
      <em>the selected model has no timestamp columns</em>
    {:else}
      <select
        class="italic hover:bg-gray-100 rounded border border-6 border-transparent hover:font-bold hover:border-gray-100"
        style="background-color: #FFF;"
        value={timeColumnSelectedValue}
        on:change={(evt) => {
          updateMetricsDefinitionHandler({ timeDimension: evt.target.value });
        }}
      >
        <option value="__DEFAULT_VALUE__" disabled selected hidden
          >select a timestamp...</option
        >
        {#each derivedModelColumns as column}
          <option value={column.name}>{column.name}</option>
        {/each}
      </select>
    {/if}
  </div>
</div>
