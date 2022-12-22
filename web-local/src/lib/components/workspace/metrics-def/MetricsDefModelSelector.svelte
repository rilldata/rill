<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { useModelNames } from "@rilldata/web-local/lib/svelte-query/models";
  import type { Readable } from "svelte/store";
  import type { MetricsInternalRepresentation } from "../../../application-state-stores/metrics-internal-store";

  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;

  $: sourceModelDisplayValue =
    $metricsInternalRep.getMetricKey("model") || "__DEFAULT_VALUE__";

  $: allModels = useModelNames($runtimeStore.instanceId);
  function updateMetricsDefinitionHandler(sourceModelName) {
    $metricsInternalRep.updateMetricKey("model", sourceModelName);
  }
</script>

<div class="flex items-center mb-3">
  <div class="text-gray-500 font-medium" style="width:10em; font-size:11px;">
    Model
  </div>
  <div>
    <select
      class="hover:bg-gray-100 rounded border border-6 border-transparent hover:border-gray-300"
      on:change={(evt) => {
        updateMetricsDefinitionHandler(evt.target?.value);
      }}
      style="background-color: #FFF; width:18em"
      value={sourceModelDisplayValue}
    >
      <option disabled selected value="__DEFAULT_VALUE__"
        >Select a model...</option
      >
      {#each $allModels.data || [] as modelName}
        <option value={modelName}>{modelName}</option>
      {/each}
    </select>
  </div>
</div>
