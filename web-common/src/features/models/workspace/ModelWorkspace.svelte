<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import ModelInspector from "./inspector/ModelInspector.svelte";
  import ModelBody from "./ModelBody.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";

  export let filePath: string;
  export let focusEditorOnMount = false;
</script>

{#key filePath}
  <WorkspaceContainer>
    <ModelWorkspaceHeader slot="header" {filePath} />
    <ModelBody slot="body" {filePath} {focusEditorOnMount} />
    <ModelInspector {filePath} slot="inspector" />
  </WorkspaceContainer>
{/key}
