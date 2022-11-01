<script lang="ts">
  import { MetricsSourceSelectionError } from "@rilldata/web-local/common/errors/ErrorMessages";
  import { updateMetricsDefsWrapperApi } from "../../../redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "../../../redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "../../../redux-store/store-root";
  import MetricsIcon from "../../icons/Metrics.svelte";
  import MetricsDefinitionExploreMetricsButton from "../../metrics-definition/MetricsDefinitionExploreMetricsButton.svelte";
  import WorkspaceHeader from "../core/WorkspaceHeader.svelte";

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
    <MetricsIcon slot="icon" />
  </WorkspaceHeader>

  {#if !metricsSourceSelectionError}
    <MetricsDefinitionExploreMetricsButton {metricsDefId} />
  {/if}
</div>
