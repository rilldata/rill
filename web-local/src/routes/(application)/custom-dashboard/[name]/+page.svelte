<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CustomDashboard from "@rilldata/web-common/features/custom-dashboards/CustomDashboard.svelte";
  import CustomDashboardEditor from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEditor.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import type { V1DashboardSpec } from "@rilldata/web-common/runtime-client";
  import { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { Vector } from "@rilldata/web-common/features/custom-dashboards/types";
  import { stringify } from "yaml";
  import ViewSelector from "@rilldata/web-common/features/custom-dashboards/ViewSelector.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import { slide } from "svelte/transition";
  import AddChartMenu from "@rilldata/web-common/features/custom-dashboards/AddChartMenu.svelte";
  import ChartsEditor from "@rilldata/web-common/features/charts/editor/ChartsEditor.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { invalidate } from "$app/navigation";
  import { parse } from "yaml";

  const DEFAULT_EDITOR_HEIGHT = 300;
  const DEFAULT_EDITOR_WIDTH = 400;

  const updateFile = createRuntimeServicePutFile();

  export let data;

  let snap = false;
  let showGrid = true;
  let showChartEditor = false;
  let selectedView = "split";

  let containerWidth: number;
  let chartEditorHeight = DEFAULT_EDITOR_HEIGHT;
  let editorWidth = DEFAULT_EDITOR_WIDTH;
  let selectedChartName: string | null = null;

  let dashboard: V1DashboardSpec = {
    columns: 10,
    gap: 1,
    components: [],
  };

  $: instanceId = $runtime.instanceId;
  $: customDashboardName = data.dashboardName;
  $: ({ path, blob: yaml = "", error } = data.file);

  $: selectedChartFilePath =
    selectedChartName &&
    getFileAPIPathFromNameAndType(selectedChartName, EntityType.Chart);

  $: if (yaml) {
    try {
      dashboard = parse(yaml) as V1DashboardSpec;
    } catch {
      // Unable to parse YAML, no-op
    }
  }

  $: ({ columns, gap, components = [] } = dashboard);

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
      path,

      EntityType.Dashboard,
    );
  };

  async function updateChartFile(e: CustomEvent<string>) {
    const content = e.detail;
    if (!content) return;

    // Update the file
    try {
      await $updateFile.mutateAsync({
        instanceId,
        path,
        data: {
          blob: content,
        },
      });

      // Timeout to ensure the parser has updated
      await new Promise((resolve) => setTimeout(resolve, 400));
      await invalidate(path);
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
    const components = [...(dashboard?.components ?? [])];

    components[e.detail.index].width = e.detail.dimensions[0];
    components[e.detail.index].height = e.detail.dimensions[1];

    components[e.detail.index].x = e.detail.position[0];
    components[e.detail.index].y = e.detail.position[1];

    yaml = stringify(<V1DashboardSpec>{
      kind: "dashboard",
      ...dashboard,
      components,
    });

    await updateChartFile(new CustomEvent("update", { detail: yaml }));
  }

  async function addChart(e: CustomEvent<{ chartName: string }>) {
    if (!dashboard) return;
    const components = [...(dashboard?.components ?? [])];
    components.push({
      chart: e.detail.chartName,
      height: 4,
      width: 4,
      x: 0,
      y: 0,
    });

    yaml = stringify(<V1DashboardSpec>{
      kind: "dashboard",
      ...dashboard,
      components,
    });

    await updateChartFile(new CustomEvent("update", { detail: yaml }));
  }
</script>

<svelte:head>
  <title>Rill Developer | {customDashboardName}</title>
</svelte:head>

<WorkspaceContainer inspector={false} bind:width={containerWidth}>
  <WorkspaceHeader
    slot="header"
    titleInput={customDashboardName}
    showInspectorToggle={false}
    on:change={onChangeCallback}
  >
    <div slot="workspace-controls" class="flex gap-x-4 items-center">
      <ViewSelector bind:selectedView />

      <div
        class="flex gap-x-1 flex-none items-center h-full bg-white rounded-full"
      >
        <Switch small id="snap" bind:checked={snap} />
        <Label for="snap" class="font-normal text-xs">Snap on change</Label>
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
          min={100}
          max={0.6 * containerWidth}
        />
        <div class="flex flex-col h-full overflow-hidden">
          <section
            class="size-full flex flex-col flex-shrink overflow-hidden bg-white"
          >
            <CustomDashboardEditor {error} {yaml} on:update={updateChartFile} />
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

    {#if dashboard && (selectedView == "viz" || selectedView == "split")}
      <CustomDashboard
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
