<script lang="ts">
  import { goto } from "$app/navigation";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import DeployDashboardCta from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardCTA.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import ExploreEditor from "@rilldata/web-common/features/explores/ExploreEditor.svelte";
  import PreviewButton from "@rilldata/web-common/features/explores/PreviewButton.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import WorkspaceEditorContainer from "../../layout/workspace/WorkspaceEditorContainer.svelte";

  export let fileArtifact: FileArtifact;

  let previewStatus: string[] = [];

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    remoteContent,
    fileName,
  } = fileArtifact);

  $: exploreName = getNameFromFile(filePath);

  $: initLocalUserPreferenceStore(exploreName);

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: ({ data: resourceData, isFetching } = $resourceQuery);
  $: isResourceLoading = resourceIsLoading(resourceData);

  $: previewDisabled =
    !$remoteContent?.length ||
    !!allErrors?.length ||
    isResourceLoading ||
    isFetching;

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
    <div class="flex gap-x-2" slot="cta">
      <PreviewButton
        dashboardName={exploreName}
        disabled={previewDisabled}
        status={previewStatus}
      />
      <DeployDashboardCta />
    </div>
  </WorkspaceHeader>

  <WorkspaceEditorContainer slot="body">
    <ExploreEditor
      bind:autoSave={$autoSave}
      {exploreName}
      {fileArtifact}
      {allErrors}
    />
  </WorkspaceEditorContainer>
</WorkspaceContainer>
