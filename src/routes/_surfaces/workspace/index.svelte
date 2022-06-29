<script lang="ts">
  import { getContext } from "svelte";
  import ModelView from "./Model.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";

  import type { ApplicationStore } from "$lib/application-state-stores/application-store";
  import type { PersistentModelStore } from "$lib/application-state-stores/model-stores";

  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import Onboarding from "./Onboarding.svelte";
  const rillAppStore = getContext("rill:app:store") as ApplicationStore;
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  $: useModelWorkspace = $rillAppStore?.activeEntity?.type === EntityType.Model;
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
{:else}
  <!-- FIXME: this placeholder is here to show where you would plug in another kind of workspace component -->
  <ModelWorkspaceHeader />
  <ModelView />
{/if}
