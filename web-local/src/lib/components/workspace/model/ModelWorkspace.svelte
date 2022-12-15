<script lang="ts">
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";

  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import ModelInspector from "./inspector/ModelInspector.svelte";
  import ModelBody from "./ModelBody.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";

  export let modelName: string;

  const switchToModel = async (modelName: string) => {
    if (!modelName) return;

    appStore.setActiveEntity(modelName, EntityType.Model);
  };

  $: switchToModel(modelName);
</script>

{#key modelName}
  <WorkspaceContainer assetID={modelName}>
    <div slot="header">
      <ModelWorkspaceHeader {modelName} />
      <!-- <WorkspaceHeader
        {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
      >
        <svelte:fragment slot="right">
          <IconButton on:click={()}>B</IconButton>
        </svelte:fragment>
      </WorkspaceHeader> -->
    </div>
    <div slot="body">
      <ModelBody {modelName} />
    </div>
    <ModelInspector {modelName} slot="inspector" />
  </WorkspaceContainer>
{/key}
