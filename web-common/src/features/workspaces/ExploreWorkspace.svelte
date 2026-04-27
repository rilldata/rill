<script lang="ts">
  import { goto } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { withEditorPrefix } from "@rilldata/web-common/layout/navigation/editor-routing";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    resourceIsLoading,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import ExploreEditor from "@rilldata/web-common/features/explores/ExploreEditor.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ExplainAndFixErrorButton from "@rilldata/web-common/features/chat/ExplainAndFixErrorButton.svelte";
  import ReconcileWarningPanel from "../entity-management/ReconcileWarningPanel.svelte";
  import Spinner from "../entity-management/Spinner.svelte";
  import PreviewButton from "../explores/PreviewButton.svelte";
  import VisualExploreEditing from "./VisualExploreEditing.svelte";
  import StateManagersProvider from "../dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateManager from "../dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import Dashboard from "../dashboards/workspace/Dashboard.svelte";

  export let fileArtifact: FileArtifact;
  export let hideCodeToggle = false;
  export let inPreviewMode = false;

  const runtimeClient = useRuntimeClient();

  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    fileName,
    remoteContent,
  } = fileArtifact);

  $: exploreName = $resourceName?.name ?? getNameFromFile(filePath);

  $: query = createRuntimeServiceGetExplore(runtimeClient, {
    name: exploreName,
  });

  $: ({ data: resources } = $query);

  $: exploreResource = resources?.explore;
  $: metricsViewResource = resources?.metricsView;

  $: resourceIsReconciling = resourceIsLoading(exploreResource);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;

  $: selectedView = $selectedViewStore ?? "code";

  $: metricsViewName = metricsViewResource?.meta?.name?.name;

  // Parse error for the editor gutter and banner
  $: parseErrorQuery = fileArtifact.getParseError(queryClient);
  $: parseError = $parseErrorQuery;

  $: reconcileError = (exploreResource ?? metricsViewResource)?.meta
    ?.reconcileError;
  $: rootCauseQuery = createRootCauseErrorQuery(
    runtimeClient,
    exploreResource ?? metricsViewResource,
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

{#key metricsViewName + exploreName}
  <StateManagersProvider
    {metricsViewName}
    {exploreName}
    visualEditing
    let:ready
  >
    <WorkspaceContainer>
      <WorkspaceHeader
        resource={exploreResource}
        hasUnsavedChanges={$hasUnsavedChanges}
        onTitleChange={onChangeCallback}
        slot="header"
        titleInput={fileName}
        {filePath}
        codeToggle={!hideCodeToggle}
        resourceKind={ResourceKind.Explore}
      >
        <div class="flex gap-x-2" slot="cta">
          {#if !inPreviewMode}
            <PreviewButton
              href={withEditorPrefix(`/explore/${exploreName}`)}
              disabled={!!parseError ||
                !!reconcileError ||
                resourceIsReconciling}
              reconciling={resourceIsReconciling}
            />
          {/if}
        </div>
      </WorkspaceHeader>

      <svelte:fragment slot="body">
        <div class="flex flex-col h-full">
          <div class="flex-1 min-h-0">
            <WorkspaceEditorContainer
              resource={exploreResource ?? metricsViewResource}
              {parseError}
              remoteContent={$remoteContent}
              {filePath}
            >
              {#if selectedView === "code"}
                <ExploreEditor
                  bind:autoSave={$autoSave}
                  {exploreName}
                  {fileArtifact}
                  {parseError}
                />
              {:else if selectedView === "viz"}
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
                {:else if exploreName && metricsViewName}
                  <DashboardStateManager {exploreName}>
                    <Dashboard {metricsViewName} {exploreName} />
                  </DashboardStateManager>
                {:else}
                  <Spinner status={1} size="48px" />
                {/if}
              {/if}
            </WorkspaceEditorContainer>
          </div>
          <ReconcileWarningPanel {fileArtifact} />
        </div>
      </svelte:fragment>

      <svelte:fragment slot="inspector">
        {#if ready}
          <VisualExploreEditing
            {metricsViewName}
            {exploreName}
            autoSave={selectedView === "viz" || $autoSave}
            exploreResource={exploreResource?.explore}
            {fileArtifact}
            viewingDashboard={selectedView === "viz"}
            switchView={() => selectedViewStore.set("code")}
          />
        {/if}
      </svelte:fragment>
    </WorkspaceContainer>
  </StateManagersProvider>
{/key}
