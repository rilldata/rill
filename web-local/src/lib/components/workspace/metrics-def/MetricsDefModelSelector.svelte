<script lang="ts">
  import type { MetricsDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import type { Readable } from "svelte/store";
  import type { PersistentModelStore } from "../../../application-state-stores/model-stores";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import ModelIcon from "../../icons/Model.svelte";
  import { getContext } from "svelte";
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import type { MetricsInternalRepresentation } from "../../../application-state-stores/metrics-internal-store";

  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;

  $: sourceModelDisplayValue =
    $metricsInternalRep.getMetricKey("from") || "__DEFAULT_VALUE__";

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  $: allModels = useModelNames($runtimeStore.repoId);
  function updateMetricsDefinitionHandler(sourceModelName) {
    $metricsInternalRep.updateMetricKey("from", sourceModelName);
  }
</script>

<div class="flex items-center mb-3">
  <div class="flex items-center gap-x-2" style="width:9em">
    <ModelIcon size="16px" /> model
  </div>
  <div>
    <select
      class="italic hover:bg-gray-100 rounded border border-6 border-transparent hover:font-bold hover:border-gray-100"
      style="background-color: #FFF; width:18em"
      value={sourceModelDisplayValue}
      on:change={(evt) => {
        updateMetricsDefinitionHandler(evt.target?.value);
      }}
    >
      <option value="__DEFAULT_VALUE__" disabled selected
        >select a model...</option
      >
      {#each $allModels.data || [] as modelName}
        <option value={modelName}>{modelName}</option>
      {/each}
    </select>
  </div>
</div>
