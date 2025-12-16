<script lang="ts">
  import { goto } from "$app/navigation";
  import CanvasEditor from "@rilldata/web-common/features/canvas/CanvasEditor.svelte";
  import VisualCanvasEditing from "@rilldata/web-common/features/canvas/inspector/VisualCanvasEditing.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    resourceIsLoading,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import PreviewButton from "../explores/PreviewButton.svelte";
  import CanvasBuilder from "../canvas/CanvasBuilder.svelte";
  import SaveDefaultsButton from "../canvas/components/SaveDefaultsButton.svelte";
  import CanvasProvider from "../canvas/CanvasProvider.svelte";

  export let fileArtifact: FileArtifact;

  let canvasName: string;
  let selectedView: "split" | "code" | "viz";
  let ready = false;

  $: ({ instanceId } = $runtime);

  $: ({
    autoSave,
    path: filePath,
    fileName,
    getResource,
    getAllErrors,
    remoteContent,
    hasUnsavedChanges,
    saveState: { saving },
  } = fileArtifact);

  // Reset ready when canvasName changes
  $: if (canvasName) ready = false;

  $: resourceQuery = getResource(queryClient, instanceId);

  $: ({ data } = $resourceQuery);

  $: allErrorsQuery = getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;

  $: resourceIsReconciling = resourceIsLoading(data);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;
  $: selectedView = $selectedViewStore ?? "code";

  $: canvasName = getNameFromFile(filePath);

  $: lineBasedRuntimeErrors = mapParseErrorsToLines(
    allErrors,
    $remoteContent ?? "",
  );

  $: mainError = lineBasedRuntimeErrors?.at(0);

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      $runtime.instanceId,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }
</script>

{#key canvasName}
  <WorkspaceContainer>
    <WorkspaceHeader
      slot="header"
      {filePath}
      resource={data}
      hasUnsavedChanges={$hasUnsavedChanges}
      titleInput={fileName}
      codeToggle
      onTitleChange={onChangeCallback}
      resourceKind={ResourceKind.Canvas}
    >
      <div class="flex gap-x-2" slot="cta">
        {#if ready}
          <SaveDefaultsButton {canvasName} {instanceId} saving={$saving} />
        {/if}

        <PreviewButton
          href="/canvas/{canvasName}"
          disabled={allErrors.length > 0 || resourceIsReconciling}
          reconciling={resourceIsReconciling}
        />
      </div>
    </WorkspaceHeader>

    <WorkspaceEditorContainer
      slot="body"
      error={mainError}
      showError={!!$remoteContent && selectedView === "code"}
    >
      {#if selectedView === "code"}
        <CanvasEditor
          bind:autoSave={$autoSave}
          {canvasName}
          {fileArtifact}
          {lineBasedRuntimeErrors}
        />
      {:else if selectedView === "viz"}
        <CanvasProvider {canvasName} {instanceId} bind:ready>
          <CanvasBuilder
            {canvasName}
            openSidebar={workspace.inspector.open}
            {fileArtifact}
          />
        </CanvasProvider>
      {/if}
    </WorkspaceEditorContainer>
    <svelte:fragment slot="inspector">
      {#if ready}
        <VisualCanvasEditing
          {canvasName}
          {fileArtifact}
          autoSave={selectedView === "viz" || $autoSave}
        />
      {/if}
    </svelte:fragment>
  </WorkspaceContainer>
{/key}
