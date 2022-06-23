<script lang="ts">
  import WorkspaceHeader from "../WorkspaceHeader.svelte";
  import { reduxReadable, store } from "$lib/redux-store/store-root";
  import { metricsDefinitionsApi } from "$lib/redux-store/metricsDefinitionsApi";
  import { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

  export let metricsDefId;

  const {
    endpoints: { getOneMetricsDefinition, updateMetricsDefinition },
  } = metricsDefinitionsApi;
  let selectedMetricsDef: MetricsDefinitionEntity;
  $: ({ data: selectedMetricsDef } =
    getOneMetricsDefinition.select(metricsDefId)($reduxReadable));
  $: if (metricsDefId) {
    store.dispatch(getOneMetricsDefinition.initiate(metricsDefId));
  }

  let titleInput;
  $: titleInput = selectedMetricsDef?.metricDefLabel;
  const onChangeCallback = async (e) => {
    store.dispatch(
      updateMetricsDefinition.initiate({
        id: metricsDefId,
        metricsDef: { metricDefLabel: e.target.value },
      })
    );
  };
</script>

<WorkspaceHeader {...{ titleInput, onChangeCallback }} />
