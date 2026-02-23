<script lang="ts">
  import { goto } from "$app/navigation";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import ThemeEditor from "@rilldata/web-common/features/themes/editor/ThemeEditor.svelte";
  import ThemeDashboardPreview from "@rilldata/web-common/features/themes/ThemeDashboardPreview.svelte";
  import VisualTheme from "@rilldata/web-common/features/themes/VisualTheme.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
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

  $: workspace = workspaces.get(filePath);
  $: selectedView = workspace.view;

  $: errors = mapParseErrorsToLines(allErrors, $remoteContent ?? "");
</script>

<WorkspaceContainer>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={ResourceKind.Theme}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    showInspectorToggle={false}
    slot="header"
    codeToggle
    titleInput={fileName}
  >
    <div slot="cta">
      <ThemeDashboardPreview />
    </div>
  </WorkspaceHeader>

  <div slot="body" class="size-full overflow-hidden flex flex-col">
    <div class="flex-1 min-h-0 overflow-hidden">
      {#if $selectedView === "code"}
        <ThemeEditor bind:autoSave={$autoSave} {fileArtifact} {errors} />
      {:else}
        <VisualTheme {filePath} />
      {/if}
    </div>
  </div>
</WorkspaceContainer>
