<script lang="ts">
  import UnsavedSourceDialog from "@rilldata/web-common/features/sources/editor/UnsavedSourceDialog.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import SourceInspector from "@rilldata/web-common/features/sources/inspector/SourceInspector.svelte";
  import SourceWorkspaceHeader from "@rilldata/web-common/features/sources/workspace/SourceWorkspaceHeader.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import SourceEditor from "@rilldata/web-common/features/sources/editor/SourceEditor.svelte";
  import WorkspaceTableContainer from "@rilldata/web-common/layout/workspace/WorkspaceTableContainer.svelte";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import ErrorPane from "@rilldata/web-common/features/sources/errors/ErrorPane.svelte";
  import Editor from "@rilldata/web-common/features/models/workspace/Editor.svelte";
  import ModelInspector from "@rilldata/web-common/features/models/workspace/inspector/ModelInspector.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { beforeNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { getContext, onMount } from "svelte";
  import { useIsLocalFileConnector } from "@rilldata/web-common/features/sources/selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import {
    refreshSource,
    replaceSourceWithUploadedFile,
  } from "@rilldata/web-common/features/sources/refreshSource";
  import { createModelFromSourceV2 } from "@rilldata/web-common/features/sources/createModel";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import type { SelectionRange } from "@codemirror/state";
  import { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
  import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import type { QueryHighlightState } from "@rilldata/web-common/features/models/query-highlight-store";
  import type { Writable } from "svelte/store";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { slide } from "svelte/transition";

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

  $: assetName = $page.params.name;
  $: type = $page.params.type as "model" | "source";
  $: entity = type === "model" ? EntityType.Model : EntityType.Table;

  $: pathname = $page.url.pathname;
  $: workspace = workspaces.get(pathname);
  $: autoSave = workspace.editor.autoSave;
  $: tableVisible = workspace.table.visible;

  $: instanceId = $runtime.instanceId;
  $: filePath = getFileAPIPathFromNameAndType(assetName, entity);

  $: fileQuery = createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      onError: () => (fileNotFound = true),
    },
  });

  $: blob = $fileQuery.data?.blob ?? "";

  // This gets updated via binding below
  $: latest = blob;

  $: hasUnsavedChanges = latest !== blob;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: allErrors = fileArtifact.getAllErrors(queryClient, instanceId);
  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  $: resourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: resource = $resourceQuery.data?.[type];
  $: connector = resource?.state?.connector;
  $: tableName = resource?.state?.table;
  $: refreshedOn = resource?.state?.refreshedOn;
  $: resourceIsReconciling = resourceIsLoading($resourceQuery.data);

  $: isLocalFileConnectorQuery = useIsLocalFileConnector(instanceId, filePath);
  $: isLocalFileConnector =
    type === "source" && !!$isLocalFileConnectorQuery.data;

  $: selections = $queryHighlight?.map((selection) => ({
    from: selection?.referenceIndex,
    to: selection?.referenceIndex + selection?.reference?.length,
  })) as SelectionRange[];

  function revert() {
    latest = blob;
  }

  const debounceSave = debounce(save, QUERY_DEBOUNCE_TIME);

  async function save() {
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
      path: getFileAPIPathFromNameAndType(assetName, entity),
      data: {
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
      entity,
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
    const modelName = await createModelFromSourceV2(
      queryClient,
      tableName ?? "",
    );
    await goto(`/model/${modelName}`);
    await behaviourEvent.fireNavigationEvent(
      modelName,
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
</script>

<svelte:head>
  <title>Rill Developer | {assetName}</title>
</svelte:head>

{#if fileNotFound}
  <div class="size-full grid place-content-center">
    <div class="flex flex-col items-center gap-y-2">
      <AlertCircleOutline size="40px" />
      <h1>
        Unable to find file {assetName}
      </h1>
    </div>
  </div>
{:else}
  <WorkspaceContainer>
    <SourceWorkspaceHeader
      slot="header"
      {type}
      {assetName}
      {refreshedOn}
      {hasUnsavedChanges}
      {resourceIsReconciling}
      {isLocalFileConnector}
      hasErrors={$hasErrors}
      on:save-source={save}
      on:revert-source={revert}
      on:refresh-source={refresh}
      on:change={handleNameChange}
      on:replace-source={replaceSource}
      on:create-model={handleCreateModelFromSource}
    />

    <div
      class="editor-pane size-full overflow-hidden flex flex-col"
      slot="body"
    >
      <WorkspaceEditorContainer>
        {#key assetName}
          {#if type === "source"}
            <SourceEditor
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
          {:else if tableName}
            <ConnectedPreviewTable objectName={tableName} />
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
      {#if tableName}
        {#if type === "source"}
          <SourceInspector
            {tableName}
            {hasUnsavedChanges}
            source={resource}
            sourceIsReconciling={resourceIsReconciling}
          />
        {:else}
          <ModelInspector
            modelName={assetName}
            hasErrors={$hasErrors}
            {resourceIsReconciling}
            modelIsEmpty={!blob.length}
          />
        {/if}
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
