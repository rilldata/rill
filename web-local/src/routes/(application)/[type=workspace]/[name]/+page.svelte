<script lang="ts">
  import { beforeNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";
  import type { SelectionRange } from "@codemirror/state";
  import WorkspaceError from "@rilldata/web-common/components/WorkspaceError.svelte";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    FileArtifact,
    fileArtifacts,
  } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import type { QueryHighlightState } from "@rilldata/web-common/features/models/query-highlight-store";
  import Editor from "@rilldata/web-common/features/models/workspace/Editor.svelte";
  import ModelWorkspaceCtAs from "@rilldata/web-common/features/models/workspace/ModelWorkspaceCTAs.svelte";
  import { createModelFromSource } from "@rilldata/web-common/features/sources/createModel";
  import SourceEditor from "@rilldata/web-common/features/sources/editor/SourceEditor.svelte";
  import UnsavedSourceDialog from "@rilldata/web-common/features/sources/editor/UnsavedSourceDialog.svelte";
  import ErrorPane from "@rilldata/web-common/features/sources/errors/ErrorPane.svelte";
  import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
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
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
    type V1ModelV2,
    type V1SourceV2,
  } from "@rilldata/web-common/runtime-client";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import { getContext, onMount } from "svelte";
  import { get, type Writable } from "svelte/store";
  import { fade, slide } from "svelte/transition";

  export let data: { fileArtifact?: FileArtifact } = {};

  let filePath: string;
  let assetName: string;
  let type: "model" | "source";
  let entity: EntityType;
  let fileArtifact: FileArtifact;

  $: if (data.fileArtifact) {
    fileArtifact = data.fileArtifact;
    filePath = fileArtifact.path;
    assetName = fileArtifact.getEntityName();
    type =
      get(fileArtifact.name)?.kind === ResourceKind.Model ? "model" : "source";
    entity = type === "model" ? EntityType.Model : EntityType.Table;
  } else {
    // needed for backwards compatibility for now
    assetName = $page.params.name;
    type = $page.params.type as "model" | "source";
    entity = type === "model" ? EntityType.Model : EntityType.Table;
    filePath = getFileAPIPathFromNameAndType(assetName, entity);
    fileArtifact = fileArtifacts.getFileArtifact(filePath);
  }
  $: [, fileName] = splitFolderAndName(filePath);

  const QUERY_DEBOUNCE_TIME = 400;

  const updateFile = createRuntimeServicePutFile();

  const queryHighlight = getContext<Writable<QueryHighlightState>>(
    "rill:app:query-highlight",
  );

  const { readOnly } = featureFlags;

  let interceptedUrl: string | null = null;
  let focusOnMount = false;
  let fileNotFound = false;

  onMount(async () => {
    if ($readOnly) await goto("/");
  });

  $: verb = type === "source" ? "Ingested" : "Computed";

  $: pathname = $page.url.pathname;
  $: workspace = workspaces.get(pathname);
  $: autoSave = workspace.editor.autoSave;
  $: tableVisible = workspace.table.visible;

  $: instanceId = $runtime.instanceId;

  $: fileQuery = createRuntimeServiceGetFile(
    instanceId,
    { path: filePath },
    {
      query: {
        onError: () => (fileNotFound = true),
      },
    },
  );

  let blob = "";
  $: blob = ($fileQuery.isFetching ? blob : $fileQuery.data?.blob) ?? "";

  // This gets updated via binding below
  $: latest = blob;

  $: hasUnsavedChanges = latest !== blob;

  $: allErrors = fileArtifact.getAllErrors(queryClient, instanceId);
  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data?.[type];
  $: connector =
    type === "model"
      ? ((resource as V1ModelV2)?.spec?.connector as string)
      : ((resource as V1SourceV2)?.spec?.sinkConnector as string);
  $: tableName = resource?.state?.table;
  $: refreshedOn = resource?.state?.refreshedOn;
  $: resourceIsReconciling = resourceIsLoading($resourceQuery.data);

  let isLocalFileConnectorQuery: CreateQueryResult<boolean>;
  $: if (type === "source") {
    isLocalFileConnectorQuery = useIsLocalFileConnector(instanceId, filePath);
  }
  $: isLocalFileConnector = !!$isLocalFileConnectorQuery?.data;

  $: selections = $queryHighlight?.map((selection) => ({
    from: selection?.referenceIndex,
    to: selection?.referenceIndex + selection?.reference?.length,
  })) as SelectionRange[];

  function revert() {
    latest = blob;
  }

  const debounceSave = debounce(save, QUERY_DEBOUNCE_TIME);

  async function save() {
    if (!hasUnsavedChanges) return;

    if (type === "source") {
      overlay.set({ title: `Importing ${filePath}` });
    } else {
      httpRequestQueue.removeByName(assetName);
      await queryClient.cancelQueries({
        predicate: (query) => isProfilingQuery(query, assetName),
      });
    }

    await $updateFile.mutateAsync({
      instanceId,
      data: {
        path: filePath,
        blob: latest,
      },
    });

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
      assetName,
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

  beforeNavigate((e) => {
    fileNotFound = false;
    if (!hasUnsavedChanges || interceptedUrl) return;

    e.cancel();

    if (e.to) interceptedUrl = e.to.url.href;
  });

  function handleConfirm() {
    if (!interceptedUrl) return;
    const url = interceptedUrl;
    latest = blob;
    hasUnsavedChanges = false;
    interceptedUrl = null;
    goto(url).catch(console.error);
  }

  function handleCancel() {
    interceptedUrl = null;
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

<svelte:head>
  <title>Rill Developer | {assetName}</title>
</svelte:head>

{#if fileNotFound}
  <WorkspaceError message="File not found." />
{:else}
  <WorkspaceContainer>
    <WorkspaceHeader
      slot="header"
      titleInput={fileName}
      showTableToggle
      {hasUnsavedChanges}
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
              {hasUnsavedChanges}
              {collapse}
              hasErrors={$hasErrors}
              {isLocalFileConnector}
              on:save-source={save}
              on:revert-source={revert}
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

    <div
      slot="body"
      class="editor-pane size-full overflow-hidden flex flex-col"
    >
      <WorkspaceEditorContainer>
        {#key assetName}
          {#if type === "source"}
            <SourceEditor
              {filePath}
              {blob}
              {hasUnsavedChanges}
              allErrors={$allErrors}
              bind:latest
              on:save={debounceSave}
            />
          {:else}
            <Editor
              {blob}
              {selections}
              {focusOnMount}
              {hasUnsavedChanges}
              bind:latest
              bind:autoSave={$autoSave}
              on:revert={revert}
              on:save={debounceSave}
            />
          {/if}
        {/key}
      </WorkspaceEditorContainer>

      {#if $tableVisible}
        <WorkspaceTableContainer fade={type === "source" && hasUnsavedChanges}>
          {#if type === "source" && $allErrors[0]?.message}
            <ErrorPane {filePath} errorMessage={$allErrors[0].message} />
          {:else}
            <ConnectedPreviewTable
              {connector}
              table={tableName ?? ""}
              loading={resourceIsReconciling}
            />
          {/if}
          <svelte:fragment slot="error">
            {#if type === "model"}
              {#if $allErrors.length > 0}
                <div
                  transition:slide={{ duration: 200 }}
                  class="error bottom-4 break-words overflow-auto p-6 border-2 border-gray-300 font-bold text-gray-700 w-full shrink-0 max-h-[60%] z-10 bg-gray-100 flex flex-col gap-2"
                >
                  {#each $allErrors as error}
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
          {hasUnsavedChanges}
          {...{
            [type]: resource,
          }}
          isEmpty={!blob.length}
          sourceIsReconciling={resourceIsReconciling}
        />
      {/if}
    </svelte:fragment>
  </WorkspaceContainer>
{/if}

{#if interceptedUrl}
  <UnsavedSourceDialog
    context={type}
    on:confirm={handleConfirm}
    on:cancel={handleCancel}
  />
{/if}
