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

  import Toggle from "@rilldata/web-common/features/custom-dashboards/Toggle.svelte";

  const updateFile = createRuntimeServicePutFile();

  let editing = true;
  let showGrid = false;
  let singlePanel = false;
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
      change: "dimension" | "position";
      vector: Vector;
    }>,
  ) {
    const components = [...(parsedYaml?.components ?? [])];

    if (e.detail.change === "dimension") {
      components[e.detail.index].width = e.detail.vector[0];
      components[e.detail.index].height = e.detail.vector[1];
    } else {
      components[e.detail.index].x = e.detail.vector[0];
      components[e.detail.index].y = e.detail.vector[1];
    }

    const stringified = stringify({ ...parsedYaml, components });

    await updateChart(new CustomEvent("update", { detail: stringified }));
  }
</script>

<svelte:head>
  <title>Rill Developer | {customDashboardName}</title>
</svelte:head>

<WorkspaceContainer assetID={customDashboardName} inspector={false}>
  <WorkspaceHeader
    slot="header"
    titleInput={customDashboardName}
    showInspectorToggle={false}
    {onChangeCallback}
  >
    <div slot="workspace-controls" class="flex gap-x-4">
      {#if !singlePanel || !editing}
        <Toggle bind:bool={showGrid} text={["Show grid", "Hide grid"]} />
      {/if}

      {#if singlePanel}
        <Toggle bind:bool={editing} text={["Preview", "Edit"]} />
      {/if}

      <Toggle bind:bool={singlePanel} text={["Two panel", "One panel"]} />
    </div>
  </WorkspaceHeader>

  <div class="flex w-full h-full flex-row overflow-hidden" slot="body">
    {#if !singlePanel || (singlePanel && editing)}
      <div
        class="relative h-full flex-shrink-0 border-r border-gray-400 w-full"
        class:width={!singlePanel}
        style:--width="{editorWidth}px"
      >
        <button class="resizer" on:mousedown={handleStartResize} />
        <CustomDashboardEditor {yaml} on:update={updateChart} />
      </div>
    {/if}

    {#if !singlePanel || (singlePanel && !editing)}
      <CustomDashboard
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
