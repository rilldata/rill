<script lang="ts">
  import { page } from "$app/stores";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useCustomDashboard } from "@rilldata/web-common/features/custom-dashboards/selectors";
  import CustomDashboard from "@rilldata/web-common/features/custom-dashboards/CustomDashboard.svelte";
  import CustomDashboardEditor from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEditor.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { goto } from "$app/navigation";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import { getRouteFromName } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { isDuplicateName } from "@rilldata/web-common/features/entity-management/name-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { Vector } from "@rilldata/web-common/features/custom-dashboards/types";
  import { parse, stringify } from "yaml";
  import type { V1DashboardSpec } from "@rilldata/web-common/runtime-client";
  import Viz from "@rilldata/web-common/components/icons/Viz.svelte";
  import Split from "@rilldata/web-common/components/icons/Split.svelte";
  import Code from "@rilldata/web-common/components/icons/Code.svelte";
  import Toggle from "@rilldata/web-common/features/custom-dashboards/Toggle.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import { slide } from "svelte/transition";
  import AddChartMenu from "@rilldata/web-common/features/custom-dashboards/AddChartMenu.svelte";

  const updateFile = createRuntimeServicePutFile();

  let showGrid = true;
  let editorWidth = 400;
  let startingWidth = 400;
  let startingX = 0;
  let currentX = 0;

  $: customDashboardName = $page.params.name;

  $: filePath = getFilePathFromNameAndType(
    customDashboardName,
    EntityType.Dashboard,
  );

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath);

  $: query = useCustomDashboard($runtime.instanceId, customDashboardName);
  $: allNamesQuery = useAllNames($runtime.instanceId);

  $: yaml = $fileQuery.data?.blob || "";

  $: parsedYaml = parse(yaml) as V1DashboardSpec;

  $: dashboard = $query.data?.dashboard?.spec;

  $: columns = dashboard?.grid?.columns ?? 10;
  $: gap = dashboard?.grid?.gap ?? 1;
  $: charts = dashboard?.components ?? ([] as V1DashboardComponent[]);

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
    if (
      isDuplicateName(
        e.currentTarget.value,
        customDashboardName,
        $allNamesQuery?.data ?? [],
      )
    ) {
      notifications.send({
        message: `Name ${e.currentTarget.value} is already in use`,
      });
      e.currentTarget.value = customDashboardName; // resets the input
      return;
    }

    try {
      const toName = e.currentTarget.value;
      const entityType = EntityType.Dashboard;
      await renameFileArtifact(
        $runtime.instanceId,
        customDashboardName,
        toName,
        entityType,
      );
      await goto(getRouteFromName(toName, entityType), {
        replaceState: true,
      });
    } catch (err) {
      console.error(err.response.data.message);
    }
  };

  function handleStartResize(e: MouseEvent) {
    startingWidth = editorWidth;
    startingX = e.clientX;

    window.addEventListener("mousemove", handleResize);
    window.addEventListener("mouseup", () => {
      window.removeEventListener("mousemove", handleResize);
    });
  }

  function handleResize(e: MouseEvent) {
    currentX = e.clientX;
    editorWidth = Math.max(
      300,
      Math.min(startingWidth + (currentX - startingX), 600),
    );
  }

  async function updateChart(e: CustomEvent<string>) {
    const content = e.detail;
    if (!content) return;
    try {
      await $updateFile.mutateAsync({
        instanceId: $runtime.instanceId,
        path: getFileAPIPathFromNameAndType(
          customDashboardName,
          EntityType.Dashboard,
        ),
        data: {
          blob: content,
        },
      });
      yaml = content;
      dashboard = parse(content) as V1DashboardSpec;
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
    const components = [...(parsedYaml?.components ?? [])];

    // if (e.detail.change === "dimension") {
    components[e.detail.index].width = e.detail.dimensions[0];
    components[e.detail.index].height = e.detail.dimensions[1];
    // } else {
    components[e.detail.index].x = e.detail.position[0];
    components[e.detail.index].y = e.detail.position[1];
    // }

    const stringified = stringify({ ...parsedYaml, components });

    await updateChart(new CustomEvent("update", { detail: stringified }));
  }

  const viewOptions = [
    { view: "code", icon: Code },
    { view: "split", icon: Split },
    { view: "viz", icon: Viz },
  ];

  let selectedView = "split";

  let snap = false;

  async function addChart(e: CustomEvent<{ chartName: string }>) {
    const components = [...(parsedYaml?.components ?? [])];
    components.push({
      chart: e.detail.chartName,
      height: 4,
      width: 4,
      x: 0,
      y: 0,
    });

    const stringified = stringify({ ...parsedYaml, components });

    await updateChart(new CustomEvent("update", { detail: stringified }));
  }
</script>

<svelte:head>
  <title>Rill Developer | {customDashboardName}</title>
</svelte:head>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    slot="header"
    titleInput={customDashboardName}
    showInspectorToggle={false}
    {onChangeCallback}
  >
    <div slot="workspace-controls" class="flex gap-x-4 items-center">
      <div
        class="flex border-primary-300 rounded-sm border w-fit h-7 items-center justify-center"
      >
        {#each viewOptions as { view, icon: Icon }}
          <input
            type="radio"
            id={view}
            name="view"
            value={view}
            class="hidden"
            checked={view === "code"}
            bind:group={selectedView}
          />
          <label
            for={view}
            class="cursor-pointer w-7 aspect-square flex items-center justify-center border-r h-full border-primary-300 last:border-r-0"
            class:bg-primary-200={selectedView === view}
          >
            <Icon size="15px" class="fill-primary-600" />
          </label>
        {/each}
      </div>

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
        class:width={selectedView === "split"}
        style:--width="{editorWidth}px"
      >
        <button class="resizer" on:mousedown={handleStartResize} />
        <CustomDashboardEditor {yaml} on:update={updateChart} />
      </div>
    {/if}

    {#if selectedView == "viz" || selectedView == "split"}
      <CustomDashboard
        {snap}
        {gap}
        {charts}
        {columns}
        {showGrid}
        on:update={handlePreviewUpdate}
      />
    {/if}
  </div>
</WorkspaceContainer>

<style lang="postcss">
  .resizer {
    @apply h-full w-2 absolute right-0 z-10 cursor-col-resize;
  }

  .width {
    width: var(--width);
  }
</style>
