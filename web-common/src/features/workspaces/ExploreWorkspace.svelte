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
  import DashboardWithProviders from "../dashboards/workspace/DashboardWithProviders.svelte";

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
  } = fileArtifact);

  $: exploreName = $resourceName?.name ?? getNameFromFile(filePath);

  $: initLocalUserPreferenceStore(exploreName);

  $: resourceQuery = getResource(queryClient, instanceId);

  $: ({ data } = $resourceQuery);

  $: allErrorsQuery = getAllErrors(queryClient, instanceId);
  $: allErrors = $allErrorsQuery;
  $: resourceIsReconciling = resourceIsLoading(data);

  $: workspace = workspaces.get(filePath);
  $: selectedView = workspace.view;

  $: exploreResource = data?.explore;

  $: metricsViewName = data?.meta?.refs?.find(
    (ref) => ref.kind === ResourceKind.MetricsView,
  )?.name;

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

      <ViewSelector allowSplit={false} bind:selectedView={$selectedView} />
    </div>
  </WorkspaceHeader>

  <svelte:fragment slot="body">
    {#if $selectedView === "code"}
      <ExploreEditor
        forceLocalUpdates
        bind:autoSave={$autoSave}
        {exploreName}
        {fileArtifact}
        {allErrors}
      />
    {:else if $selectedView === "viz"}
      {#key fileArtifact}
        <div
          class="size-full border overflow-hidden rounded-[2px] bg-background flex flex-col items-center justify-center"
        >
          {#if metricsViewName && exploreName}
            <DashboardWithProviders {exploreName} {metricsViewName} />
          {/if}
        </div>
      {/key}
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="inspector">
    {#if exploreResource && metricsViewName}
      <VisualExploreEditing
        {exploreResource}
        {metricsViewName}
        {exploreName}
        {fileArtifact}
        viewingDashboard={$selectedView === "viz"}
        switchView={() => selectedView.set("code")}
      />
    {/if}
  </svelte:fragment>
</WorkspaceContainer>
