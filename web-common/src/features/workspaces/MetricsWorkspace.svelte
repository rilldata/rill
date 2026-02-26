<script lang="ts">
  import { goto } from "$app/navigation";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/MetricsInspector.svelte";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/editor/MetricsEditor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    useIsModelingSupportedForConnectorOLAP as useIsModelingSupportedForConnector,
    useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver,
  } from "../connectors/selectors";
  import PreviewButton from "../explores/PreviewButton.svelte";
  import GoToDashboardButton from "../metrics-views/GoToDashboardButton.svelte";
  import VisualMetrics from "./VisualMetrics.svelte";

  export let fileArtifact: FileArtifact;

  const runtimeClient = useRuntimeClient();

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

  $: resourceQuery = fileArtifact.getResource(queryClient);
  $: ({ data: resource } = $resourceQuery);

  $: isOldMetricsView = !$remoteContent?.includes("version: 1");
  $: connector = resource?.metricsView?.state?.validSpec?.connector ?? "";
  $: database = resource?.metricsView?.state?.validSpec?.database ?? "";
  $: databaseSchema =
    resource?.metricsView?.state?.validSpec?.databaseSchema ?? "";
  $: table = resource?.metricsView?.state?.validSpec?.table ?? "";

  $: isModelingSupportedForDefaultOlapDriver =
    useIsModelingSupportedForDefaultOlapDriver(runtimeClient);
  $: isModelingSupportedForConnector = useIsModelingSupportedForConnector(
    runtimeClient,
    connector,
  );
  $: isModelingSupported = connector
    ? $isModelingSupportedForConnector.data
    : $isModelingSupportedForDefaultOlapDriver.data;

  $: selectedView = workspace.view;

  // Parse error for the editor gutter and banner
  $: parseErrorQuery = fileArtifact.getParseError(queryClient);
  $: parseError = $parseErrorQuery;

  // Reconcile error resolved to root cause for the banner
  $: reconcileError = resource?.meta?.reconcileError;
  $: rootCauseQuery = createRootCauseErrorQuery(
    runtimeClient,
    resource,
    reconcileError,
  );
  $: rootCauseReconcileError = reconcileError
    ? ($rootCauseQuery?.data ?? reconcileError)
    : undefined;

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      runtimeClient,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }
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
          disabled={!!parseError || !!reconcileError}
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
        {rootCauseReconcileError}
        {fileArtifact}
        {filePath}
        {parseError}
        {metricsViewName}
      />
    {:else}
      {#key fileArtifact}
        <VisualMetrics
          {parseError}
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
