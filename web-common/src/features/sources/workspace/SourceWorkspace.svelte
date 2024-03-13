<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import BeforeLeavingUnsavedSource from "../editor/BeforeLeavingUnsavedSource.svelte";
  import SourceInspector from "../inspector/SourceInspector.svelte";
  import SourceWorkspaceBody from "./SourceWorkspaceBody.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceName: string;
</script>

{#key sourceName}
  <BeforeLeavingUnsavedSource {sourceName}>
    <WorkspaceContainer>
      <SourceWorkspaceHeader {sourceName} slot="header" />
      <SourceWorkspaceBody {sourceName} slot="body" />
      <SourceInspector {sourceName} slot="inspector" />
    </WorkspaceContainer>
  </BeforeLeavingUnsavedSource>
{/key}
