<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import {
    getFileAPIPathFromNameAndType,
    getNameFromFile,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import ModelEditor from "@rilldata/web-common/features/models/workspace/ModelEditor.svelte";
  import ModelWorkspaceCtAs from "@rilldata/web-common/features/models/workspace/ModelWorkspaceCTAs.svelte";
  import { createModelFromSource } from "@rilldata/web-common/features/sources/createModel";
  import SourceEditor from "@rilldata/web-common/features/sources/editor/SourceEditor.svelte";
  import ErrorPane from "@rilldata/web-common/features/sources/errors/ErrorPane.svelte";
  import WorkspaceInspector from "@rilldata/web-common/features/sources/inspector/WorkspaceInspector.svelte";
  import {
    refreshSource,
    replaceSourceWithUploadedFile,
  } from "@rilldata/web-common/features/sources/refreshSource";
  import { useIsLocalFileConnector } from "@rilldata/web-common/features/sources/selectors";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import SourceCTAs from "@rilldata/web-common/features/sources/workspace/SourceCTAs.svelte";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
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
  import type {
    V1ModelV2,
    V1SourceV2,
  } from "@rilldata/web-common/runtime-client";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import { fade, slide } from "svelte/transition";

  const { readOnly } = featureFlags;

  export let data: { fileArtifact?: FileArtifact } = {};

  let type: "model" | "source";

  onMount(async () => {
    if ($readOnly) await goto("/");
  });

  $: type = data.fileArtifact
    ? get(fileArtifact.name)?.kind === ResourceKind.Model
      ? "model"
      : "source"
    : ($page.params.type as "model" | "source");
  $: entity = type === "model" ? EntityType.Model : EntityType.Table;

  $: assetName = getNameFromFile(filePath);

  $: fileArtifact = data?.fileArtifact ?? getLegacyFileArtifact();

  $: verb = type === "source" ? "Ingested" : "Computed";

  $: pathname = $page.url.pathname;
  $: workspace = workspaces.get(pathname);
  $: tableVisible = workspace.table.visible;

  $: instanceId = $runtime.instanceId;

  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    fileName,
    remoteContent,
  } = fileArtifact);

  $: allErrorsStore = fileArtifact.getAllErrors(queryClient, instanceId);
  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  $: allErrors = $allErrorsStore;

  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data?.[type];
  $: connector =
    type === "model"
      ? ((resource as V1ModelV2)?.spec?.outputConnector as string)
      : ((resource as V1SourceV2)?.spec?.sinkConnector as string);
  $: tableName =
    type === "model"
      ? ((resource as V1ModelV2)?.state?.resultTable as string)
      : ((resource as V1SourceV2)?.state?.table as string);
  $: refreshedOn = resource?.state?.refreshedOn;
  $: resourceIsReconciling = resourceIsLoading($resourceQuery.data);

  let isLocalFileConnectorQuery: CreateQueryResult<boolean>;
  $: if (type === "source") {
    isLocalFileConnectorQuery = useIsLocalFileConnector(instanceId, filePath);
  }
  $: isLocalFileConnector = !!$isLocalFileConnectorQuery?.data;

  async function save() {
    if (type === "source") {
      overlay.set({ title: `Importing ${filePath}` });
    } else {
      httpRequestQueue.removeByName(assetName);
      await queryClient.cancelQueries({
        predicate: (query) => isProfilingQuery(query, assetName),
      });
    }

    if (type === "source") {
      await checkSourceImported(queryClient, filePath);
      overlay.set(null);
    }
  }

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
    const [newModelPath, newModelName] = await createModelFromSource(
      assetName,
      tableName ?? "",
      "models",
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

  function getLegacyFileArtifact() {
    const assetName = $page.params.name;
    const filePath = getFileAPIPathFromNameAndType(assetName, entity);
    return fileArtifacts.getFileArtifact(filePath);
  }
</script>

<svelte:head>
  <title>Rill Developer | {fileName}</title>
</svelte:head>

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
          {verb} on {formatRefreshedOn(refreshedOn)}
        {/if}
      </p>
    </svelte:fragment>

    <svelte:fragment slot="cta" let:width>
      {@const collapse = width < 800}

      <div class="flex gap-x-2 items-center">
        {#if type === "source"}
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
        {:else}
          <ModelWorkspaceCtAs
            {collapse}
            modelHasError={$hasErrors}
            modelName={assetName}
          />
        {/if}
      </div>
    </svelte:fragment>
  </WorkspaceHeader>

  <div slot="body" class="editor-pane size-full overflow-hidden flex flex-col">
    <WorkspaceEditorContainer>
      {#key assetName}
        {#if type === "source"}
          <SourceEditor {fileArtifact} {allErrors} onSave={save} />
        {:else}
          <ModelEditor {fileArtifact} bind:autoSave={$autoSave} onSave={save} />
        {/if}
      {/key}
    </WorkspaceEditorContainer>

    {#if $tableVisible}
      <WorkspaceTableContainer fade={type === "source" && $hasUnsavedChanges}>
        {#if type === "source" && allErrors[0]?.message}
          <ErrorPane {filePath} errorMessage={allErrors[0].message} />
        {:else if !allErrors.length}
          <ConnectedPreviewTable
            {connector}
            table={tableName ?? ""}
            loading={resourceIsReconciling}
          />
        {/if}
        <svelte:fragment slot="error">
          {#if type === "model"}
            {#if allErrors.length > 0}
              <div
                transition:slide={{ duration: 200 }}
                class="error bottom-4 break-words overflow-auto p-6 border-2 border-gray-300 font-bold text-gray-700 w-full shrink-0 max-h-[60%] z-10 bg-gray-100 flex flex-col gap-2"
              >
                {#each allErrors as error (error.message)}
                  <div>{error.message}</div>
                {/each}
              </div>
            {/if}
          {/if}
        </svelte:fragment>
      </WorkspaceTableContainer>
    {/if}
  </div>

  <svelte:fragment slot="inspector">
    {#if tableName && resource}
      <WorkspaceInspector
        {tableName}
        hasErrors={$hasErrors}
        hasUnsavedChanges={$hasUnsavedChanges}
        {...{
          [type]: resource,
        }}
        isEmpty={!$remoteContent?.length}
        sourceIsReconciling={resourceIsReconciling}
      />
    {/if}
  </svelte:fragment>
</WorkspaceContainer>
