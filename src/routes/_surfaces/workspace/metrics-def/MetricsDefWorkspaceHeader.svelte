<script lang="ts">
  import WorkspaceHeader from "../WorkspaceHeader.svelte";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import { store } from "$lib/redux-store/store-root";
  import { updateMetricsDefsApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";

  export let metricsDefId;

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

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

<WorkspaceHeader {...{ titleInput, onChangeCallback }}>
  <svelte:fragment slot="icon">
    <MetricsIcon />
  </svelte:fragment>
</WorkspaceHeader>
