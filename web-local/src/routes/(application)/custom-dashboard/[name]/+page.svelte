<script lang="ts">
  import { page } from "$app/stores";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import ChartsEditor from "@rilldata/web-common/features/charts/editor/ChartsEditor.svelte";
  import AddChartMenu from "@rilldata/web-common/features/custom-dashboards/AddChartMenu.svelte";
  import CustomDashboardEditor from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEditor.svelte";
  import CustomDashboardPreview from "@rilldata/web-common/features/custom-dashboards/CustomDashboardPreview.svelte";
  import ViewSelector from "@rilldata/web-common/features/custom-dashboards/ViewSelector.svelte";
  import type { Vector } from "@rilldata/web-common/features/custom-dashboards/types";
  import {
    getFileAPIPathFromNameAndType,
    removeLeadingSlash,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    FileArtifact,
    fileArtifacts,
  } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type {
    V1DashboardComponent,
    V1DashboardSpec,
  } from "@rilldata/web-common/runtime-client";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { slide } from "svelte/transition";
  import { parse, stringify } from "yaml";

  export let data: { fileArtifact?: FileArtifact } = {};

  let fileArtifact: FileArtifact;
  let filePath: string;
  let customDashboardName: string;
  $: if (data.fileArtifact) {
    fileArtifact = data.fileArtifact;
    filePath = fileArtifact.path;
    customDashboardName = fileArtifact.getEntityName();
  } else {
    customDashboardName = $page.params.name;
    filePath = getFileAPIPathFromNameAndType(
      customDashboardName,
      EntityType.Dashboard,
    );
    fileArtifact = fileArtifacts.getFileArtifact(filePath);
  }

  const DEFAULT_EDITOR_HEIGHT = 300;
  const DEFAULT_EDITOR_WIDTH = 400;

  const updateFile = createRuntimeServicePutFile();

  let selectedView = "split";
  let showGrid = true;
  let snap = false;
  let showChartEditor = false;
  let containerWidth: number;
  let editorWidth = DEFAULT_EDITOR_WIDTH;
  let chartEditorHeight = DEFAULT_EDITOR_HEIGHT;
  let selectedChartName: string | null = null;
  let dashboard: V1DashboardSpec = {
    columns: 10,
    gap: 1,
    components: [],
  };

  $: instanceId = $runtime.instanceId;

  $: errors = fileArtifact.getAllErrors(queryClient, instanceId);
  $: fileQuery = createRuntimeServiceGetFile(
    $runtime.instanceId,
    removeLeadingSlash(filePath),
  );
  $: [, fileName] = splitFolderAndName(filePath);

  $: yaml = $fileQuery.data?.blob || "";

  $: if (yaml) {
    try {
      const potentialDb = parse(yaml) as V1DashboardSpec;
      dashboard = {
        ...potentialDb,
        components: potentialDb.components?.filter(isComponent) ?? [],
      };
    } catch {
      // Unable to parse YAML, no-op
    }
  }

  $: selectedChartFileArtifact = fileArtifacts.findFileArtifact(
    ResourceKind.Chart,
    selectedChartName ?? "",
  );
  $: selectedChartFilePath = selectedChartFileArtifact?.path;

  $: ({ columns, gap, components = [] } = dashboard ?? ({} as V1DashboardSpec));

  const onChangeCallback = async (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => {
    if (!e.currentTarget) return;
    if (!e.currentTarget.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Model name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.currentTarget.value = customDashboardName; // resets the input
      return;
    }
    await handleEntityRename(
      instanceId,
      e.currentTarget,
      filePath,
      customDashboardName,
    );
  };

  async function updateChartFile(e: CustomEvent<string>) {
    const content = e.detail;
    if (!content) return;
    try {
      await $updateFile.mutateAsync({
        instanceId,
        path: removeLeadingSlash(filePath),
        data: {
          blob: content,
        },
      });
    } catch (err) {
      console.error(err);
    }
  }

  async function handlePreviewUpdate(
    e: CustomEvent<{
      index: number;
      position: Vector;
      dimensions: Vector;
    }>,
  ) {
    const newComponents = [...components];

    newComponents[e.detail.index].width = e.detail.dimensions[0];
    newComponents[e.detail.index].height = e.detail.dimensions[1];

    newComponents[e.detail.index].x = e.detail.position[0];
    newComponents[e.detail.index].y = e.detail.position[1];

    yaml = stringify(<V1DashboardSpec>{
      kind: "dashboard",
      ...dashboard,
      components: newComponents,
    });

    await updateChartFile(new CustomEvent("update", { detail: yaml }));
  }

  async function addChart(e: CustomEvent<{ chartName: string }>) {
    const newComponents = [...components];
    newComponents.push({
      chart: e.detail.chartName,
      height: 4,
      width: 4,
      x: 0,
      y: 0,
    });

    yaml = stringify(<V1DashboardSpec>{
      kind: "dashboard",
      ...dashboard,
      newComponents,
    });

    await updateChartFile(new CustomEvent("update", { detail: yaml }));
  }

  function isComponent(
    component: V1DashboardComponent | null | undefined,
  ): component is V1DashboardComponent {
    return !!component;
  }
</script>

<svelte:head>
  <title>Rill Developer | {customDashboardName}</title>
</svelte:head>

<WorkspaceContainer bind:width={containerWidth} inspector={false}>
  <WorkspaceHeader
    on:change={onChangeCallback}
    showInspectorToggle={false}
    slot="header"
    titleInput={fileName}
  >
    <div class="flex gap-x-4 items-center" slot="workspace-controls">
      <ViewSelector bind:selectedView />

      <div
        class="flex gap-x-1 flex-none items-center h-full bg-white rounded-full"
      >
        <Switch bind:checked={snap} id="snap" small />
        <Label class="font-normal text-xs" for="snap">Snap on change</Label>
      </div>

      {#if selectedView === "split" || selectedView === "viz"}
        <div
          class="flex gap-x-1 flex-none items-center h-full bg-white rounded-full"
        >
          <Switch small id="grid" bind:checked={showGrid} />
          <Label for="grid" class="font-normal text-xs">Grid</Label>
        </div>
      {/if}

      <AddChartMenu on:add-chart={addChart} />
      <Button>Preview</Button>
    </div>
  </WorkspaceHeader>

  <div class="flex w-full h-full flex-row overflow-hidden" slot="body">
    {#if selectedView == "code" || selectedView == "split"}
      <div
        transition:slide={{ duration: 400, axis: "x" }}
        class="relative h-full flex-shrink-0 w-full"
        class:!w-full={selectedView === "code"}
        style:width="{editorWidth}px"
      >
        <Resizer
          direction="EW"
          side="right"
          bind:dimension={editorWidth}
          min={300}
          max={0.6 * containerWidth}
        />
        <div class="flex flex-col h-full overflow-hidden">
          <section class="size-full flex flex-col flex-shrink overflow-hidden">
            <CustomDashboardEditor
              errors={$errors}
              {yaml}
              on:update={updateChartFile}
            />
          </section>

          <section
            class="size-full flex flex-col bg-white flex-shrink-0 relative h-fit"
          >
            <Resizer
              max={500}
              direction="NS"
              bind:dimension={chartEditorHeight}
            />
            <header
              class="flex justify-between items-center bg-gray-100 px-4 h-12"
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
              <div style:height="{chartEditorHeight}px">
                {#if selectedChartFilePath && showChartEditor}
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
        {snap}
        {gap}
        {components}
        {columns}
        {showGrid}
        bind:selectedChartName
        on:update={handlePreviewUpdate}
      />
    {/if}
  </div>
</WorkspaceContainer>
