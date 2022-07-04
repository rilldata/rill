<script lang="ts">
  import { getContext } from "svelte";
  import ModelView from "./Model.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";

  import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import MetricsDefWorkspaceHeader from "./metrics-def/MetricsDefWorkspaceHeader.svelte";
  import MetricsDefWorkspace from "./metrics-def/MetricsDefWorkspace.svelte";
  import MetricsLeaderboard from "./leaderboard/MetricsLeaderboard.svelte";
  const rillAppStore = getContext("rill:app:store") as ApplicationStore;

  $: useModelWorkspace = $rillAppStore?.activeEntity?.type === EntityType.Model;
  $: useMetricsDefWorkspace =
    $rillAppStore?.activeEntity?.type === EntityType.MetricsDefinition;
  $: useMetricsLeaderboard =
    $rillAppStore?.activeEntity?.type === EntityType.MetricsLeaderboard;
  $: activeEntityID = $rillAppStore?.activeEntity?.id;
</script>

{#if useModelWorkspace}
  <ModelWorkspaceHeader />
  <ModelView />
{:else if useMetricsDefWorkspace}
  <MetricsDefWorkspaceHeader metricsDefId={activeEntityID} />
  <MetricsDefWorkspace metricsDefId={activeEntityID} />
{:else if useMetricsLeaderboard}
  <MetricsLeaderboard metricsDefId={activeEntityID} />
{/if}
