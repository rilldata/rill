<script lang="ts">
  import { goto } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { withEditorPrefix } from "@rilldata/web-common/layout/navigation/editor-routing";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils";
  import {
    resourceIsLoading,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/actions/ui-actions.ts";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import MetricsInspector from "@rilldata/web-common/features/metrics-views/MetricsInspector.svelte";
  import MetricsEditor from "@rilldata/web-common/features/metrics-views/editor/MetricsEditor.svelte";
  import {
    parseInlineExploreState,
    type MetricsWorkspaceView,
  } from "@rilldata/web-common/features/metrics-views/inline-explore";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ExplainAndFixErrorButton from "@rilldata/web-common/features/chat/ExplainAndFixErrorButton.svelte";
  import { Code2Icon } from "lucide-svelte";
  import {
    useIsModelingSupportedForConnectorOLAP as useIsModelingSupportedForConnector,
    useIsModelingSupportedForDefaultOlapDriverOLAP as useIsModelingSupportedForDefaultOlapDriver,
  } from "../connectors/selectors";
  import DashboardStateManager from "../dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "../dashboards/state-managers/StateManagersProvider.svelte";
  import Dashboard from "../dashboards/workspace/Dashboard.svelte";
  import ReconcileWarningPanel from "../entity-management/ReconcileWarningPanel.svelte";
  import Spinner from "../entity-management/Spinner.svelte";
  import PreviewButton from "../explores/PreviewButton.svelte";
  import GoToDashboardButton from "../metrics-views/GoToDashboardButton.svelte";
  import type { ViewOption } from "../visual-editing/CodeToggle.svelte";
  import VisualExploreEditing from "./VisualExploreEditing.svelte";
  import VisualMetrics from "./VisualMetrics.svelte";

  export let fileArtifact: FileArtifact;
  export let hideCodeToggle = false;
  export let inPreviewMode = false;

  const runtimeClient = useRuntimeClient();

  const threeWayViews: ViewOption[] = [
    { view: "code", icon: Code2Icon, label: "Code view" },
    {
      view: "viz",
      icon: resourceIconMapping[ResourceKind.MetricsView],
      label: "Metrics view editor",
    },
    {
      view: "explore",
      icon: resourceIconMapping[ResourceKind.Explore],
      label: "Explore dashboard editor",
    },
  ];

  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    editorContent,
    remoteContent,
    fileName,
  } = fileArtifact);

  $: workspace = workspaces.get<MetricsWorkspaceView>(filePath);

  $: metricsViewName = $resourceName?.name ?? getNameFromFile(filePath);

  $: resourceQuery = fileArtifact.getResource(queryClient);
  $: ({ data: resource } = $resourceQuery);

  $: ({
    isLatestVersion,
    exploreEnabled,
    exploreName: exploreNameOverride,
  } = parseInlineExploreState($editorContent ?? $remoteContent));
  $: inlineExploreName = exploreNameOverride ?? metricsViewName;

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

  // Fall back to the visual metrics editor if the inline explore is disabled while
  // the explore view is selected (e.g. the block was deleted, or stale localStorage).
  $: effectiveView =
    $selectedView === "explore" && !exploreEnabled ? "viz" : $selectedView;

  // Parse error for the editor gutter and banner
  $: parseErrorQuery = fileArtifact.getParseError(queryClient);
  $: parseError = $parseErrorQuery;

  $: reconcileError = resource?.meta?.reconcileError;

  // The explore resource emitted by the inline `explore:` block
  $: exploreQuery = createRuntimeServiceGetExplore(
    runtimeClient,
    { name: inlineExploreName },
    { query: { enabled: exploreEnabled } },
  );
  $: exploreResource = exploreEnabled ? $exploreQuery.data?.explore : undefined;
  $: exploreIsReconciling = resourceIsLoading(exploreResource);
  $: exploreReconcileError = exploreResource?.meta?.reconcileError;
  $: rootCauseQuery = createRootCauseErrorQuery(
    runtimeClient,
    exploreResource,
    exploreReconcileError,
  );
  $: rootCauseReconcileError = exploreReconcileError
    ? ($rootCauseQuery?.data ?? exploreReconcileError)
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

