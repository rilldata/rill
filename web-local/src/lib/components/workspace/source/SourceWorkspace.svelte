<script lang="ts">
  import { EntityType } from "../../../../common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "../../../application-state-stores/application-store";
  import { ConnectedPreviewTable } from "../../preview-table";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
  import SourceInspector from "./SourceInspector.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceName: string;

  const switchToSource = async (name: string) => {
    if (!name) return;

    await dataModelerService.dispatch("setActiveAsset", [
      EntityType.Table,
      name,
    ]);
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
