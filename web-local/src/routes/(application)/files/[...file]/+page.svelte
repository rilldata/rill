<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import WorkspaceError from "@rilldata/web-common/components/WorkspaceError.svelte";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import { getExtensionsForFiles } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import {
    addLeadingSlash,
    removeLeadingSlash,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import { extractFileExtension } from "@rilldata/web-common/features/sources/extract-file-name";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServicePutFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
  import SourceModelPage from "../../[type=workspace]/[name]/+page.svelte";
  import ChartPage from "../../chart/[name]/+page.svelte";
  import CustomDashboardPage from "../../custom-dashboards/[name]/+page.svelte";
  import DashboardPage from "../../dashboard/[name]/edit/+page.svelte";

  const UNSUPPORTED_EXTENSIONS = [".parquet", ".db", ".db.wal"];
  const FILE_SAVE_DEBOUNCE_TIME = 400;

  $: filePath = addLeadingSlash($page.params.file);
  $: fileExtension = extractFileExtension(filePath);
  $: fileTypeUnsupported = UNSUPPORTED_EXTENSIONS.includes(fileExtension);

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath, {
    query: {
      enabled: !fileTypeUnsupported,
    },
  });
  $: fileError = !!$fileQuery.error;
  $: fileErrorMessage = $fileQuery.error?.response.data.message;
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: name = fileArtifact.name;
  $: resourceKind = $name?.kind;

  $: isSource = resourceKind === ResourceKind.Source;
  $: isModel = resourceKind === ResourceKind.Model;
  $: isDashboard = resourceKind === ResourceKind.MetricsView;
  $: isChart = resourceKind === ResourceKind.Component;
  $: isCustomDashboard = resourceKind === ResourceKind.Dashboard;
  $: isOther =
    !isSource && !isModel && !isDashboard && !isChart && !isCustomDashboard;

  // TODO: optimistically update the get file cache
  const putFile = createRuntimeServicePutFile();

  onMount(() => {
    expandDirectory(filePath);

    // TODO: Focus on the code editor
  });

  afterNavigate(() => {
    expandDirectory(filePath);

    // TODO: Focus on the code editor
  });

  const debounceSave = debounce(save, FILE_SAVE_DEBOUNCE_TIME);
  let blob = "";
  $: blob = $fileQuery.data?.blob ?? blob;

  $: latest = blob;

  function save(content: string) {
    if ($fileQuery.data?.blob === content) return;
    return $putFile.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        blob: content,
      },
      path: removeLeadingSlash(filePath),
    });
  }

  // TODO: move this logic into the DirectoryState
  // TODO: expand all directories in the path, not just the last one
  function expandDirectory(filePath: string) {
    const directory = filePath.split("/").slice(0, -1).join("/");
    directoryState.expand(directory);
  }
</script>

{#if fileTypeUnsupported}
  <WorkspaceError message="Unsupported file type." />
{:else if fileError}
  <WorkspaceError message={`Error loading file: ${fileErrorMessage}`} />
{:else if isSource || isModel}
  <SourceModelPage data={{ fileArtifact }} />
{:else if isDashboard}
  <DashboardPage data={{ fileArtifact }} />
{:else if isChart}
  {#key $page.params.file}
    <ChartPage data={{ fileArtifact }} />
  {/key}
{:else if isCustomDashboard}
  <CustomDashboardPage data={{ fileArtifact }} />
{:else if isOther}
  <WorkspaceContainer inspector={false}>
    <FileWorkspaceHeader filePath={$page.params.file} slot="header" />
    <div class="editor-pane size-full" slot="body">
      <div class="editor flex flex-col h-full">
        <div class="grow flex bg-white overflow-y-auto">
          <Editor
            {blob}
            bind:latest
            extensions={getExtensionsForFiles(filePath)}
            on:update={({ detail: { content } }) => debounceSave(content)}
          />
        </div>
      </div>
    </div>
  </WorkspaceContainer>
{/if}
