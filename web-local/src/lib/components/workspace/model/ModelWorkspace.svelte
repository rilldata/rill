<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import ModelInspector from "../inspector/model/ModelInspector.svelte";
  import ModelBody from "./ModelBody.svelte";
  export let modelID;

  const switchToModel = async (modelID) => {
    if (!modelID) return;

    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.Model,
      modelID,
    ]);
  };

  $: switchToModel(modelID);
</script>

{#key modelID}
  <WorkspaceContainer assetID={modelID}>
    <div slot="body">
      <ModelBody {modelID} />
    </div>
    <ModelInspector slot="inspector" />
  </WorkspaceContainer>
{/key}
