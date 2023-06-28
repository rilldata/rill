<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import SourceInspector from "./SourceInspector.svelte";
  import SourceWorkspaceBody from "./SourceWorkspaceBody.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceName: string;
  export let embedded = false;
  export let path: string = undefined;
</script>

{#key sourceName}
  <WorkspaceContainer assetID={sourceName}>
    <SourceWorkspaceHeader {sourceName} {path} {embedded} slot="header" />
    <SourceWorkspaceBody {sourceName} slot="body" />
    <SourceInspector {sourceName} slot="inspector" />
  </WorkspaceContainer>
{/key}
