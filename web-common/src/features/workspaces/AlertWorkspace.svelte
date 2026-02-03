<script lang="ts">
  import { goto } from "$app/navigation";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { mapParseErrorsToLines } from "../metrics-views/errors";
  import AlertEditor from "../alerts/AlertEditor.svelte";
  import AlertVisualWorkspaceEditor from "../alerts/AlertVisualWorkspaceEditor.svelte";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    fileName,
    getAllErrors,
    remoteContent,
    getResource,
  } = fileArtifact);

  $: alertName = $resourceName?.name ?? getNameFromFile(filePath);

  $: resourceQuery = getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data;

  $: allErrorsQuery = getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;

  $: selectedView = $selectedViewStore ?? "code";

  $: lineBasedRuntimeErrors = mapParseErrorsToLines(
    allErrors,
    $remoteContent ?? "",
  );

  $: mainError = lineBasedRuntimeErrors?.at(0);

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      instanceId,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }
</script>

<WorkspaceContainer>
  <WorkspaceHeader
    {resource}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    slot="header"
    titleInput={fileName}
    {filePath}
    codeToggle
    resourceKind={ResourceKind.Alert}
  >
    <div class="flex gap-x-2" slot="cta">
      <!-- Alerts don't have a preview button like explores -->
    </div>
  </WorkspaceHeader>

  <WorkspaceEditorContainer
    slot="body"
    error={mainError}
    showError={!!$remoteContent && selectedView === "code"}
  >
    {#if selectedView === "code"}
      <AlertEditor bind:autoSave={$autoSave} {fileArtifact} />
    {:else if selectedView === "viz"}
      <AlertVisualWorkspaceEditor
        {fileArtifact}
        {alertName}
        errors={lineBasedRuntimeErrors ?? []}
      />
    {/if}
  </WorkspaceEditorContainer>
</WorkspaceContainer>
