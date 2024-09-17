<script lang="ts">
  import { goto } from "$app/navigation";
  import type { EditorView } from "@codemirror/view";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import ComponentsEditor from "@rilldata/web-common/features/canvas-components/editor/ComponentsEditor.svelte";
  import ComponentsEditorContainer from "@rilldata/web-common/features/canvas-components/editor/ComponentsEditorContainer.svelte";
  import AddComponentMenu from "@rilldata/web-common/features/canvas-dashboards/AddComponentMenu.svelte";
  import CanvasDashboardPreview from "@rilldata/web-common/features/canvas-dashboards/CanvasDashboardPreview.svelte";
  import type { Vector } from "@rilldata/web-common/features/canvas-dashboards/types";
  import ViewSelector from "@rilldata/web-common/features/canvas-dashboards/ViewSelector.svelte";
  import DeployDashboardCta from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardCTA.svelte";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { FileExtensionToEditorExtension } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import PreviewButton from "@rilldata/web-common/features/metrics-views/workspace/PreviewButton.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { V1DashboardSpec } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { setContext } from "svelte";
  import { slide } from "svelte/transition";
  import Button from "web-common/src/components/button/Button.svelte";
  import { parseDocument } from "yaml";

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
  let spec: V1DashboardSpec = {
    columns: 20,
    gap: 4,
    items: [],
  };

  $: canvasDashboardName = getNameFromFile(filePath);
  $: setContext("rill::canvas-dashboard:name", canvasDashboardName);

  $: ({ instanceId } = $runtime);

  $: errorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: errors = $errorsQuery;

  $: ({
    saveLocalContent: updateComponentFile,
    autoSave,
    path: filePath,
    fileName,
    updateLocalContent,
    localContent,
    remoteContent,
    hasUnsavedChanges,
  } = fileArtifact);

  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);

  $: spec = structuredClone($resourceQuery.data?.dashboard?.spec ?? spec);

  $: ({ items = [], columns = 20, gap = 4, variables = [] } = spec);
  $: if (
    items.filter(
      (item) =>
        !item.definedInDashboard && item.component === selectedComponentName,
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

  async function onChangeCallback(
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    const newRoute = await handleEntityRename(
      $runtime.instanceId,
      e.currentTarget,
      filePath,
      fileName,
      fileArtifacts.getNamesForKind(ResourceKind.Dashboard),
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
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");
    const items = parsedDocument.get("items") as any;

    const node = items.get(e.detail.index);

    node.set("width", e.detail.dimensions[0]);
    node.set("height", e.detail.dimensions[1]);
    node.set("x", e.detail.position[0]);
    node.set("y", e.detail.position[1]);

    updateLocalContent(parsedDocument.toString(), true);

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

  async function handleDeleteEvent(
    e: CustomEvent<{
      index: number;
    }>,
  ) {
    if (!e.detail.index) return;
    await deleteComponent(e.detail.index);
  }

  async function deleteComponent(index: number) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

    const items = parsedDocument.get("items") as any;

    if (!items) return;

    items.delete(index);

    updateLocalContent(parsedDocument.toString(), true);

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
    hasUnsavedChanges={$hasUnsavedChanges}
    on:change={onChangeCallback}
    showInspectorToggle={false}
    slot="header"
    titleInput={fileName}
  >
    <div class="flex gap-x-4 items-center" slot="workspace-controls">
      <ViewSelector bind:selectedView />

      {#if selectedView === "split" || selectedView === "viz"}
        <div
          class="flex gap-x-1 flex-none items-center h-full bg-white rounded-full"
        >
          <Switch small id="grid" bind:checked={showGrid} />
          <Label for="grid" class="font-normal text-xs">Grid</Label>
        </div>
      {/if}

      <AddComponentMenu {addComponent} />

      <PreviewButton
        dashboardName={canvasDashboardName}
        disabled={Boolean(errors.length)}
        type="custom"
      />

      <DeployDashboardCta />
      <LocalAvatarButton />
    </div>
  </WorkspaceHeader>

  <div class="flex w-full h-full flex-row overflow-hidden" slot="body">
    {#if selectedView === "code" || selectedView === "split"}
      <div
        transition:slide={{ duration: 400, axis: "x" }}
        class="relative h-full flex-shrink-0 w-full border-r"
        class:!w-full={selectedView === "code"}
        style:width="{editorPercentage * 100}%"
      >
        <Resizer
          direction="EW"
          side="right"
          dimension={editorWidth}
          min={300}
          max={0.65 * containerWidth}
          onUpdate={(width) => (editorPercentage = width / containerWidth)}
        />
        <div class="flex flex-col h-full overflow-hidden">
          <section class="size-full flex flex-col flex-shrink overflow-hidden">
            <ComponentsEditorContainer error={errors[0]}>
              <Editor
                bind:editor
                {fileArtifact}
                extensions={FileExtensionToEditorExtension[".yaml"]}
                autoSave
                showSaveBar={false}
                forceLocalUpdates
                onRevert={() => {
                  spec = structuredClone(spec);
                }}
              />
            </ComponentsEditorContainer>
          </section>

          <section
            style:height="{componentEditorPercentage * 100}%"
            class:!h-12={!showComponentEditor}
            class="size-full flex flex-col flex-none bg-white flex-shrink-0 relative border-t !min-h-12"
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
              class="flex justify-between items-center bg-gray-100 px-4 h-12 flex-none"
            >
              <h1 class="font-semibold text-[16px] truncate">
                {#if selectedComponentName}
                  {selectedComponentName}.yaml
                {:else}
                  Select a component to edit
                {/if}
              </h1>
              {#if selectedComponentName || showComponentEditor}
                <Button
                  type="text"
                  on:click={() => (showComponentEditor = !showComponentEditor)}
                >
                  {showComponentEditor ? "Close" : "Open"}
                </Button>
              {/if}
            </header>

            {#if showComponentEditor}
              <div class="size-full overflow-hidden">
                {#if selectedComponentFilePath}
                  <ComponentsEditor filePath={selectedComponentFilePath} />
                {/if}
              </div>
            {/if}
          </section>
        </div>
      </div>
    {/if}

    {#if selectedView === "viz" || selectedView === "split"}
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
    {/if}
  </div>
</WorkspaceContainer>
