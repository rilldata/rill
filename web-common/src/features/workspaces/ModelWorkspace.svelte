<script lang="ts">
  import { goto } from "$app/navigation";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import WorkspaceInspector from "@rilldata/web-common/features/models/inspector/WorkspaceInspector.svelte";
  import ModelEditor from "@rilldata/web-common/features/models/workspace/ModelEditor.svelte";
  import ModelWorkspaceCtAs from "@rilldata/web-common/features/models/workspace/ModelWorkspaceCTAs.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import WorkspaceTableContainer from "@rilldata/web-common/layout/workspace/WorkspaceTableContainer.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { fade, slide } from "svelte/transition";
  import ReconcilingSpinner from "../entity-management/ReconcilingSpinner.svelte";
  import { getUserFriendlyError } from "../models/error-utils";

  export let fileArtifact: FileArtifact;

  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    fileName,
    remoteContent,
  } = fileArtifact);

  $: assetName = getNameFromFile(filePath);

  $: workspace = workspaces.get(filePath);
  $: tableVisible = workspace.table.visible;

  $: ({ instanceId } = $runtime);

  $: allErrorsStore = fileArtifact.getAllErrors(queryClient, instanceId);
  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  $: allErrors = $allErrorsStore;

  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data;
  $: model = $resourceQuery.data?.model;
  $: connector = (model as V1Model)?.spec?.outputConnector as string;
  const database = ""; // models use the default database
  const databaseSchema = ""; // models use the default databaseSchema
  $: tableName = (model as V1Model)?.state?.resultTable as string;

  $: refreshedOn = model?.state?.refreshedOn;
  $: isResourceReconciling = resourceIsLoading($resourceQuery.data);

  async function save() {
    httpRequestQueue.removeByName(assetName);
    await queryClient.cancelQueries({
      predicate: (query) => isProfilingQuery(query, assetName),
    });
  }

  async function handleNameChange(newTitle: string) {
    const newRoute = await handleEntityRename(
      instanceId,
      newTitle,
      filePath,
      fileName,
    );

    if (newRoute) await goto(newRoute);
  }

  function formatRefreshedOn(refreshedOn: string) {
    const date = new Date(refreshedOn);
    return date.toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }
</script>

<WorkspaceContainer>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={ResourceKind.Model}
    slot="header"
    titleInput={fileName}
    showTableToggle
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={handleNameChange}
  >
    <svelte:fragment slot="workspace-controls">
      <p
        class="ui-copy-muted line-clamp-1 mr-2 text-[11px]"
        transition:fade={{ duration: 200 }}
      >
        {#if refreshedOn}
          Computed on {formatRefreshedOn(refreshedOn)}
        {/if}
      </p>
    </svelte:fragment>

    <svelte:fragment slot="cta" let:width>
      {@const collapse = width < 800}

      <div class="flex gap-x-2 items-center">
        <ModelWorkspaceCtAs
          {resource}
          {connector}
          {collapse}
          modelHasError={$hasErrors}
          modelName={assetName}
          hasUnsavedChanges={$hasUnsavedChanges}
        />
      </div>
    </svelte:fragment>
  </WorkspaceHeader>

  <div
    slot="body"
    class="editor-pane size-full overflow-hidden flex flex-col gap-y-0"
  >
    <WorkspaceEditorContainer>
      {#key assetName}
        <ModelEditor {fileArtifact} bind:autoSave={$autoSave} onSave={save} />
      {/key}
    </WorkspaceEditorContainer>

    {#if $tableVisible}
      <WorkspaceTableContainer {filePath}>
        {#if isResourceReconciling}
          <ReconcilingSpinner />
        {:else if connector && tableName}
          <ConnectedPreviewTable {connector} table={tableName} />
        {/if}
        <svelte:fragment slot="error">
          {#if allErrors.length > 0}
            <div
              transition:slide={{ duration: 200 }}
              class="error bottom-4 break-words overflow-auto p-6 border-2 border-gray-300 font-bold text-gray-700 w-full shrink-0 max-h-[60%] bg-gray-100 flex flex-col gap-2"
            >
              {#each allErrors as error (error.message)}
                <div>
                  {getUserFriendlyError(error.message ?? "")}
                </div>
              {/each}
            </div>
          {/if}
        </svelte:fragment>
      </WorkspaceTableContainer>
    {/if}
  </div>

  <WorkspaceInspector
    slot="inspector"
    {filePath}
    {connector}
    {database}
    {databaseSchema}
    {tableName}
    hasErrors={$hasErrors}
    hasUnsavedChanges={$hasUnsavedChanges}
    {resource}
    isEmpty={!$remoteContent?.length}
    {isResourceReconciling}
  />
</WorkspaceContainer>
