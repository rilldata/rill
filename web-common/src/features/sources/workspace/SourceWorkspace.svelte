<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import BeforeLeavingUnsavedSource from "../editor/BeforeLeavingUnsavedSource.svelte";
  import SourceInspector from "../inspector/SourceInspector.svelte";
  import SourceWorkspaceBody from "./SourceWorkspaceBody.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let filePath: string;
</script>

{#key filePath}
  <BeforeLeavingUnsavedSource {filePath}>
    <WorkspaceContainer>
      <SourceWorkspaceHeader {filePath} slot="header" />
      <SourceWorkspaceBody {filePath} slot="body" />
      <SourceInspector {filePath} slot="inspector" />
    </WorkspaceContainer>
  </BeforeLeavingUnsavedSource>
{/key}
