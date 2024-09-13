<script lang="ts">
  import { goto } from "$app/navigation";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import SourceEditor from "@rilldata/web-common/features/sources/editor/SourceEditor.svelte";
  import ErrorPane from "@rilldata/web-common/features/sources/errors/ErrorPane.svelte";
  import WorkspaceInspector from "@rilldata/web-common/features/sources/inspector/WorkspaceInspector.svelte";
  import {
    refreshSource,
    replaceSourceWithUploadedFile,
  } from "@rilldata/web-common/features/sources/refreshSource";
  import { useIsLocalFileConnector } from "@rilldata/web-common/features/sources/selectors";
  import SourceCTAs from "@rilldata/web-common/features/sources/workspace/SourceCTAs.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import WorkspaceTableContainer from "@rilldata/web-common/layout/workspace/WorkspaceTableContainer.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import type { V1SourceV2 } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { fade } from "svelte/transition";
  import { createModelFromTable } from "../connectors/olap/createModel";

  export let fileArtifact: FileArtifact;

  $: ({ instanceId } = $runtime);

  $: ({
    hasUnsavedChanges,
    path: filePath,
    fileName,
    remoteContent,
  } = fileArtifact);

  $: assetName = getNameFromFile(filePath);

  $: workspace = workspaces.get(filePath);
  $: tableVisible = workspace.table.visible;

  $: allErrorsStore = fileArtifact.getAllErrors(queryClient, instanceId);
  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  $: allErrors = $allErrorsStore;

  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data?.source;
  $: connector = (resource as V1SourceV2)?.spec?.sinkConnector as string;
  const database = ""; // Sources are ingested into the default database
  const databaseSchema = ""; // Sources are ingested into the default database schema
  $: tableName = (resource as V1SourceV2)?.state?.table as string;
  $: refreshedOn = resource?.state?.refreshedOn;
  $: resourceIsReconciling = resourceIsLoading($resourceQuery.data);

  $: isLocalFileConnectorQuery = useIsLocalFileConnector(instanceId, filePath);
  $: isLocalFileConnector = !!$isLocalFileConnectorQuery?.data;

  async function replaceSource() {
    await replaceSourceWithUploadedFile(instanceId, filePath);
  }

  async function handleNameChange(
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    const newRoute = await handleEntityRename(
      instanceId,
      e.currentTarget,
      filePath,
      fileName,
      [
        ...fileArtifacts.getNamesForKind(ResourceKind.Source),
        ...fileArtifacts.getNamesForKind(ResourceKind.Model),
      ],
    );

    if (newRoute) await goto(newRoute);
  }

  function refresh() {
    if (connector === undefined) return;

    refreshSource(
      connector,
      filePath,
      $resourceQuery.data?.meta?.name?.name ?? "",
      instanceId,
    ).catch(() => {});
  }

  async function handleCreateModelFromSource() {
    const addDevLimit = false; // Typically, the `dev` limit would be applied on the Source itself
    const [newModelPath, newModelName] = await createModelFromTable(
      queryClient,
      connector,
      database,
      databaseSchema,
      tableName,
      addDevLimit,
    );
    await goto(`/files${newModelPath}`);
    await behaviourEvent.fireNavigationEvent(
      newModelName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.RightPanel,
      MetricsEventScreenName.Source,
      MetricsEventScreenName.Model,
    );
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
    slot="header"
    titleInput={fileName}
    showTableToggle
    hasUnsavedChanges={$hasUnsavedChanges}
    on:change={handleNameChange}
  >
    <svelte:fragment slot="workspace-controls">
      <p
        class="ui-copy-muted line-clamp-1 mr-2 text-[11px]"
        transition:fade={{ duration: 200 }}
      >
        {#if refreshedOn}
          Ingested on {formatRefreshedOn(refreshedOn)}
        {/if}
      </p>
    </svelte:fragment>

    <svelte:fragment slot="cta" let:width>
      {@const collapse = width < 800}

      <div class="flex gap-x-2 items-center">
        <SourceCTAs
          hasUnsavedChanges={$hasUnsavedChanges}
          {collapse}
          hasErrors={$hasErrors}
          {isLocalFileConnector}
          on:save-source={fileArtifact.saveLocalContent}
          on:refresh-source={refresh}
          on:replace-source={replaceSource}
          on:create-model={handleCreateModelFromSource}
        />
      </div>
    </svelte:fragment>
  </WorkspaceHeader>

  <div slot="body" class="editor-pane size-full overflow-hidden flex flex-col">
    <WorkspaceEditorContainer>
      {#key assetName}
        <SourceEditor {fileArtifact} {allErrors} />
      {/key}
    </WorkspaceEditorContainer>

    {#if $tableVisible}
      <WorkspaceTableContainer fade={$hasUnsavedChanges}>
        {#if allErrors[0]?.message}
          <ErrorPane {filePath} errorMessage={allErrors[0].message} />
        {:else if !allErrors.length}
          <ConnectedPreviewTable
            {connector}
            table={tableName ?? ""}
            loading={resourceIsReconciling}
          />
        {/if}
      </WorkspaceTableContainer>
    {/if}
  </div>

  <svelte:fragment slot="inspector">
    {#if connector && tableName && resource}
      <WorkspaceInspector
        {connector}
        {database}
        {databaseSchema}
        {tableName}
        hasErrors={$hasErrors}
        hasUnsavedChanges={$hasUnsavedChanges}
        source={resource}
        isEmpty={!$remoteContent?.length}
        sourceIsReconciling={resourceIsReconciling}
      />
    {/if}
  </svelte:fragment>
</WorkspaceContainer>
