<script lang="ts">
  import { goto } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import Canvas from "@rilldata/web-common/features/canvas/Canvas.svelte";
  import CanvasEditor from "@rilldata/web-common/features/canvas/CanvasEditor.svelte";
  import CanvasThemeProvider from "@rilldata/web-common/features/canvas/CanvasThemeProvider.svelte";
  import AddComponentMenu from "@rilldata/web-common/features/canvas/components/AddComponentMenu.svelte";
  import VisualCanvasEditing from "@rilldata/web-common/features/canvas/inspector/VisualCanvasEditing.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/canvas/state-managers/StateManagersProvider.svelte";
  import CanvasStateProvider from "@rilldata/web-common/features/canvas/stores/CanvasStateProvider.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    resourceIsLoading,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import ViewSelector from "@rilldata/web-common/features/visual-editing/ViewSelector.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { setContext } from "svelte";
  import { parseDocument } from "yaml";
  import PreviewButton from "../explores/PreviewButton.svelte";

  export let fileArtifact: FileArtifact;

  let canvasName: string;
  let selectedView: "split" | "code" | "viz";

  $: ({
    saveLocalContent: updateComponentFile,
    autoSave,
    path: filePath,
    fileName,
    updateLocalContent,
    localContent,
    getResource,
    getAllErrors,
    remoteContent,
    hasUnsavedChanges,
  } = fileArtifact);

  $: resourceQuery = getResource(queryClient, instanceId);

  $: ({ data } = $resourceQuery);

  $: allErrorsQuery = getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceIsReconciling = resourceIsLoading(data);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;
  $: selectedView = $selectedViewStore ?? "code";

  $: canvasResource = data?.canvas;

  $: canvasName = getNameFromFile(filePath);
  $: setContext("rill::canvas:name", canvasName);

  $: ({ instanceId } = $runtime);

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

  async function addComponent(componentName: string) {
    const newComponent = {
      component: componentName,
      height: 4,
      width: 4,
      x: 0,
      y: 0,
    };
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

    const items = parsedDocument.get("items") as any;

    if (!items) {
      parsedDocument.set("items", [newComponent]);
    } else {
      items.add(newComponent);
    }

    updateLocalContent(parsedDocument.toString(), true);

    if ($autoSave) await updateComponentFile();
  }
</script>

{#if canvasResource && fileArtifact}
  {#key canvasName}
    <StateManagersProvider {canvasName} {canvasResource} {fileArtifact}>
      <CanvasStateProvider>
        <CanvasThemeProvider>
          <WorkspaceContainer>
            <WorkspaceHeader
              slot="header"
              {filePath}
              hasUnsavedChanges={$hasUnsavedChanges}
              titleInput={fileName}
              onTitleChange={onChangeCallback}
              resourceKind={ResourceKind.Canvas}
            >
              <div class="flex gap-x-2" slot="cta">
                <PreviewButton
                  href="/custom/{canvasName}"
                  disabled={allErrors.length > 0 || resourceIsReconciling}
                  reconciling={resourceIsReconciling}
                />

                <AddComponentMenu {addComponent} />
                <ViewSelector
                  allowSplit={false}
                  bind:selectedView={$selectedViewStore}
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
                    header="Unable to load dashboard preview"
                    statusCode={404}
                  />
                {:else if canvasResource}
                  <Canvas />
                {/if}
              {/if}
            </WorkspaceEditorContainer>

            <VisualCanvasEditing slot="inspector" />
          </WorkspaceContainer>
        </CanvasThemeProvider>
      </CanvasStateProvider>
    </StateManagersProvider>
  {/key}
{:else}
  <div class="grid place-items-center size-full">
    <DelayedSpinner isLoading={true} size="40px" />
  </div>
{/if}
