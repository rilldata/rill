<script lang="ts">
  import { useModelNames } from "@rilldata/web-common/features/models/selectors";
  import type { Readable } from "svelte/store";
  import { runtime } from "../../../runtime-client/runtime-store";
  import type { MetricsInternalRepresentation } from "../metrics-internal-store";

  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;

  $: sourceModelDisplayValue =
    $metricsInternalRep.getMetricKey("model") || "__DEFAULT_VALUE__";

  $: allModels = useModelNames($runtime.instanceId);
  function updateMetricsDefinitionHandler(sourceModelName) {
    // Reset timeseries as some models might not have a timeseries
    $metricsInternalRep.updateMetricKeys({
      model: sourceModelName,
      timeseries: "",
    });
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
