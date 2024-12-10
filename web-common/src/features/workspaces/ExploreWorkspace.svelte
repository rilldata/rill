<script lang="ts">
  import { goto } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    resourceIsLoading,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import ExploreEditor from "@rilldata/web-common/features/explores/ExploreEditor.svelte";
  import ViewSelector from "@rilldata/web-common/features/visual-editing/ViewSelector.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import DashboardWithProviders from "../dashboards/workspace/DashboardWithProviders.svelte";
  import Spinner from "../entity-management/Spinner.svelte";
  import PreviewButton from "../explores/PreviewButton.svelte";
  import MetricsEditorContainer from "../metrics-views/editor/MetricsEditorContainer.svelte";
  import { mapParseErrorsToLines } from "../metrics-views/errors";
  import VisualExploreEditing from "./VisualExploreEditing.svelte";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    fileName,
    getAllErrors,
    remoteContent,
  } = fileArtifact);

  $: exploreName = $resourceName?.name ?? getNameFromFile(filePath);

  $: query = createRuntimeServiceGetExplore(instanceId, { name: exploreName });

  $: ({ data: resources } = $query);

  $: initLocalUserPreferenceStore(exploreName);

  $: exploreResource = resources?.explore;
  $: metricsViewResource = resources?.metricsView;

  $: allErrorsQuery = getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceIsReconciling = resourceIsLoading(exploreResource);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;

  $: selectedView = $selectedViewStore ?? "code";

  $: metricsViewName = metricsViewResource?.meta?.name?.name;

  $: lineBasedRuntimeErrors = mapParseErrorsToLines(
    allErrors,
    $remoteContent ?? "",
  );

  $: mainError = lineBasedRuntimeErrors?.at(0);

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

<WorkspaceContainer>
  <WorkspaceHeader
    resource={exploreResource}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    slot="header"
    titleInput={fileName}
    {filePath}
    resourceKind={ResourceKind.Explore}
  >
    <div class="flex gap-x-2" slot="cta">
      <PreviewButton
        href="/explore/{exploreName}"
        disabled={allErrors.length > 0 || resourceIsReconciling}
        reconciling={resourceIsReconciling}
      />

      <ViewSelector allowSplit={false} bind:selectedView={$selectedViewStore} />
    </div>
  </WorkspaceHeader>

  <MetricsEditorContainer
    slot="body"
    error={mainError}
    showError={!!$remoteContent && selectedView === "code"}
  >
    {#if selectedView === "code"}
      <ExploreEditor
        bind:autoSave={$autoSave}
        {exploreName}
        {fileArtifact}
        {lineBasedRuntimeErrors}
        forceLocalUpdates
      />
    {:else if selectedView === "viz"}
      {#if mainError}
        <ErrorPage
          body={mainError.message}
          fatal
          header="Unable to load dashboard preview"
          statusCode={404}
        />
      {:else if exploreName && metricsViewName}
        <DashboardWithProviders {exploreName} {metricsViewName} />
      {:else}
        <Spinner status={1} size="48px" />
      {/if}
    {/if}
  </MetricsEditorContainer>

  <VisualExploreEditing
    autoSave={selectedView === "viz" || $autoSave}
    slot="inspector"
    exploreResource={exploreResource?.explore}
    {metricsViewName}
    {exploreName}
    {fileArtifact}
    viewingDashboard={selectedView === "viz"}
    switchView={() => selectedViewStore.set("code")}
  />
</WorkspaceContainer>
