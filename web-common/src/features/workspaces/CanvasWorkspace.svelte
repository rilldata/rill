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
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import LeaderboardIcon from "../canvas/icons/LeaderboardIcon.svelte";
  import CheckCircleNew from "@rilldata/web-common/components/icons/CheckCircleNew.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";

  export let fileArtifact: FileArtifact;

  let canvasName: string;
  let selectedView: "split" | "code" | "viz";
  let justClickedSaveAsDefault = false;

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

  $: ({
    canvasEntity: { _rows, setDefaultFilters, _viewingDefaults },
  } = getCanvasStore(canvasName, instanceId));

  $: viewingDefaults = $_viewingDefaults;

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
        <Button
          label="Preview"
          type={!viewingDefaults ? "secondary" : "ghost"}
          preload={false}
          disabled={viewingDefaults}
          onClick={async () => {
            justClickedSaveAsDefault = true;
            await setDefaultFilters();
            setTimeout(() => {
              justClickedSaveAsDefault = false;
            }, 2500);
          }}
        >
          {#if $saving && justClickedSaveAsDefault}
            <LoadingSpinner size="15px" />
            <div class="flex gap-x-1 items-center">Saving default filters</div>
          {:else if viewingDefaults}
            {#if justClickedSaveAsDefault}
              <CheckCircleNew size="15px" className="fill-green-600" />
              <div class="flex gap-x-1 items-center text-green-600">
                Saved default filters
              </div>
            {:else}
              <LeaderboardIcon size="16px" color="currentColor" />
              <div class="flex gap-x-1 items-center">Viewing default state</div>
            {/if}
          {:else}
            <LeaderboardIcon size="16px" color="currentColor" />
            <div class="flex gap-x-1 items-center">Save as default</div>
          {/if}
        </Button>
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
