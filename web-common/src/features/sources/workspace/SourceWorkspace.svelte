<script lang="ts">
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { WorkspaceContainer } from "@rilldata/web-local/lib/components/workspace/index.js";
  import SourceInspector from "./SourceInspector.svelte";
  import SourceWorkspaceErrorStates from "./SourceWorkspaceErrorStates.svelte";
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
    <SourceWorkspaceHeader {sourceName} slot="header" />
    <SourceWorkspaceErrorStates {sourceName} slot="body" />
    <SourceInspector {sourceName} slot="inspector" />
  </WorkspaceContainer>
{/key}
