<script lang="ts">
  import { getContext } from "svelte";
  import ModelView from "./Model.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";
  import MetricsDefWorkspaceHeader from "./MetricsDefWorkspaceHeader.svelte";

  import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  const rillAppStore = getContext("rill:app:store") as ApplicationStore;

  $: useModelWorkspace = $rillAppStore?.activeEntity?.type === EntityType.Model;
  $: useMetricsDefWorkspace =
    $rillAppStore?.activeEntity?.type === EntityType.MetricsDef;
  $: activeEntityID = $rillAppStore?.activeEntity?.id;
</script>

{#if useModelWorkspace}
  <ModelWorkspaceHeader />
  <ModelView />
{:else if useMetricsDefWorkspace}
  <MetricsDefWorkspaceHeader metricsDefId={activeEntityID} />
  <!-- <ModelView /> -->
{/if}
