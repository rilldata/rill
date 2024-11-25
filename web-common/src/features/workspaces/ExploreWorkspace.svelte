<script lang="ts">
  import { goto } from "$app/navigation";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    resourceIsLoading,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import ExploreEditor from "@rilldata/web-common/features/explores/ExploreEditor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import PreviewButton from "../explores/PreviewButton.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import ViewSelector from "@rilldata/web-common/features/visual-editing/ViewSelector.svelte";
  import VisualExploreEditing from "./VisualExploreEditing.svelte";
  import MetricsEditorContainer from "../metrics-views/editor/MetricsEditorContainer.svelte";
  import { mapParseErrorsToLines } from "../metrics-views/errors";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import DashboardPage from "/Users/burg/OKAY/rill/web-local/src/routes/(viz)/explore/[name]/+page.svelte";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import Spinner from "../entity-management/Spinner.svelte";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);
  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    fileName,
    getResource,
    getAllErrors,
    remoteContent,
  } = fileArtifact);

  $: exploreName = $resourceName?.name ?? getNameFromFile(filePath);

  $: query = createRuntimeServiceGetExplore(instanceId, { name: exploreName });

  $: ({ data } = $query);

  $: initLocalUserPreferenceStore(exploreName);

  $: resourceQuery = getResource(queryClient, instanceId);

  $: ({ data: resource } = $resourceQuery);

  $: allErrorsQuery = getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceIsReconciling = resourceIsLoading(resource);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;

  $: selectedView = $selectedViewStore ?? "code";

  $: exploreResource = resource?.explore;

  $: metricsViewName = resource?.meta?.refs?.find(
    (ref) => ref.kind === ResourceKind.MetricsView,
  )?.name;

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
    {resource}
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
      />
    {:else if selectedView === "viz"}
      {#if mainError}
        <ErrorPage
          body={mainError.message}
          fatal
          header="Unable to load dashboard preview"
          statusCode={404}
        />
      {:else if data?.explore && data.metricsView}
        <DashboardPage {data} />
      {:else}
        <Spinner status={1} size="48px" />
      {/if}
    {/if}
  </MetricsEditorContainer>

  <VisualExploreEditing
    slot="inspector"
    {exploreResource}
    {metricsViewName}
    {exploreName}
    {fileArtifact}
    viewingDashboard={selectedView === "viz"}
    switchView={() => selectedViewStore.set("code")}
  />
</WorkspaceContainer>
