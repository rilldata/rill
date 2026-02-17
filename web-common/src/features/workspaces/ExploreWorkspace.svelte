<script lang="ts">
  import { goto } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { createRootCauseErrorQuery } from "@rilldata/web-common/features/entity-management/error-utils";
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
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Spinner from "../entity-management/Spinner.svelte";
  import PreviewButton from "../explores/PreviewButton.svelte";
  import VisualExploreEditing from "./VisualExploreEditing.svelte";
  import StateManagersProvider from "../dashboards/state-managers/StateManagersProvider.svelte";
  import DashboardStateManager from "../dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import Dashboard from "../dashboards/workspace/Dashboard.svelte";
  import SaveDefaultsButton from "../dashboards/workspace/SaveDefaultsButton.svelte";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    fileName,
    remoteContent,
    saveState: { saving },
  } = fileArtifact);

  $: exploreName = $resourceName?.name ?? getNameFromFile(filePath);

  $: query = createRuntimeServiceGetExplore(instanceId, { name: exploreName });

  $: ({ data: resources } = $query);

  $: exploreResource = resources?.explore;
  $: metricsViewResource = resources?.metricsView;

  $: resourceIsReconciling = resourceIsLoading(exploreResource);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;

  $: selectedView = $selectedViewStore ?? "code";

  $: metricsViewName = metricsViewResource?.meta?.name?.name;

  // Parse error for the editor gutter and banner
  $: parseErrorQuery = fileArtifact.getParseError(queryClient, instanceId);
  $: parseError = $parseErrorQuery;

  // Reconcile error resolved to root cause for the banner
  $: reconcileError = (exploreResource ?? metricsViewResource)?.meta
    ?.reconcileError;
  $: rootCauseQuery = createRootCauseErrorQuery(
    instanceId,
    exploreResource ?? metricsViewResource,
    reconcileError,
  );
  $: rootCauseReconcileError = reconcileError
    ? ($rootCauseQuery?.data ?? reconcileError)
    : undefined;

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      instanceId,
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
        codeToggle
        resourceKind={ResourceKind.Explore}
      >
        <div class="flex gap-x-2" slot="cta">
          {#if ready && selectedView === "viz"}
            <SaveDefaultsButton {fileArtifact} saving={$saving} />
          {/if}
          <PreviewButton
            href="/explore/{exploreName}"
            disabled={!!parseError || !!reconcileError || resourceIsReconciling}
            reconciling={resourceIsReconciling}
          />
        </div>
      </WorkspaceHeader>

      <WorkspaceEditorContainer
        slot="body"
        error={parseError?.message ?? rootCauseReconcileError}
        showError={!!$remoteContent && selectedView === "code"}
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
            />
          {:else if exploreName && metricsViewName}
            <DashboardStateManager {exploreName}>
              <Dashboard {metricsViewName} {exploreName} />
            </DashboardStateManager>
          {:else}
            <Spinner status={1} size="48px" />
          {/if}
        {/if}
      </WorkspaceEditorContainer>

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
