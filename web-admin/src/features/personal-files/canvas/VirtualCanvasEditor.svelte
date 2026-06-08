<script lang="ts">
  import CanvasInitialization from "@rilldata/web-common/features/canvas/CanvasInitialization.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import ReconcileWarningPanel from "@rilldata/web-common/features/entity-management/ReconcileWarningPanel.svelte";
  import VisualCanvasEditing from "@rilldata/web-common/features/canvas/inspector/VisualCanvasEditing.svelte";
  import SaveDefaultsButton from "@rilldata/web-common/features/canvas/components/SaveDefaultsButton.svelte";
  import CanvasLoadingState from "@rilldata/web-common/features/canvas/CanvasLoadingState.svelte";
  import CanvasBuilder from "@rilldata/web-common/features/canvas/CanvasBuilder.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { getPersonalFilteredResourceByName } from "@rilldata/web-admin/features/personal-files/selectors.ts";
  import { parseDocument } from "yaml";
  import { Button } from "@rilldata/web-common/components/button/index.ts";
  import { Play, Trash } from "lucide-svelte";

  let {
    fileArtifact,
    name,
    onPreview,
    onDelete,
  }: {
    fileArtifact: FileArtifact;
    name: string;
    onPreview: () => void;
    onDelete: () => void;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let {
    path: filePath,
    remoteContent,
    editorContent,
    saveState: { saving },
  } = $derived(fileArtifact);

  let resourceQuery = $derived(
    getPersonalFilteredResourceByName(runtimeClient, name),
  );

  let { data } = $derived($resourceQuery);
  $effect(() => {
    if (data) fileArtifact.updateResource(data);
  });
  let titleValue = $derived(data?.canvas?.spec?.displayName ?? name);

  let canvasName = $derived(getNameFromFile(filePath));

  // Parse error for the editor gutter and banner
  let parseErrorQuery = $derived(fileArtifact.getParseError(queryClient));
  let parseError = $derived($parseErrorQuery);

  let reconcileError = $derived(data?.meta?.reconcileError);
  let rootCauseQuery = $derived(
    createRootCauseErrorQuery(runtimeClient, data, reconcileError),
  );
  let rootCauseReconcileError = $derived(
    reconcileError ? ($rootCauseQuery?.data ?? reconcileError) : undefined,
  );

  function onTitleChange(newTitle: string) {
    const trimmed = newTitle.trim();
    if (!trimmed || trimmed === titleValue) return;

    const current = $editorContent ?? "";
    let yamlOut: string;
    try {
      const doc = parseDocument(current);
      doc.set("display_name", trimmed);
      yamlOut = doc.toString();
    } catch (e) {
      console.error("Failed to update display_name in YAML", e);
      return;
    }
    // Optimistic update
    titleValue = trimmed;
    fileArtifact.updateEditorContent(yamlOut, false, true);
  }
</script>

<CanvasInitialization
  {canvasName}
  instanceId={runtimeClient.instanceId}
  allowUnvalidatedSpec={true}
  let:ready
  let:isReconciling
  let:isLoading
>
  <WorkspaceContainer>
    <WorkspaceHeader
      slot="header"
      {filePath}
      resource={data}
      hasUnsavedChanges={false}
      titleInput={titleValue}
      {onTitleChange}
      codeToggle={false}
      resourceKind={ResourceKind.Canvas}
      showBreadcrumbs={false}
    >
      {#snippet cta()}
        <div class="flex gap-x-2">
          {#if ready}
            <SaveDefaultsButton
              {canvasName}
              instanceId={runtimeClient.instanceId}
              saving={$saving}
            />
          {/if}
          <Button label="Preview" type="secondary" compact onClick={onDelete}>
            <Trash size={14} />
            Delete
          </Button>
          <Button label="Preview" type="secondary" compact onClick={onPreview}>
            <Play size={14} />
            Preview
          </Button>
        </div>
      {/snippet}
    </WorkspaceHeader>

    <svelte:fragment slot="body">
      <div class="flex flex-col h-full">
        <div class="flex-1 min-h-0">
          <WorkspaceEditorContainer
            resource={data}
            {parseError}
            remoteContent={$remoteContent}
            {filePath}
            showError={ready}
          >
            <CanvasLoadingState
              {ready}
              {isReconciling}
              {isLoading}
              errorMessage={rootCauseReconcileError}
              {filePath}
            >
              <CanvasBuilder
                {canvasName}
                {fileArtifact}
                openSidebar={() => {}}
              />
            </CanvasLoadingState>
          </WorkspaceEditorContainer>
        </div>
        <ReconcileWarningPanel {fileArtifact} />
      </div>
    </svelte:fragment>
    <svelte:fragment slot="inspector">
      {#if ready}
        <VisualCanvasEditing {canvasName} {fileArtifact} autoSave />
      {/if}
    </svelte:fragment>
  </WorkspaceContainer>
</CanvasInitialization>
