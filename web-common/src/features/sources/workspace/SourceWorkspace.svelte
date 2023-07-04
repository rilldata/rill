<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import SourceInspector from "./SourceInspector.svelte";
  import SourceWorkspaceBody from "./SourceWorkspaceBody.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceName: string;
</script>

{#key sourceName}
  <WorkspaceContainer assetID={sourceName}>
    <SourceWorkspaceHeader {sourceName} slot="header" />
    <SourceWorkspaceBody {sourceName} slot="body" />
    <SourceInspector {sourceName} slot="inspector" />
  </WorkspaceContainer>
{/key}
