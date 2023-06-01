<script lang="ts">
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardStateProvider from "@rilldata/web-common/features/dashboards/proto-state/DashboardStateProvider.svelte";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FilesWorkspaceHeader from "@rilldata/web-common/features/editor/FilesWorkspaceHeader.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { ModelWorkspace } from "@rilldata/web-common/features/models";
  import SourceInspector from "@rilldata/web-common/features/sources/workspace/SourceInspector.svelte";
  import SourceWorkspaceBody from "@rilldata/web-common/features/sources/workspace/SourceWorkspaceBody.svelte";
  import SourceWorkspaceHeader from "@rilldata/web-common/features/sources/workspace/SourceWorkspaceHeader.svelte";
  import { appStore } from "@rilldata/web-common/layout/app-store";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import {
    createRuntimeServiceGetCatalogEntry,
    createRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";

  $: getFile = createRuntimeServiceGetFile("default", $page.params.file);
  $: catalogEntryName = getCatalogEntryNameFromFilePath($page.params.file);

  function getCatalogEntryNameFromFilePath(filePath: string) {
    // remove file extension
    const pathWithoutExtension = filePath.replace(/\.[^/.]+$/, "");
    // get the last part of the path
    const pathParts = pathWithoutExtension.split("/");
    return pathParts[pathParts.length - 1];
  }

  // TODO: conditionally fetch catalog entry based on file type
  $: getCatalogEntry = createRuntimeServiceGetCatalogEntry(
    "default",
    catalogEntryName
  );
  $: isSource = !!$getCatalogEntry.data?.entry.source;
  $: isModel = !!$getCatalogEntry.data?.entry.model;
  $: isDashboard = !!$getCatalogEntry.data?.entry.metricsView;

  $: console.log(
    "isSource",
    isSource,
    "isModel",
    isModel,
    "isDashboard",
    isDashboard
  );

  const switchToSource = async (name: string) => {
    if (!name) return;
    appStore.setActiveEntity(name, EntityType.Table);
  };

  $: if (isSource) switchToSource(catalogEntryName);

  // TODO: optimistically update the get file cache
  // const putFile = createRuntimeServicePutFileAndReconcile();
</script>

<!-- on:write={(evt) => $putFile.mutate(evt.detail.blob)} -->
{#if isSource}
  <WorkspaceContainer assetID={$page.params.file}>
    <!-- <SourceWorkspace sourceName={$getCatalogEntry.data.entry.name} /> -->
    <SourceWorkspaceHeader
      sourceName={catalogEntryName}
      path={$page.params.file}
      slot="header"
    />
    <SourceWorkspaceBody sourceName={catalogEntryName} slot="body" />
    <SourceInspector sourceName={catalogEntryName} slot="inspector" />
  </WorkspaceContainer>
{:else if isModel}
  <ModelWorkspace modelName={catalogEntryName} />
{:else if isDashboard}
  <WorkspaceContainer
    top="0px"
    assetID={catalogEntryName}
    bgClass="bg-white"
    inspector={false}
  >
    <DashboardStateProvider metricViewName={catalogEntryName} slot="body">
      <Dashboard metricViewName={catalogEntryName} hasTitle />
    </DashboardStateProvider>
  </WorkspaceContainer>
{:else}
  <WorkspaceContainer assetID={$page.params.file}>
    <FilesWorkspaceHeader filePath={$page.params.file} slot="header" />
    <Editor
      content={$getFile.data?.blob}
      on:write={(evt) => console.log(evt)}
      slot="body"
    />
  </WorkspaceContainer>
{/if}
