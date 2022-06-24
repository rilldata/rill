<script lang="ts">
  import WorkspaceHeader from "../WorkspaceHeader.svelte";
  import {
    createReadableStoreWithSelector,
    store,
  } from "$lib/redux-store/store-root";
  import { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import type { Readable } from "svelte/store";
  import {
    singleMetricsDefSelector,
    updateMetricsDefsApi,
  } from "$lib/redux-store/metrics-definition-slice";

  export let metricsDefId;

  let selectedMetricsDef: Readable<MetricsDefinitionEntity>;
  $: if (metricsDefId) {
    selectedMetricsDef = createReadableStoreWithSelector(
      singleMetricsDefSelector(metricsDefId)
    );
  }

  let titleInput;
  $: titleInput = $selectedMetricsDef?.metricDefLabel;
  const onChangeCallback = async (e) => {
    store.dispatch(
      updateMetricsDefsApi({
        id: metricsDefId,
        changes: { metricDefLabel: e.target.value },
      })
    );
  };
</script>

<WorkspaceHeader {...{ titleInput, onChangeCallback }} />
