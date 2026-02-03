<script lang="ts">
  import { goto } from "$app/navigation";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import APIEditor from "@rilldata/web-common/features/apis/editor/APIEditor.svelte";
  import VisualAPIEditor from "@rilldata/web-common/features/apis/editor/VisualAPIEditor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    remoteContent,
    fileName,
  } = fileArtifact);

  $: workspace = workspaces.get(filePath);

  $: apiName = $resourceName?.name ?? getNameFromFile(filePath);

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: ({ data: resource } = $resourceQuery);

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      instanceId,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }

  $: selectedView = workspace.view;

  $: errors = mapParseErrorsToLines(allErrors, $remoteContent ?? "");
</script>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={ResourceKind.API}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    slot="header"
    codeToggle
    showInspectorToggle={false}
    titleInput={fileName}
  />

  <svelte:fragment slot="body">
    {#if $selectedView === "code"}
      <APIEditor bind:autoSave={$autoSave} {fileArtifact} {errors} {apiName} />
    {:else}
      {#key fileArtifact}
        <VisualAPIEditor
          {errors}
          {fileArtifact}
          {apiName}
          switchView={() => {
            $selectedView = "code";
          }}
        />
      {/key}
    {/if}
  </svelte:fragment>
</WorkspaceContainer>
