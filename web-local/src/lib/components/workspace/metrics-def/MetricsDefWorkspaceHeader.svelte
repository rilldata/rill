<script lang="ts">
  import { MetricsSourceSelectionError } from "$web-local/common/errors/ErrorMessages";
  import MetricsIcon from "../../icons/Metrics.svelte";
  import MetricsDefinitionExploreMetricsButton from "../../metrics-definition/MetricsDefinitionExploreMetricsButton.svelte";
  import { updateMetricsDefsWrapperApi } from "../../../redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "../../../redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "../../../redux-store/store-root";
  import WorkspaceHeader from "../WorkspaceHeader.svelte";

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
