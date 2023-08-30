<script lang="ts">
  import ColumnProfileProvider from "@rilldata/web-common/components/column-profile/ColumnProfileProvider.svelte";
  import { WorkspaceContainer } from "../../../layout/workspace";
  import BeforeLeavingUnsavedSource from "../editor/BeforeLeavingUnsavedSource.svelte";
  import SourceInspector from "../inspector/SourceInspector.svelte";
  import SourceWorkspaceBody from "./SourceWorkspaceBody.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceName: string;
</script>

{#key sourceName}
  <BeforeLeavingUnsavedSource {sourceName}>
    <WorkspaceContainer assetID={sourceName}>
      <SourceWorkspaceHeader {sourceName} slot="header" />
      <SourceWorkspaceBody {sourceName} slot="body" />
      <ColumnProfileProvider objectName={sourceName} slot="inspector">
        <SourceInspector {sourceName} />
      </ColumnProfileProvider>
    </WorkspaceContainer>
  </BeforeLeavingUnsavedSource>
{/key}
