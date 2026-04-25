<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/MetricsInspector.svelte";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/editor/MetricsEditor.svelte";
  import { editorMode } from "@rilldata/web-common/layout/editor-mode-store";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    useIsModelingSupportedForConnectorOLAP as useIsModelingSupportedForConnector,
    useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver,
  } from "../connectors/selectors";
  import Dashboard from "../dashboards/workspace/Dashboard.svelte";
  import DashboardStateManager from "../dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "../dashboards/state-managers/StateManagersProvider.svelte";
  import GoToDashboardButton from "../metrics-views/GoToDashboardButton.svelte";
  import ReconcileWarningPanel from "../entity-management/ReconcileWarningPanel.svelte";
  import VisualMetrics from "./VisualMetrics.svelte";

  export let fileArtifact: FileArtifact;
  export let inPreviewMode = false;

  const runtimeClient = useRuntimeClient();

  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    remoteContent,
    fileName,
  } = fileArtifact);

  // `?view=dashboard` is set by the visual nav when the user clicks a
  // synthetic explore (one emitted from this metrics view's inline `explore:`
  // block or v0 defaults). In that mode we render the dashboard preview
  // inside the workspace chrome instead of the metrics editor.
  $: dashboardView = $page.url.searchParams.get("view") === "dashboard";

  $: selectedView = $editorMode === "visual" ? "viz" : "code";

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

  // Parse error for the editor gutter and banner
  $: parseErrorQuery = fileArtifact.getParseError(queryClient);
  $: parseError = $parseErrorQuery;

  $: reconcileError = resource?.meta?.reconcileError;

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

<WorkspaceContainer
  inspector={!dashboardView && selectedView === "code" && isModelingSupported}
>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={dashboardView
      ? ResourceKind.Explore
      : ResourceKind.MetricsView}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    showInspectorToggle={!dashboardView &&
      selectedView === "code" &&
      isModelingSupported}
    slot="header"
    titleInput={fileName}
  >
    <div class="flex gap-x-2" slot="cta">
      {#if !inPreviewMode && !isOldMetricsView && !dashboardView}
        <GoToDashboardButton {resource} />
      {/if}
    </div>
  </WorkspaceHeader>

  <svelte:fragment slot="body">
    <div class="flex flex-col h-full">
      <div class="flex-1 overflow-hidden">
        {#if dashboardView}
          {#key metricsViewName}
            <StateManagersProvider
              metricsViewName={metricsViewName ?? ""}
              exploreName={metricsViewName ?? ""}
            >
              <DashboardStateManager exploreName={metricsViewName ?? ""}>
                <Dashboard
                  metricsViewName={metricsViewName ?? ""}
                  exploreName={metricsViewName ?? ""}
                />
              </DashboardStateManager>
            </StateManagersProvider>
          {/key}
        {:else}
          <WorkspaceEditorContainer
            {resource}
            {parseError}
            remoteContent={$remoteContent}
            {filePath}
          >
            {#if selectedView === "code"}
              <MetricsEditor
                bind:autoSave={$autoSave}
                {fileArtifact}
                {filePath}
                {parseError}
                {metricsViewName}
              />
            {:else}
              {#key fileArtifact}
                <VisualMetrics
                  {fileArtifact}
                  switchView={() => editorMode.set("code")}
                />
              {/key}
            {/if}
          </WorkspaceEditorContainer>
        {/if}
      </div>
      <ReconcileWarningPanel {fileArtifact} />
    </div>
  </svelte:fragment>

  <svelte:fragment slot="inspector">
    {#if !dashboardView}
      <MetricsInspector
        {filePath}
        {connector}
        {database}
        {databaseSchema}
        {table}
      />
    {/if}
  </svelte:fragment>
</WorkspaceContainer>
