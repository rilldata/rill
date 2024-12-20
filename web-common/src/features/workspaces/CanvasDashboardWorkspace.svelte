<script lang="ts">
  import { goto } from "$app/navigation";
  import type { EditorView } from "@codemirror/view";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import ComponentsEditor from "@rilldata/web-common/features/canvas-components/editor/ComponentsEditor.svelte";
  import ComponentsEditorContainer from "@rilldata/web-common/features/canvas-components/editor/ComponentsEditorContainer.svelte";
  import AddComponentMenu from "@rilldata/web-common/features/canvas/AddComponentMenu.svelte";
  import CanvasDashboardPreview from "@rilldata/web-common/features/canvas/CanvasDashboardPreview.svelte";
  import ViewSelector from "@rilldata/web-common/features/visual-editing/ViewSelector.svelte";
  import type { Vector } from "@rilldata/web-common/features/canvas/types";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { FileExtensionToEditorExtension } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { setContext } from "svelte";
  import Button from "web-common/src/components/button/Button.svelte";
  import { parseDocument } from "yaml";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "../entity-management/resource-icon-mapping";
  import PreviewButton from "../explores/PreviewButton.svelte";

  export let fileArtifact: FileArtifact;

  let canvasDashboardName: string;
  let selectedComponentFileArtifact: FileArtifact | undefined;
  let selectedView: "split" | "code" | "viz";
  let showGrid = true;
  let showComponentEditor = false;
  let containerWidth: number;
  let containerHeight: number;
  let editorPercentage = 0.5;
  let editor: EditorView;
  let selectedIndex: number | null = null;
  let componentEditorPercentage = 0.4;
  let selectedComponentName: string | null = null;
  let spec: V1CanvasSpec = {
    columns: 20,
    gap: 4,
    items: [],
  };

  $: canvasDashboardName = getNameFromFile(filePath);
  $: setContext("rill::canvas:name", canvasDashboardName);

  $: ({ instanceId } = $runtime);

  $: errorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: errors = $errorsQuery;

  $: ({
    saveLocalContent: updateComponentFile,
    autoSave,
    path: filePath,
    fileName,
    updateEditorContent,
    editorContent,
    hasUnsavedChanges,
  } = fileArtifact);

  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);

  $: spec = structuredClone($resourceQuery.data?.canvas?.spec ?? spec);

  $: ({ items = [], columns = 20, gap = 4, variables = [] } = spec);
  $: if (
    items.filter(
      (item) =>
        !item.definedInCanvas && item.component === selectedComponentName,
    ).length
  ) {
    selectedComponentFileArtifact = fileArtifacts.findFileArtifact(
      ResourceKind.Component,
      selectedComponentName ?? "",
    );
  } else {
    selectedComponentName = null;
    selectedComponentFileArtifact = undefined;
    showComponentEditor = false;
  }
  $: selectedComponentFilePath = selectedComponentFileArtifact?.path;
  $: editorWidth = editorPercentage * containerWidth;
  $: componentEditorHeight = componentEditorPercentage * containerHeight;

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      $runtime.instanceId,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }

  async function handlePreviewUpdate(
    e: CustomEvent<{
      index: number;
      position: Vector;
      dimensions: Vector;
    }>,
  ) {
    const parsedDocument = parseDocument($editorContent ?? "");
    const items = parsedDocument.get("items") as any;

    const node = items.get(e.detail.index);

    node.set("width", e.detail.dimensions[0]);
    node.set("height", e.detail.dimensions[1]);
    node.set("x", e.detail.position[0]);
    node.set("y", e.detail.position[1]);

    updateEditorContent(parsedDocument.toString());

    if ($autoSave) await updateComponentFile();
  }

  async function addComponent(componentName: string) {
    const newComponent = {
      component: componentName,
      height: 4,
      width: 4,
      x: 0,
      y: 0,
    };
    const parsedDocument = parseDocument($editorContent ?? "");

    const items = parsedDocument.get("items") as any;

    if (!items) {
      parsedDocument.set("items", [newComponent]);
    } else {
      items.add(newComponent);
    }

    updateEditorContent(parsedDocument.toString(), true);

    if ($autoSave) await updateComponentFile();
  }

  async function handleDeleteEvent(
    e: CustomEvent<{
      index: number;
    }>,
  ) {
    if (!e.detail.index) return;
    await deleteComponent(e.detail.index);
  }

  async function deleteComponent(index: number) {
    const parsedDocument = parseDocument($editorContent ?? "");

    const items = parsedDocument.get("items") as any;

    if (!items) return;

    items.delete(index);

    updateEditorContent(parsedDocument.toString(), true);

    if ($autoSave) await updateComponentFile();
  }
