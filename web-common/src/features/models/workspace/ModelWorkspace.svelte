<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import ModelInspector from "./inspector/ModelInspector.svelte";
  import ModelBody from "./ModelBody.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";

  export let modelName: string;
  export let focusEditorOnMount = false;
</script>

{#key modelName}
  <WorkspaceContainer>
    <ModelWorkspaceHeader slot="header" {modelName} />
    <ModelBody slot="body" {modelName} {focusEditorOnMount} />
    <ModelInspector {modelName} slot="inspector" />
  </WorkspaceContainer>
{/key}
