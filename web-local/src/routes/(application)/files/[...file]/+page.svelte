<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import { page } from "$app/stores";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
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
  import CustomDashboardPage from "../../custom-dashboard/[name]/+page.svelte";
  import DashboardPage from "../../dashboard/[name]/edit/+page@dashboard.svelte";

  const UNSUPPORTED_EXTENSIONS = [".parquet", ".db", ".db.wal"];
  const FILE_SAVE_DEBOUNCE_TIME = 400;

  $: filePath = $page.params.file;
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
  $: isChart = resourceKind === ResourceKind.Chart;
  $: isCustomDashboard = resourceKind === ResourceKind.Dashboard;
  $: isOther =
    !isSource && !isModel && !isDashboard && !isChart && !isCustomDashboard;

  $: initialLoading = !resourceKind && $fileQuery.isFetching;

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
  $: blob = $fileQuery.data?.blob ?? "";

  // This gets updated via binding below
  $: latest = blob;

  function save(content: string) {
    if ($fileQuery.data?.blob === content) return;
    return $putFile.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        blob: content,
      },
      path: filePath,
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
  <div class="size-full grid place-content-center">
    <div class="flex flex-col items-center gap-y-2">
      <AlertCircleOutline size="40px" />
      <h1>Unsupported file type.</h1>
    </div>
  </div>
{:else if fileError}
  <div class="size-full grid place-content-center">
    <div class="flex flex-col items-center gap-y-2">
      <AlertCircleOutline size="40px" />
      <h1>
        Error loading file: {fileErrorMessage}
      </h1>
    </div>
  </div>
{:else if isSource || isModel}
  <SourceModelPage data={{ fileArtifact }} />
{:else if isDashboard}
  <DashboardPage data={{ fileArtifact }} />
{:else if isChart}
  <ChartPage data={{ fileArtifact }} />
{:else if isCustomDashboard}
  <CustomDashboardPage data={{ fileArtifact }} />
{:else if isOther}
  <WorkspaceContainer inspector={false}>
    <FileWorkspaceHeader filePath={$page.params.file} slot="header" />
    <div class="editor-pane size-full" slot="body">
      <div class="editor flex flex-col border border-gray-200 rounded h-full">
        <div class="grow flex bg-white overflow-y-auto rounded">
          <Editor
            {blob}
            bind:latest
            on:update={({ detail: { content } }) => debounceSave(content)}
          />
        </div>
      </div>
    </div>
  </WorkspaceContainer>
{/if}
