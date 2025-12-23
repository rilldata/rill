<script lang="ts">
  import { goto } from "$app/navigation";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import ThemeInspector from "@rilldata/web-common/features/themes/ThemeInspector.svelte";
  import ColorSchemePreviewPanel from "@rilldata/web-common/features/themes/ColorSchemePreviewPanel.svelte";
  import ThemeEditor from "@rilldata/web-common/features/themes/editor/ThemeEditor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { mapParseErrorsToLines } from "../metrics-views/errors";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    remoteContent,
    fileName,
  } = fileArtifact);

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

  $: errors = mapParseErrorsToLines(allErrors, $remoteContent ?? "");
</script>

<WorkspaceContainer inspector={true}>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={ResourceKind.Theme}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    showInspectorToggle={true}
    slot="header"
    titleInput={fileName}
  />

  <svelte:fragment slot="body">
    <div class="flex flex-col size-full overflow-hidden">
      <div class="flex-1 overflow-hidden">
        <ThemeEditor bind:autoSave={$autoSave} {fileArtifact} {errors} />
      </div>
      <ColorSchemePreviewPanel {filePath} />
    </div>
  </svelte:fragment>

  <ThemeInspector {filePath} slot="inspector" />
</WorkspaceContainer>