</script>

<svelte:window
  on:keydown={async (e) => {
    if (e.target !== document.body || selectedIndex === null) return;
    if (e.key === "Delete" || e.key === "Backspace") {
      await deleteComponent(selectedIndex);
    }
  }}
/>

<WorkspaceContainer
  bind:height={containerHeight}
  bind:width={containerWidth}
  inspector={false}
>
  <WorkspaceHeader
    slot="header"
    {filePath}
    resourceKind={ResourceKind.Canvas}
    hasUnsavedChanges={$hasUnsavedChanges}
    showInspectorToggle={false}
    titleInput={fileName}
    onTitleChange={onChangeCallback}
  >
    <div class="flex gap-x-2 items-center" slot="workspace-controls">
      <PreviewButton
        href="/custom/{canvasDashboardName}"
        disabled={errors?.length > 0}
      />

      <AddComponentMenu {addComponent} />
      <ViewSelector bind:selectedView />
    </div>
  </WorkspaceHeader>

  <div class="flex w-full h-full flex-row overflow-hidden" slot="body">
    {#if selectedView === "code" || selectedView === "split"}
      <div
        class="relative h-full flex-shrink-0 w-full"
        class:!w-full={selectedView === "code"}
        style:width="{editorPercentage * 100}%"
      >
        <div class="flex flex-col h-full overflow-hidden">
          <section class="size-full flex flex-col flex-shrink overflow-hidden">
            <ComponentsEditorContainer error={errors[0]}>
              <Editor
                bind:editor
                {fileArtifact}
                extensions={FileExtensionToEditorExtension[".yaml"]}
                autoSave
                showSaveBar={false}
                onRevert={() => {
                  spec = structuredClone(spec);
                }}
              />
            </ComponentsEditorContainer>
          </section>

          {#if selectedComponentName || showComponentEditor}
            <section
              style:height="{componentEditorPercentage * 100}%"
              class:!h-12={!showComponentEditor}
              class="size-full flex flex-col flex-none flex-shrink-0 relative !min-h-12"
            >
              <Resizer
                direction="NS"
                dimension={componentEditorHeight}
                min={80}
                max={0.85 * containerHeight}
                onUpdate={(height) =>
                  (componentEditorPercentage = height / containerHeight)}
              />
              <header
                class="flex justify-between items-center pr-2 bg-gray-100 flex-none py-2"
              >
                <h1
                  class="font-semibold text-xl truncate flex items-center gap-x-2"
                >
                  {#if selectedComponentName}
                    <svelte:component
                      this={resourceIconMapping[ResourceKind.Component]}
                      size="18px"
                      color={resourceColorMapping[ResourceKind.Component]}
                    />
                    {selectedComponentName}.yaml
                  {/if}
                </h1>

                <Button
                  type="subtle"
                  on:click={() => (showComponentEditor = !showComponentEditor)}
                >
                  {showComponentEditor ? "Close" : "Open"}
                </Button>
              </header>

              {#if showComponentEditor}
                <div class="size-full overflow-hidden">
                  {#if selectedComponentFilePath}
                    <ComponentsEditor filePath={selectedComponentFilePath} />
                  {/if}
                </div>
              {/if}
            </section>
          {/if}
        </div>
      </div>
    {/if}

    {#if selectedView === "split"}
      <Resizer
        absolute={false}
        direction="EW"
        side="right"
        dimension={editorWidth}
        min={300}
        max={0.65 * containerWidth}
        onUpdate={(width) => (editorPercentage = width / containerWidth)}
      />
    {/if}

    {#if selectedView === "viz" || selectedView === "split"}
      <section
        class="size-full flex flex-col relative overflow-hidden border border-gray-300 rounded-[2px]"
      >
        <CanvasDashboardPreview
          {canvasDashboardName}
          {gap}
          {items}
          {columns}
          {showGrid}
          {variables}
          bind:selectedComponentName
          bind:selectedIndex
          on:update={handlePreviewUpdate}
          on:delete={handleDeleteEvent}
        />

        <div class="floating-grid-wrapper">
          <Switch small id="grid" bind:checked={showGrid} />
          <Label for="grid" class="font-medium text-xs text-gray-600">
            Grid
          </Label>
        </div>
      </section>
    {/if}
  </div>
</WorkspaceContainer>

<style lang="postcss">
  .floating-grid-wrapper {
    @apply transition-all;
    @apply opacity-50 shadow-lg border border-slate-200 bg-slate-100;
    @apply flex gap-x-1 flex-none py-1 px-2 items-center h-fit rounded-full;
    @apply absolute bottom-2 right-2;
  }

  .floating-grid-wrapper:hover {
    @apply opacity-100;
  }
</style>
