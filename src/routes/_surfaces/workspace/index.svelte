<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";
  import { getContext } from "svelte";
  import Explore from "./explore/Explore.svelte";
  import MetricsDefWorkspace from "./metrics-def/MetricsDefWorkspace.svelte";
  import MetricsDefWorkspaceHeader from "./metrics-def/MetricsDefWorkspaceHeader.svelte";
  import ModelView from "./Model.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";
  import Onboarding from "./Onboarding.svelte";

  const rillAppStore = getContext("rill:app:store") as ApplicationStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  $: useModelWorkspace = $rillAppStore?.activeEntity?.type === EntityType.Model;
  $: useMetricsDefWorkspace =
    $rillAppStore?.activeEntity?.type === EntityType.MetricsDefinition;
  $: useExplore =
    $rillAppStore?.activeEntity?.type === EntityType.MetricsExplore;
  $: activeEntityID = $rillAppStore?.activeEntity?.id;

  $: isModelActive =
    $rillAppStore?.activeEntity && $persistentModelStore?.entities
      ? $persistentModelStore.entities.find(
          (q) => q.id === $rillAppStore.activeEntity.id
        )
      : undefined;
</script>

{#if useModelWorkspace}
  {#if !isModelActive}
    <Onboarding />
  {:else}
    <ModelWorkspaceHeader />
    <ModelView />
  {/if}
{:else if useMetricsDefWorkspace}
  <MetricsDefWorkspaceHeader metricsDefId={activeEntityID} />
  <MetricsDefWorkspace metricsDefId={activeEntityID} />
{:else if useExplore}
  <Explore metricsDefId={activeEntityID} />
{/if}
