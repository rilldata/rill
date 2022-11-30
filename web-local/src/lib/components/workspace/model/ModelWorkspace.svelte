<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { getContext } from "svelte";
  import type { PersistentModelStore } from "../../../application-state-stores/model-stores";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import ModelInspector from "./inspector/ModelInspector.svelte";
  import ModelBody from "./ModelBody.svelte";

  export let modelName: string;

  const persistentModelStore = getContext(
    "rill:app:persistent-model-store"
  ) as PersistentModelStore;

  $: model = $persistentModelStore?.entities
    ? $persistentModelStore.entities.find(
        (model) => model.tableName === modelName
      )
    : undefined;

  const switchToModel = async (modelName: string) => {
    if (!modelName) return;

    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.Model,
      modelName,
    ]);
  };

  $: switchToModel(modelName);
</script>

{#key model?.id}
  <WorkspaceContainer assetID={modelName}>
    <div slot="body">
      <ModelBody {modelName} />
    </div>
    <ModelInspector {modelName} slot="inspector" />
  </WorkspaceContainer>
{/key}
