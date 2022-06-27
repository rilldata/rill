<script lang="ts">
  import WorkspaceHeader from "../WorkspaceHeader.svelte";
  import { store, reduxReadable } from "$lib/redux-store/store-root";
  import { updateMetricsDefsApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { selectMetricsDefinitionById } from "$lib/redux-store/metrics-definition/metrics-definitioin-selectors";

  export let metricsDefId;

  $: selectedMetricsDef =
    selectMetricsDefinitionById(metricsDefId)($reduxReadable);

  $: titleInput = selectedMetricsDef?.metricDefLabel;

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
