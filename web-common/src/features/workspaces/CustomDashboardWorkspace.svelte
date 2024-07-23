<script lang="ts">
  import { goto } from "$app/navigation";
  import type { EditorView } from "@codemirror/view";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import ChartsEditor from "@rilldata/web-common/features/charts/editor/ChartsEditor.svelte";
  import ChartsEditorContainer from "@rilldata/web-common/features/charts/editor/ChartsEditorContainer.svelte";
  import AddChartMenu from "@rilldata/web-common/features/custom-dashboards/AddChartMenu.svelte";
  import CustomDashboardPreview from "@rilldata/web-common/features/custom-dashboards/CustomDashboardPreview.svelte";
  import ViewSelector from "@rilldata/web-common/features/custom-dashboards/ViewSelector.svelte";
  import type { Vector } from "@rilldata/web-common/features/custom-dashboards/types";
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

  let customDashboardName: string;
  let selectedChartFileArtifact: FileArtifact | undefined;
  let selectedView = "split";
  let showGrid = true;
  let showChartEditor = false;
  let containerWidth: number;
  let containerHeight: number;
  let editorPercentage = 0.5;
  let editor: EditorView;
  let selectedIndex: number | null = null;
  let chartEditorPercentage = 0.4;
  let selectedChartName: string | null = null;
  let spec: V1DashboardSpec = {
    columns: 20,
    gap: 4,
    items: [],
  };

  $: customDashboardName = getNameFromFile(filePath);
  $: setContext("rill::custom-dashboard:name", customDashboardName);

  $: ({ instanceId } = $runtime);

  $: errorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: errors = $errorsQuery;

  $: ({
    saveLocalContent: updateChartFile,
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
        !item.definedInDashboard && item.component === selectedChartName,
    ).length
  ) {
    selectedChartFileArtifact = fileArtifacts.findFileArtifact(
      ResourceKind.Component,
      selectedChartName ?? "",
    );
  } else {
    selectedChartName = null;
    selectedChartFileArtifact = undefined;
    showChartEditor = false;
  }
  $: selectedChartFilePath = selectedChartFileArtifact?.path;
  $: editorWidth = editorPercentage * containerWidth;
  $: chartEditorHeight = chartEditorPercentage * containerHeight;

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

    if ($autoSave) await updateChartFile();
  }

  async function addChart(chartName: string) {
    const newChart = {
      component: chartName,
      height: 4,
      width: 4,
      x: 0,
      y: 0,
    };
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

    const items = parsedDocument.get("items") as any;

    if (!items) {
      parsedDocument.set("items", [newChart]);
    } else {
      items.add(newChart);
    }

    updateLocalContent(parsedDocument.toString(), true);

    if ($autoSave) await updateChartFile();
  }

  async function handleDeleteEvent(
    e: CustomEvent<{
      index: number;
    }>,
  ) {
    if (!e.detail.index) return;
    await deleteChart(e.detail.index);
  }

  async function deleteChart(index: number) {
    const parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

    const items = parsedDocument.get("items") as any;

    if (!items) return;

    items.delete(index);

    updateLocalContent(parsedDocument.toString(), true);

    if ($autoSave) await updateChartFile();
  }
</script>

<svelte:window
  on:keydown={async (e) => {
    if (e.target !== document.body || selectedIndex === null) return;
    if (e.key === "Delete" || e.key === "Backspace") {
      await deleteChart(selectedIndex);
    }
  }}
/>

<WorkspaceContainer
  bind:width={containerWidth}
  bind:height={containerHeight}
  inspector={false}
>
  <WorkspaceHeader
    on:change={onChangeCallback}
    showInspectorToggle={false}
    slot="header"
    titleInput={fileName}
    hasUnsavedChanges={$hasUnsavedChanges}
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

      <AddChartMenu {addChart} />

      <PreviewButton
        dashboardName={customDashboardName}
        disabled={Boolean(errors.length)}
        type="custom"
      />
    </div>
  </WorkspaceHeader>

  <div class="flex w-full h-full flex-row overflow-hidden" slot="body">
    {#if selectedView == "code" || selectedView == "split"}
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
            <ChartsEditorContainer error={errors[0]}>
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
            </ChartsEditorContainer>
          </section>

          <section
            style:height="{chartEditorPercentage * 100}%"
            class:!h-12={!showChartEditor}
            class="size-full flex flex-col flex-none bg-white flex-shrink-0 relative border-t !min-h-12"
          >
            <Resizer
              direction="NS"
              dimension={chartEditorHeight}
              min={80}
              max={0.85 * containerHeight}
              onUpdate={(height) =>
                (chartEditorPercentage = height / containerHeight)}
            />
            <header
              class="flex justify-between items-center bg-gray-100 px-4 h-12 flex-none"
            >
              <h1 class="font-semibold text-[16px] truncate">
                {#if selectedChartName}
                  {selectedChartName}.yaml
                {:else}
                  Select a chart to edit
                {/if}
              </h1>
              {#if selectedChartName || showChartEditor}
                <Button
                  type="text"
                  on:click={() => (showChartEditor = !showChartEditor)}
                >
                  {showChartEditor ? "Close" : "Open"}
                </Button>
              {/if}
            </header>

            {#if showChartEditor}
              <div class="size-full overflow-hidden">
                {#if selectedChartFilePath}
                  <ChartsEditor filePath={selectedChartFilePath} />
                {/if}
              </div>
            {/if}
          </section>
        </div>
      </div>
    {/if}

    {#if selectedView == "viz" || selectedView == "split"}
      <CustomDashboardPreview
        {customDashboardName}
        {gap}
        {items}
        {columns}
        {showGrid}
        {variables}
        bind:selectedChartName
        bind:selectedIndex
        on:update={handlePreviewUpdate}
        on:delete={handleDeleteEvent}
      />
    {/if}
  </div>
</WorkspaceContainer>
