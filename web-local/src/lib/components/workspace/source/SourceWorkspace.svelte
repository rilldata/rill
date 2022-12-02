<script lang="ts">
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { EntityType } from "@rilldata/web-local/lib/temp/entity";
  import { ConnectedPreviewTable } from "../../preview-table";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import SourceInspector from "./SourceInspector.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceName: string;

  const switchToSource = async (name: string) => {
    if (!name) return;

    appStore.setActiveEntity(name, EntityType.Table);
  };

  $: switchToSource(sourceName);
</script>

{#key sourceName}
  <WorkspaceContainer assetID={sourceName}>
    <div
      slot="body"
      class="grid pb-6"
      style:grid-template-rows="max-content auto"
      style:height="100vh"
    >
      <SourceWorkspaceHeader {sourceName} />
      <div
        style:overflow="auto"
        style:height="100%"
        class="m-6 mt-0 border border-gray-300 rounded"
      >
        {#key sourceName}
          <ConnectedPreviewTable objectName={sourceName} />
        {/key}
      </div>
    </div>

    <SourceInspector {sourceName} slot="inspector" />
  </WorkspaceContainer>
{/key}
