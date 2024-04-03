<script lang="ts">
  import UnsavedSourceDialog from "@rilldata/web-common/features/sources/editor/UnsavedSourceDialog.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import SourceInspector from "@rilldata/web-common/features/sources/inspector/SourceInspector.svelte";
  import SourceWorkspaceHeader from "@rilldata/web-common/features/sources/workspace/SourceWorkspaceHeader.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import SourceEditor from "@rilldata/web-common/features/sources/editor/SourceEditor.svelte";
  import WorkspaceTableContainer from "@rilldata/web-common/layout/workspace/WorkspaceTableContainer.svelte";
  import ConnectedPreviewTable from "@rilldata/web-common/components/preview-table/ConnectedPreviewTable.svelte";
  import ErrorPane from "@rilldata/web-common/features/generic-yaml-editor/ErrorPane.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { resourceIsLoading } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { saveAndRefresh } from "@rilldata/web-common/features/sources/saveAndRefresh";
  import { checkSourceImported } from "@rilldata/web-common/features/sources/source-imported-utils";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { beforeNavigate, goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
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

  const { readOnly } = featureFlags;

  let latest: string;
  let interceptedUrl: string | null = null;

  onMount(async () => {
    if ($readOnly) await goto("/");
  });

  $: sourceName = $page.params.name;
  $: instanceId = $runtime.instanceId;
  $: filePath = getFileAPIPathFromNameAndType(sourceName, EntityType.Table);

  $: fileQuery = createRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      onError: () => {
        goto("/").catch(() => {});
      },
    },
  });

  $: yaml = $fileQuery.data?.blob ?? "";

  $: isSourceUnsaved = latest !== yaml;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: allErrors = fileArtifact.getAllErrors(queryClient, instanceId);
  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  $: sourceQuery = fileArtifact.getResource(queryClient, instanceId);
  $: source = $sourceQuery.data?.source;
  $: connector = source?.state?.connector;
  $: tableName = source?.state?.table ?? "";
  $: refreshedOn = source?.state?.refreshedOn;
  $: sourceIsReconciling = resourceIsLoading($sourceQuery.data);

  $: isLocalFileConnectorQuery = useIsLocalFileConnector(instanceId, filePath);
  $: isLocalFileConnector = !!$isLocalFileConnectorQuery.data;

  function revert() {
    latest = yaml;
  }

  async function save() {
    overlay.set({ title: `Importing ${filePath}` });
    await saveAndRefresh(filePath, latest);
    checkSourceImported(queryClient, filePath);
    overlay.set(null);
  }

  async function replaceSource() {
    await replaceSourceWithUploadedFile(instanceId, filePath);
  }

  function onChangeCallback(
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    return handleEntityRename(
      queryClient,
      instanceId,
      e.currentTarget,
      filePath,
      EntityType.Table,
    );
  }

  function refresh() {
    if (connector === undefined) return;

    refreshSource(
      connector,
      filePath,
      $sourceQuery.data?.meta?.name?.name ?? "",
      instanceId,
    ).catch(() => {});
  }

  async function handleCreateModelFromSource() {
    const modelName = await createModelFromSourceV2(
      queryClient,
      source?.state?.table ?? "",
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
    if (!isSourceUnsaved || interceptedUrl) return;

    e.cancel();

    if (e.to) interceptedUrl = e.to.url.href;
  });

  async function handleConfirm() {
    if (interceptedUrl) await goto(interceptedUrl);

    interceptedUrl = null;
  }

  function handleCancel() {
    interceptedUrl = null;
  }
</script>

<svelte:head>
  <title>Rill Developer | {sourceName}</title>
</svelte:head>

<WorkspaceContainer>
  <SourceWorkspaceHeader
    slot="header"
    {sourceName}
    {refreshedOn}
    {isSourceUnsaved}
    {sourceIsReconciling}
    {isLocalFileConnector}
    hasErrors={$hasErrors}
    on:save={save}
    on:revert={revert}
    on:refresh-source={refresh}
    on:change={onChangeCallback}
    on:replace-source={replaceSource}
    on:create-model={handleCreateModelFromSource}
  />

  <div class="editor-pane size-full overflow-hidden flex flex-col" slot="body">
    <WorkspaceEditorContainer>
      <SourceEditor
        {yaml}
        {isSourceUnsaved}
        allErrors={$allErrors}
        bind:latest
        on:save={save}
      />
    </WorkspaceEditorContainer>

    <WorkspaceTableContainer fade={isSourceUnsaved}>
      {#if $allErrors[0]?.message}
        <ErrorPane errorMessage={$allErrors[0].message} />
      {:else}
        <ConnectedPreviewTable
          objectName={$sourceQuery?.data?.source?.state?.table}
          loading={resourceIsLoading($sourceQuery?.data)}
        />
      {/if}
    </WorkspaceTableContainer>
  </div>

  <SourceInspector
    slot="inspector"
    {source}
    {tableName}
    {isSourceUnsaved}
    {sourceIsReconciling}
  />
</WorkspaceContainer>

{#if interceptedUrl}
  <UnsavedSourceDialog
    context="source"
    on:confirm={handleConfirm}
    on:cancel={handleCancel}
  />
{/if}
