<script lang="ts">
  import MetricsIcon from "../../icons/Metrics.svelte";
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

<WorkspaceHeader
  {...{ titleInput, onChangeCallback }}
  showStatus={false}
  showInspectorToggle={false}
>
  <MetricsIcon slot="icon" />
  <!-- {#if !metricsSourceSelectionError} -->
  <MetricsDefinitionExploreMetricsButton
    slot="cta"
    {metricsDefName}
    {metricsInternalRep}
  />
  <!-- {/if} -->
</WorkspaceHeader>
