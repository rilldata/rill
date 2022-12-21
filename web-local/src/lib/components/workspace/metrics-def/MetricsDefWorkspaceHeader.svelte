<script lang="ts">
  import MetricsIcon from "@rilldata/web-common/components/icons/Metrics.svelte";
  import MetricsDefinitionExploreMetricsButton from "../../metrics-definition/MetricsDefinitionExploreMetricsButton.svelte";
  import WorkspaceHeader from "../core/WorkspaceHeader.svelte";

  export let metricsDefName;
  export let metricsInternalRep;

  $: titleInput = $metricsInternalRep.getMetricKey("display_name");

  const onChangeCallback = async (e) => {
    $metricsInternalRep.updateMetricKey("display_name", e.target.value);
  };

  $: metricsSourceSelectionError = false;
</script>

<div
  class="grid gap-x-3 items-center pr-4"
  style:grid-template-columns="auto max-content"
>
  <WorkspaceHeader {...{ titleInput, onChangeCallback }} showStatus={false}>
    <MetricsIcon slot="icon" />
  </WorkspaceHeader>

  {#if !metricsSourceSelectionError}
    <MetricsDefinitionExploreMetricsButton
      {metricsDefName}
      {metricsInternalRep}
    />
  {/if}
</div>
