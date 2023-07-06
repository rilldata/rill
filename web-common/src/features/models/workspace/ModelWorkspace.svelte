<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import ModelInspector from "./inspector/ModelInspector.svelte";
  import ModelBody from "./ModelBody.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";

  export let modelName: string;
  export let focusEditorOnMount = false;
</script>

{#key modelName}
  <WorkspaceContainer assetID={modelName}>
    <div slot="header">
      <ModelWorkspaceHeader {modelName} />
    </div>
    <div slot="body">
      <ModelBody {modelName} {focusEditorOnMount} />
    </div>
    <ModelInspector {modelName} slot="inspector" />
  </WorkspaceContainer>
{/key}
