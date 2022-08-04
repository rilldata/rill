<script lang="ts">
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import MetricsDefinitionExploreMetricsButton from "$lib/components/metrics-definition/MetricsDefinitionExploreMetricsButton.svelte";
  import { updateMetricsDefsWrapperApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import WorkspaceHeader from "../WorkspaceHeader.svelte";
  import { MetricsSourceSelectionError } from "$common/errors/ErrorMessages";

  export let metricsDefId;

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  $: titleInput = $selectedMetricsDef?.metricDefLabel;

  const onChangeCallback = async (e) => {
    store.dispatch(
      updateMetricsDefsWrapperApi({
        id: metricsDefId,
        changes: { metricDefLabel: e.target.value },
      })
    );
  };

  $: metricsSourceSelectionError = $selectedMetricsDef
    ? MetricsSourceSelectionError($selectedMetricsDef)
    : "";
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

  {#if !metricsSourceSelectionError}
    <MetricsDefinitionExploreMetricsButton {metricsDefId} />
  {/if}
</div>
