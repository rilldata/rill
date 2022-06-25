<script lang="ts">
  import ModelIcon from "$lib/components/icons/Code.svelte";

  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { getContext } from "svelte";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import {
    singleMetricsDefSelector,
    updateMetricsDefsApi,
  } from "$lib/redux-store/metrics-definition-slice";

  export let metricsDefId;
  $: selectedMetricsDef =
    singleMetricsDefSelector(metricsDefId)($reduxReadable);

  $: sourceModelDisplayValue =
    selectedMetricsDef?.sourceModelId || "__DEFAULT_VALUE__";

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

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
    <ModelIcon size="16px" /> model
  </div>
  <div>
    <select
      class="italic hover:bg-gray-100 rounded border border-6 border-transparent hover:font-bold hover:border-gray-100"
      style="background-color: #FFF;"
      value={sourceModelDisplayValue}
      on:change={(evt) => {
        updateMetricsDefinitionHandler({ sourceModelId: evt.target.value });
      }}
    >
      <option value="__DEFAULT_VALUE__" disabled selected
        >select a model...</option
      >
      {#each $persistentModelStore?.entities as entity}
        <option value={entity.id}>{entity.name}</option>
      {/each}
    </select>
  </div>
</div>