{#snippet workspaceContent(ready: boolean)}
  <WorkspaceContainer
    inspector={(effectiveView === "code" && isModelingSupported) ||
      effectiveView === "explore"}
  >
    <WorkspaceHeader
      {filePath}
      {resource}
      resourceKind={ResourceKind.MetricsView}
      hasUnsavedChanges={$hasUnsavedChanges}
      onTitleChange={onChangeCallback}
      showInspectorToggle={(effectiveView === "code" && isModelingSupported) ||
        effectiveView === "explore"}
      slot="header"
      codeToggle={!hideCodeToggle}
      codeToggleViews={exploreEnabled ? threeWayViews : undefined}
      titleInput={fileName}
    >
      {#snippet cta()}
        <div class="flex gap-x-2">
          {#if !inPreviewMode}
            {#if !isLatestVersion}
              <PreviewButton
                href={withEditorPrefix(`/explore/${metricsViewName}`)}
                disabled={!!parseError || !!reconcileError}
              />
            {:else if exploreEnabled}
              <GoToDashboardButton {resource} hasInlineExplore />
              <PreviewButton
                label="Preview Explore"
                href={withEditorPrefix(`/explore/${inlineExploreName}`)}
                disabled={!!parseError ||
                  !!reconcileError ||
                  !!exploreReconcileError ||
                  exploreIsReconciling}
                reconciling={exploreIsReconciling}
              />
            {:else}
              <GoToDashboardButton {resource} />
            {/if}
          {/if}
        </div>
      {/snippet}
    </WorkspaceHeader>

    <svelte:fragment slot="body">
      <div class="flex flex-col h-full">
        <div class="flex-1 overflow-hidden">
          <WorkspaceEditorContainer
            {resource}
            {parseError}
            remoteContent={$remoteContent}
            {filePath}
            showError={effectiveView !== "explore"}
          >
            {#if effectiveView === "code"}
              <MetricsEditor
                bind:autoSave={$autoSave}
                {fileArtifact}
                {filePath}
                {parseError}
                {metricsViewName}
              />
            {:else if effectiveView === "explore"}
              {#if parseError || rootCauseReconcileError}
                <ErrorPage
                  body={parseError?.message ?? rootCauseReconcileError ?? ""}
                  fatal
                  header="Unable to load dashboard preview"
                  statusCode={404}
                >
                  <svelte:fragment slot="cta">
                    <ExplainAndFixErrorButton {filePath} variant="cta" />
                  </svelte:fragment>
                </ErrorPage>
              {:else if exploreResource && metricsViewName}
                <DashboardStateManager exploreName={inlineExploreName}>
                  <Dashboard
                    {metricsViewName}
                    exploreName={inlineExploreName}
                  />
                </DashboardStateManager>
              {:else}
                <Spinner status={1} size="48px" />
              {/if}
            {:else}
              {#key fileArtifact}
                <VisualMetrics
                  {fileArtifact}
                  switchView={() => {
                    $selectedView = "code";
                  }}
                />
              {/key}
            {/if}
          </WorkspaceEditorContainer>
        </div>
        <ReconcileWarningPanel {fileArtifact} />
      </div>
    </svelte:fragment>

    <svelte:fragment slot="inspector">
      {#if effectiveView === "explore"}
        {#if ready}
          <VisualExploreEditing
            {metricsViewName}
            exploreName={inlineExploreName}
            keyPath={["explore"]}
            autoSave={effectiveView === "explore" || $autoSave}
            exploreResource={exploreResource?.explore}
            {fileArtifact}
            viewingDashboard={effectiveView === "explore"}
            switchView={() => {
              $selectedView = "code";
            }}
          />
        {/if}
      {:else}
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
{/snippet}

{#if exploreEnabled && metricsViewName}
  {#key metricsViewName + inlineExploreName}
    <StateManagersProvider
      {metricsViewName}
      exploreName={inlineExploreName}
      visualEditing
      let:ready
    >
      {@render workspaceContent(ready)}
    </StateManagersProvider>
  {/key}
{:else}
  {@render workspaceContent(false)}
{/if}
