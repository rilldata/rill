<script lang="ts">
  import WorkspaceHeader from "../WorkspaceHeader.svelte";
  import { store, reduxReadable } from "$lib/redux-store/store-root";
  import {
    singleMetricsDefSelector,
    updateMetricsDefsApi,
  } from "$lib/redux-store/metrics-definition-slice";

  export let metricsDefId;

  $: selectedMetricsDef =
    singleMetricsDefSelector(metricsDefId)($reduxReadable);

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
