<script lang="ts">
  import {
    deleteFileArtifact,
    renameFileArtifact,
  } from "@rilldata/web-common/features/entity-management/actions";
  import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-selectors";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import {
    NavDragData,
    navEntryDragDropStore,
  } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import NavEntryPortal from "@rilldata/web-common/features/file-explorer/NavEntryPortal.svelte";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import NavDirectory from "./NavDirectory.svelte";
  import { transformFileList } from "./transform-file-list";

  $: getFileTree = createRuntimeServiceListFiles("default", undefined, {
    query: {
      select: (data) => {
        if (!data || !data.files?.length) return;

        const files = data.files
          // remove leading slash
          .map((file) => ({
            path: file.path?.slice(1) ?? "",
            isDir: !!file.isDir,
          }))
          // sort alphabetically case-insensitive
          .sort((a, b) =>
            a.path.localeCompare(b.path, undefined, { sensitivity: "base" }),
          );

        const fileTree = transformFileList(files);

        return fileTree;
      },
    },
  });
  $: instanceId = $runtime.instanceId;

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
  }

  const { navDragging, offset, dragStart } = navEntryDragDropStore;
  function handleDropSuccess(
    fromDragData: NavDragData,
    toDragData: NavDragData,
  ) {
    const tarDir = toDragData.isDir
      ? toDragData.filePath
      : splitFolderAndName(toDragData.filePath)[0];
    const [, srcFile] = splitFolderAndName(fromDragData.filePath);
    const newFilePath = `${tarDir}/${srcFile}`;
    if (fromDragData.filePath !== newFilePath) {
      void renameFileArtifact(instanceId, fromDragData.filePath, newFilePath);
    }
  }
</script>

<svelte:window
  on:mousemove={() => navEntryDragDropStore.onMouseMove()}
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

{#if $navDragging}
  <NavEntryPortal offset={$offset} position={$dragStart} />
{/if}
