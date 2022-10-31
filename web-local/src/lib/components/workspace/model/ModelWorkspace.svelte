<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type {
    DerivedModelStore,
    PersistentModelStore,
  } from "@rilldata/web-local/lib/application-state-stores/model-stores";
  import { getContext } from "svelte";
  import ModelInspector from "../inspector/model/ModelInspector.svelte";
  import WorkspaceContainer from "../WorkspaceContainer.svelte";
  import ModelBody from "./ModelBody.svelte";
  export let modelID;

  const queryHighlight = getContext("rill:app:query-highlight");
  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;
  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  $: currentModel = $persistentModelStore?.entities
    ? $persistentModelStore.entities.find((q) => q.id === modelID)
    : undefined;

  $: currentDerivedModel = $derivedModelStore?.entities
    ? $derivedModelStore.entities.find((q) => q.id === modelID)
    : undefined;

  const switchToModel = async (modelID) => {
    if (!modelID) return;

    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.Model,
      modelID,
    ]);
  };

  $: switchToModel(modelID);
</script>

{#if currentModel?.id}
  {#key currentModel?.id}
    <WorkspaceContainer assetID={modelID}>
      <div slot="body">
        <ModelBody {modelID} />
      </div>
      <ModelInspector slot="inspector" />
    </WorkspaceContainer>
  {/key}
{/if}

<style>
  .editor-pane {
    height: calc(100vh - var(--header-height));
  }
</style>
