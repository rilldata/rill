<script lang="ts">
  import { goto } from "$app/navigation";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/MetricsInspector.svelte";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/editor/MetricsEditor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    useIsModelingSupportedForConnectorOLAP as useIsModelingSupportedForConnector,
    useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver,
  } from "../connectors/selectors";
  import PreviewButton from "../explores/PreviewButton.svelte";
  import GoToDashboardButton from "../metrics-views/GoToDashboardButton.svelte";
  import { mapParseErrorsToLines } from "../metrics-views/errors";
  import VisualMetrics from "./VisualMetrics.svelte";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    remoteContent,
    fileName,
  } = fileArtifact);

  $: workspace = workspaces.get(filePath);

  $: metricsViewName = $resourceName?.name ?? getNameFromFile(filePath);

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: ({ data: resource } = $resourceQuery);

  $: isOldMetricsView = !$remoteContent?.includes("version: 1");
  $: connector = resource?.metricsView?.state?.validSpec?.connector ?? "";
  $: database = resource?.metricsView?.state?.validSpec?.database ?? "";
  $: databaseSchema =
    resource?.metricsView?.state?.validSpec?.databaseSchema ?? "";
  $: table = resource?.metricsView?.state?.validSpec?.table ?? "";

  $: isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver(instanceId);
  $: isModelingSupportedForConnector = useIsModelingSupportedForConnector(
    instanceId,
    connector,
  );
  $: isModelingSupported = connector
    ? $isModelingSupportedForConnector.data
    : $isModelingSupportedForDefaultOlapDriver.data;

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      instanceId,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }

  $: selectedView = workspace.view;

  $: errors = mapParseErrorsToLines(allErrors, $remoteContent ?? "");
</script>

<WorkspaceContainer inspector={$selectedView === "code" && isModelingSupported}>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={ResourceKind.MetricsView}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    showInspectorToggle={$selectedView === "code" && isModelingSupported}
    slot="header"
    codeToggle
    titleInput={fileName}
  >
    <div class="flex gap-x-2" slot="cta">
      {#if isOldMetricsView}
        <PreviewButton
          href="/explore/{metricsViewName}"
          disabled={errors.length > 0}
        />
      {:else}
        <GoToDashboardButton {resource} />
      {/if}
    </div>
  </WorkspaceHeader>

  <svelte:fragment slot="body">
    {#if $selectedView === "code"}
      <MetricsEditor
        bind:autoSave={$autoSave}
        {fileArtifact}
        {filePath}
        {errors}
        {metricsViewName}
      />
    {:else}
      {#key fileArtifact}
        <VisualMetrics
          {errors}
          {fileArtifact}
          switchView={() => {
            $selectedView = "code";
          }}
        />
      {/key}
    {/if}
  </svelte:fragment>

  <MetricsInspector
    {filePath}
    {connector}
    {database}
    {databaseSchema}
    {table}
    slot="inspector"
  />
</WorkspaceContainer>
