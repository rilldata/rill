<script lang="ts">
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import MetricsDefinitionExploreMetricsButton from "$lib/components/metrics-definition/MetricsDefinitionExploreMetricsButton.svelte";
  import MetricsDefinitionGoToModelButton from "$lib/components/metrics-definition/MetricsDefinitionGoToModelButton.svelte";
  import { updateMetricsDefsApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import WorkspaceHeader from "../WorkspaceHeader.svelte";

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

<div
  class="grid gap-x-3 items-center pr-4"
  style:grid-template-columns="auto max-content"
>
  <WorkspaceHeader {...{ titleInput, onChangeCallback }} showStatus={false}>
    <svelte:fragment slot="icon">
      <MetricsIcon />
    </svelte:fragment>
  </WorkspaceHeader>

  <div class="grid grid-flow-col gap-x-2">
    <MetricsDefinitionGoToModelButton {metricsDefId} />
    <MetricsDefinitionExploreMetricsButton {metricsDefId} />
  </div>
</div>
