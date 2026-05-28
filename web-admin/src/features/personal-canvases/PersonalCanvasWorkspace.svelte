<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    V1PersonalVirtualFileType,
    adminServiceDeletePersonalVirtualFile,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import CanvasBuilder from "@rilldata/web-common/features/canvas/CanvasBuilder.svelte";
  import CanvasEditor from "@rilldata/web-common/features/canvas/CanvasEditor.svelte";
  import CanvasInitialization from "@rilldata/web-common/features/canvas/CanvasInitialization.svelte";
  import CanvasLoadingState from "@rilldata/web-common/features/canvas/CanvasLoadingState.svelte";
  import SaveDefaultsButton from "@rilldata/web-common/features/canvas/components/SaveDefaultsButton.svelte";
  import VisualCanvasEditing from "@rilldata/web-common/features/canvas/inspector/VisualCanvasEditing.svelte";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ReconcileWarningPanel from "@rilldata/web-common/features/entity-management/ReconcileWarningPanel.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { parseDocument } from "yaml";
  import type { VirtualFilePersistence } from "./VirtualFilePersistence";

  export let persistence: VirtualFilePersistence;
  export let canvasName: string;
  /** Initial display name. The header reflects this and updates after rename. */
  export let initialDisplayName: string;

  const runtimeClient = useRuntimeClient();

  $: ({
    autoSave,
    path: filePath,
    getResource,
    remoteContent,
    hasUnsavedChanges,
    saveState: { saving },
  } = persistence);

  $: resourceQuery = getResource(queryClient);
  $: ({ data } = $resourceQuery);

  $: resourceIsReconciling = resourceIsLoading(data);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;
  $: selectedView = $selectedViewStore ?? "viz";

  // Parse error for the editor gutter and banner
  $: parseErrorQuery = persistence.getParseError(queryClient);
  $: parseError = $parseErrorQuery;

  $: reconcileError = data?.meta?.reconcileError;
  $: rootCauseQuery = createRootCauseErrorQuery(
    runtimeClient,
    data,
    reconcileError,
  );
  $: rootCauseReconcileError = reconcileError
    ? ($rootCauseQuery?.data ?? reconcileError)
    : undefined;

  // The title shown in the header reflects the YAML display_name. Track it locally so
  // a rename round-trips immediately without waiting for the remote refetch.
  let titleValue = initialDisplayName;
  $: if ($remoteContent) {
    try {
      const doc = parseDocument($remoteContent);
      const name = (doc.get("display_name") as string | undefined) ?? undefined;
      if (name) titleValue = name;
    } catch (e) {
      /* ignore */
    }
  }

  async function onTitleChange(newTitle: string) {
    const trimmed = newTitle.trim();
    if (!trimmed || trimmed === titleValue) return;

    const current = $remoteContent ?? "";
    let yamlOut: string;
    try {
      const doc = parseDocument(current);
      doc.set("display_name", trimmed);
      yamlOut = doc.toString();
    } catch (e) {
      console.error("Failed to update display_name in YAML", e);
      return;
    }
    persistence.updateEditorContent(yamlOut, false, true);
    titleValue = trimmed;
  }

  async function done() {
    // Force-flush any pending debounced edits before navigating away. The admin
    // EditPersonalVirtualFile RPC blocks on TriggerParserAndAwaitResource, so the
    // canvas resource will be updated by the time saveLocalContent resolves.
    try {
      await persistence.saveLocalContent(true);
    } catch (e) {
      console.error("Save before exiting edit mode failed", e);
    }
    // Invalidate the resource caches so the preview re-fetches the reconciled canvas
    // (otherwise the user might see a stale render right after their edit).
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGetResourceQueryKey(runtimeClient.instanceId, {
        name: { kind: ResourceKind.Canvas, name: canvasName },
      }),
    });
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
        {},
      ),
    });

    const url = new URL($page.url);
    url.searchParams.delete("mode");
    await goto(url.toString(), { replaceState: false, keepFocus: true });
  }

  async function deleteCanvas() {
    if (!confirm(`Delete "${titleValue}"? This cannot be undone.`)) return;
    await adminServiceDeletePersonalVirtualFile(
      $page.params.organization,
      $page.params.project,
      V1PersonalVirtualFileType.PERSONAL_VIRTUAL_FILE_TYPE_CANVAS,
      canvasName,
    );
    await goto(`/${$page.params.organization}/${$page.params.project}`);
  }
</script>

{#key canvasName}
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
        hasUnsavedChanges={$hasUnsavedChanges}
        titleInput={titleValue}
        codeToggle
        showBreadcrumbs={false}
        {onTitleChange}
        resourceKind={ResourceKind.Canvas}
      >
        {#snippet cta()}
          <div class="flex gap-x-2 items-center">
            {#if ready}
              <SaveDefaultsButton
                {canvasName}
                instanceId={runtimeClient.instanceId}
                saving={$saving}
              />
            {/if}
            <Button type="secondary" onClick={deleteCanvas}>Delete</Button>
            <Button type="primary" onClick={done}>Done editing</Button>
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
              showError={selectedView === "code" || ready}
            >
              {#if selectedView === "code"}
                <CanvasEditor
                  bind:autoSave={$autoSave}
                  {canvasName}
                  fileArtifact={persistence}
                  {parseError}
                />
              {:else if selectedView === "viz"}
                <CanvasLoadingState
                  {ready}
                  {isReconciling}
                  {isLoading}
                  errorMessage={rootCauseReconcileError}
                  {filePath}
                >
                  <CanvasBuilder
                    {canvasName}
                    openSidebar={workspace.inspector.open}
                    fileArtifact={persistence}
                  />
                </CanvasLoadingState>
              {/if}
            </WorkspaceEditorContainer>
          </div>
          <ReconcileWarningPanel fileArtifact={persistence} />
        </div>
      </svelte:fragment>

      <svelte:fragment slot="inspector">
        {#if ready}
          <VisualCanvasEditing
            {canvasName}
            fileArtifact={persistence}
            autoSave={selectedView === "viz" || $autoSave}
          />
        {/if}
      </svelte:fragment>
    </WorkspaceContainer>
  </CanvasInitialization>
{/key}
