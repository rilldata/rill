<script lang="ts">
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { Inspector } from "@rilldata/web-local/lib/components/workspace";
  import WorkspaceBody from "@rilldata/web-local/lib/components/workspace/core/WorkspaceBody.svelte";
  import WorkspaceContainer from "@rilldata/web-local/lib/components/workspace/core/WorkspaceContainer.svelte";
  import ModelInspector from "./inspector/ModelInspector.svelte";
  import ModelBody from "./ModelBody.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";

  export let modelName: string;
  export let focusEditorOnMount = false;

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
    </div>

    <div>
      <WorkspaceBody top="var(--header-height)">
        <ModelBody {modelName} {focusEditorOnMount} />
      </WorkspaceBody>
      <Inspector>
        <ModelInspector {modelName} />
      </Inspector>
    </div>

    <!-- <div slot="body">
      <ModelBody {modelName} {focusEditorOnMount} />
    </div> -->
  </WorkspaceContainer>
{/key}
