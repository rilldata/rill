<script lang="ts">
  import { goto } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import CanvasEditor from "@rilldata/web-common/features/canvas/CanvasEditor.svelte";
  import VisualCanvasEditing from "@rilldata/web-common/features/canvas/inspector/VisualCanvasEditing.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
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
  import DelayedSpinner from "../entity-management/DelayedSpinner.svelte";
  import { useCanvas } from "../canvas/selector";

  export let fileArtifact: FileArtifact;

  let canvasName: string;
  let selectedView: "split" | "code" | "viz";

  $: ({
    autoSave,
    path: filePath,
    fileName,
    getResource,
    getAllErrors,
    remoteContent,
    hasUnsavedChanges,
  } = fileArtifact);

  $: ({
    canvasEntity: { _rows },
  } = getCanvasStore(canvasName, instanceId));

  $: resourceQuery = getResource(queryClient, instanceId);

  $: ({ data } = $resourceQuery);

  $: allErrorsQuery = getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;

  $: resourceIsReconciling = resourceIsLoading(data);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;
  $: selectedView = $selectedViewStore ?? "code";

  $: canvasName = getNameFromFile(filePath);

  $: ({ instanceId } = $runtime);

  $: lineBasedRuntimeErrors = mapParseErrorsToLines(
    allErrors,
    $remoteContent ?? "",
  );

  $: mainError = lineBasedRuntimeErrors?.at(0);

  $: canvasResolverQuery = useCanvas(instanceId, canvasName);
  $: canvasResolverQueryResult = $canvasResolverQuery;
  $: canvasData = canvasResolverQueryResult.data;

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
        {#if mainError}
          <ErrorPage
            body={mainError.message}
            fatal
            detail={allErrors.map((error) => error.message).join("\n")}
            header="Unable to load canvas preview"
            statusCode={404}
          />
        {:else if canvasResolverQueryResult.isLoading}
          <DelayedSpinner isLoading={true} size="48px" />
        {:else if canvasData}
          <CanvasBuilder
            {canvasName}
            openSidebar={workspace.inspector.open}
            {fileArtifact}
          />
        {/if}
      {/if}
    </WorkspaceEditorContainer>

    <svelte:fragment slot="inspector">
      {#key $_rows}
        <VisualCanvasEditing
          {canvasName}
          {fileArtifact}
          autoSave={selectedView === "viz" || $autoSave}
        />
      {/key}
    </svelte:fragment>
  </WorkspaceContainer>
{/key}
