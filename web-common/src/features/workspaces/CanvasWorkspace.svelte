<script lang="ts">
  import { goto } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import Canvas from "@rilldata/web-common/features/canvas/Canvas.svelte";
  import CanvasEditor from "@rilldata/web-common/features/canvas/CanvasEditor.svelte";
  import CanvasThemeProvider from "@rilldata/web-common/features/canvas/CanvasThemeProvider.svelte";
  import { getComponentRegistry } from "@rilldata/web-common/features/canvas/components/util";
  import VisualCanvasEditing from "@rilldata/web-common/features/canvas/inspector/VisualCanvasEditing.svelte";
  import { useDefaultMetrics } from "@rilldata/web-common/features/canvas/selector";
  import StateManagersProvider from "@rilldata/web-common/features/canvas/state-managers/StateManagersProvider.svelte";
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
  import PreviewButton from "../explores/PreviewButton.svelte";
  import type { CanvasComponentType } from "../canvas/components/types";
  import { parseDocument } from "yaml";
  import { findNextAvailablePosition } from "../canvas/util";
  import AddComponentMenu from "../canvas/components/AddComponentMenu.svelte";

  export let fileArtifact: FileArtifact;

  let canvasName: string;
  let selectedView: "split" | "code" | "viz";

  const componentRegistry = getComponentRegistry();

  $: ({
    autoSave,
    path: filePath,
    fileName,
    updateEditorContent,
    editorContent,
    getResource,
    getAllErrors,
    remoteContent,
    hasUnsavedChanges,
    saveLocalContent,
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

  $: metricsViewQuery = useDefaultMetrics(instanceId);

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

  async function addComponent(componentType: CanvasComponentType) {
    const defaultMetrics = $metricsViewQuery?.data;
    if (!defaultMetrics) return;

    const newSpec = componentRegistry[componentType].newComponentSpec(
      defaultMetrics.metricsView,
      defaultMetrics.measure,
      defaultMetrics.dimension,
    );

    const { width, height } = componentRegistry[componentType].defaultSize;

    const parsedDocument = parseDocument($editorContent ?? "");
    const items = parsedDocument.get("items") as any;
    const itemsJson = parsedDocument.toJSON();
    const existingItems = itemsJson?.items || [];

    const [x, y] = findNextAvailablePosition(existingItems, width, height);

    const newComponent = {
      component: { [componentType]: newSpec },
      height,
      width,
      x,
      y,
    };

    if (!items) {
      parsedDocument.set("items", [newComponent]);
    } else {
      items.add(newComponent);
    }

    const newIndex = existingItems.length;
    updateEditorContent(parsedDocument.toString(), true);
    await saveLocalContent();
    scrollToComponent(newIndex);
  }

  function scrollToComponent(index: number) {
    setTimeout(() => {
      const component = document.querySelector(`[data-index="${index}"]`);
      if (component) {
        component.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 100);
  }
</script>

{#key canvasName}
  <StateManagersProvider {canvasName}>
    <CanvasThemeProvider>
      <WorkspaceContainer>
        <WorkspaceHeader
          slot="header"
          {filePath}
          resource={data}
          hasUnsavedChanges={$hasUnsavedChanges}
          titleInput={fileName}
          onTitleChange={onChangeCallback}
          resourceKind={ResourceKind.Canvas}
        >
          <div class="flex gap-x-2" slot="cta">
            <PreviewButton
              href="/canvas/{canvasName}"
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
                detail={allErrors.map((error) => error.message).join("\n")}
                header="Unable to load canvas preview"
                statusCode={404}
              />
            {:else if canvasResource}
              <Canvas {fileArtifact} />
            {/if}
          {/if}
        </WorkspaceEditorContainer>

        <VisualCanvasEditing
          {fileArtifact}
          autoSave={selectedView === "viz" || $autoSave}
          slot="inspector"
        />
      </WorkspaceContainer>
    </CanvasThemeProvider>
  </StateManagersProvider>
{/key}
