<script lang="ts">
  import { goto } from "$app/navigation";
  import CanvasEditor from "@rilldata/web-common/features/canvas/CanvasEditor.svelte";
  import VisualCanvasEditing from "@rilldata/web-common/features/canvas/inspector/VisualCanvasEditing.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    resourceIsLoading,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
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
  import CanvasLoadingState from "../canvas/CanvasLoadingState.svelte";
  import CanvasInitialization from "../canvas/CanvasInitialization.svelte";

  export let fileArtifact: FileArtifact;

  let canvasName: string;
  let selectedView: "split" | "code" | "viz";

  $: ({ instanceId } = $runtime);

  $: ({
    autoSave,
    path: filePath,
    fileName,
    getResource,
    remoteContent,
    hasUnsavedChanges,
    saveState: { saving },
  } = fileArtifact);

  $: resourceQuery = getResource(queryClient, instanceId);

  $: ({ data } = $resourceQuery);

  $: resourceIsReconciling = resourceIsLoading(data);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;
  $: selectedView = $selectedViewStore ?? "code";

  $: canvasName = getNameFromFile(filePath);

  // Parse error for the editor gutter and banner
  $: parseErrorQuery = fileArtifact.getParseError(queryClient, instanceId);
  $: parseError = $parseErrorQuery;

  // Reconcile error resolved to root cause for the banner
  $: reconcileError = data?.meta?.reconcileError;
  $: rootCauseQuery = createRootCauseErrorQuery(
    instanceId,
    data,
    reconcileError,
  );
  $: rootCauseReconcileError = reconcileError
    ? ($rootCauseQuery?.data ?? reconcileError)
    : undefined;

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
  <CanvasInitialization
    {canvasName}
    {instanceId}
    let:ready
    let:isReconciling
    let:isLoading
  >
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
            disabled={!!parseError || !!reconcileError || resourceIsReconciling}
            reconciling={resourceIsReconciling}
          />
        </div>
      </WorkspaceHeader>

      <WorkspaceEditorContainer
        slot="body"
        error={parseError?.message ?? rootCauseReconcileError}
        showError={!!$remoteContent && selectedView === "code"}
      >
        {#if selectedView === "code"}
          <CanvasEditor
            bind:autoSave={$autoSave}
            {canvasName}
            {fileArtifact}
            {parseError}
          />
        {:else if selectedView === "viz"}
          <CanvasLoadingState
            {ready}
            {isReconciling}
            {isLoading}
            errorMessage={rootCauseReconcileError}
          >
            <CanvasBuilder
              {canvasName}
              openSidebar={workspace.inspector.open}
              {fileArtifact}
            />
          </CanvasLoadingState>
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
  </CanvasInitialization>
{/key}
