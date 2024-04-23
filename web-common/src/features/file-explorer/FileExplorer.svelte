<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import GenerateChartYAMLPrompt from "@rilldata/web-common/features/charts/prompt/GenerateChartYAMLPrompt.svelte";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import {
    deleteFileArtifact,
    renameFileArtifact,
  } from "@rilldata/web-common/features/entity-management/actions";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import NavEntryPortal from "@rilldata/web-common/features/file-explorer/NavEntryPortal.svelte";
  import {
    NavDragData,
    navEntryDragDropStore,
  } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { PROTECTED_DIRECTORIES } from "@rilldata/web-common/features/file-explorer/protected-paths";
  import {
    getTopLevelFolder,
    splitFolderAndName,
  } from "@rilldata/web-common/features/sources/extract-file-name";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import NavDirectory from "./NavDirectory.svelte";
  import { transformFileList } from "./transform-file-list";

  $: instanceId = $runtime.instanceId;
  $: getFileTree = createRuntimeServiceListFiles("default", undefined, {
    query: {
      select: (data) => {
        if (!data || !data.files?.length) return;

        const files = data.files
          // Sort alphabetically case-insensitive
          .sort(
            (a, b) =>
              a.path?.localeCompare(b.path ?? "", undefined, {
                sensitivity: "base",
              }) ?? 0,
          )
          // Hide dot directories
          .filter(
            (file) =>
              !(
                file.isDir &&
                // Check both the top-level directory and subdirectories
                (file.path?.startsWith(".") || file.path?.includes("/."))
              ),
          )
          // Hide the `tmp` directory
          .filter((file) => !file.path?.startsWith("/tmp"));

        return transformFileList(files);
      },
    },
  });

  let showRenameModelModal = false;
  let renameFilePath: string;
  let renameIsDir: boolean;

  function onRename(filePath: string, isDir: boolean) {
    showRenameModelModal = true;
    renameFilePath = filePath;
    renameIsDir = isDir;
  }

  async function onDelete(filePath: string) {
    await deleteFileArtifact(instanceId, filePath);
    if (
      !!$page.params.file &&
      removeLeadingSlash($page.params.file) === removeLeadingSlash(filePath)
    ) {
      await goto("/");
    }
  }

  let showGenerateChartModal = false;
  let generateChartTable: string;
  let generateChartConnector: string;
  let generateChartMetricsView: string;

  function onGenerateChart({
    table,
    connector,
    metricsView,
  }: {
    table?: string;
    connector?: string;
    metricsView?: string;
  }) {
    showGenerateChartModal = true;
    generateChartTable = table ?? "";
    generateChartConnector = connector ?? "";
    generateChartMetricsView = metricsView ?? "";
  }

  const { dragData, position } = navEntryDragDropStore;

  async function handleDropSuccess(
    fromDragData: NavDragData,
    toDragData: NavDragData,
  ) {
    const isCurrentFile =
      removeLeadingSlash(fromDragData.filePath) ===
      removeLeadingSlash($page.params.file);
    const tarDir = toDragData.isDir
      ? toDragData.filePath
      : splitFolderAndName(toDragData.filePath)[0];
    const [, srcFile] = splitFolderAndName(fromDragData.filePath);
    const newFilePath = `${tarDir}/${srcFile}`;

    if (fromDragData.filePath !== newFilePath) {
      const newTopLevelPath = getTopLevelFolder(newFilePath);
      if (PROTECTED_DIRECTORIES.includes(newTopLevelPath)) {
        notifications.send({
          message: "cannot move to restricted directories",
        });
        return;
      }
      await renameFileArtifact(instanceId, fromDragData.filePath, newFilePath);

      if (isCurrentFile) {
        await goto(`/files${newFilePath}`);
      }
    }
  }
</script>

<svelte:window
  on:mousemove={(e) => navEntryDragDropStore.onMouseMove(e)}
  on:mouseup={(e) =>
    navEntryDragDropStore.onMouseUp(e, null, handleDropSuccess)}
/>

<div class="flex flex-col items-start gap-y-2">
  <!-- File tree -->
  <div class="flex flex-col w-full items-start justify-start overflow-auto">
    {#if $getFileTree.data}
      <NavDirectory
        directory={$getFileTree.data}
        {onRename}
        {onDelete}
        {onGenerateChart}
        onMouseDown={(e, dragData) =>
          navEntryDragDropStore.onMouseDown(e, dragData)}
        onMouseUp={(e, dragData) =>
          navEntryDragDropStore.onMouseUp(e, dragData, handleDropSuccess)}
      />
    {/if}
  </div>
</div>

{#if showRenameModelModal}
  <RenameAssetModal
    closeModal={() => (showRenameModelModal = false)}
    filePath={renameFilePath}
    isDir={renameIsDir}
  />
{/if}

<GenerateChartYAMLPrompt
  bind:open={showGenerateChartModal}
  connector={generateChartConnector}
  metricsView={generateChartMetricsView}
  table={generateChartTable}
/>

{#if $dragData}
  <NavEntryPortal position={$position} dragData={$dragData} />
{/if}
