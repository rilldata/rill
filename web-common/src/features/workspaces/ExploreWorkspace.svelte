<script lang="ts">
  import { goto } from "$app/navigation";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    resourceIsLoading,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import ExploreEditor from "@rilldata/web-common/features/explores/ExploreEditor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import PreviewButton from "../explores/PreviewButton.svelte";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    fileName,
    getResource,
    getAllErrors,
  } = fileArtifact);

  $: exploreName = $resourceName?.name ?? getNameFromFile(filePath);

  $: initLocalUserPreferenceStore(exploreName);

  $: resourceQuery = getResource(queryClient, instanceId);

  $: allErrorsQuery = getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceIsReconciling = resourceIsLoading($resourceQuery.data);

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      instanceId,
      newTitle,
      filePath,
      fileName,
      fileArtifacts.getNamesForKind(ResourceKind.Explore),
    );
    if (newRoute) await goto(newRoute);
  }
</script>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    showInspectorToggle={false}
    slot="header"
    titleInput={fileName}
    {filePath}
    resourceKind={ResourceKind.Explore}
  >
    <PreviewButton
      slot="cta"
      href="/explore/{exploreName}"
      disabled={allErrors.length > 0 || resourceIsReconciling}
    />
  </WorkspaceHeader>

  <ExploreEditor
    slot="body"
    bind:autoSave={$autoSave}
    {exploreName}
    {fileArtifact}
    {allErrors}
  />
</WorkspaceContainer>
