<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import {
    deleteFileArtifact,
    duplicateFileArtifact,
    renameFileArtifact,
  } from "@rilldata/web-common/features/entity-management/actions";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    getTopLevelFolder,
    splitFolderAndFileName,
  } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import ForceDeleteConfirmation from "@rilldata/web-common/features/file-explorer/ForceDeleteConfirmationDialog.svelte";
  import NavEntryPortal from "@rilldata/web-common/features/file-explorer/NavEntryPortal.svelte";
  import { navEntryDragDropStore } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { PROTECTED_DIRECTORIES } from "@rilldata/web-common/features/file-explorer/protected-paths";
  import { isCurrentActivePage } from "@rilldata/web-common/features/file-explorer/utils";
  import { editorMode } from "@rilldata/web-common/layout/editor-mode-store";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { get } from "svelte/store";
  import { eventBus } from "../../lib/event-bus/event-bus";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { fileArtifacts } from "../entity-management/file-artifacts";
  import NavDirectory from "./NavDirectory.svelte";
  import {
    findDirectory,
    transformFileList,
    type Directory,
  } from "./transform-file-list";
  import QuickView from "@rilldata/web-common/features/resource-graph/quick-view/QuickView.svelte";

  const VISUAL_KINDS = new Set<ResourceKind>([
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
  ]);

  function filterTreeToVisualKinds(tree: Directory): Directory {
    const files = tree.files.filter((fileName) => {
      const filePath =
        tree.path === "/" ? `/${fileName}` : `${tree.path}/${fileName}`;
      const artifact = fileArtifacts.getFileArtifact(filePath);
      const kind = get(artifact.resourceName)?.kind as ResourceKind | undefined;
      return kind !== undefined && VISUAL_KINDS.has(kind);
    });

    const directories = tree.directories
      .map(filterTreeToVisualKinds)
      .filter((d) => d.files.length > 0 || d.directories.length > 0);

    return { ...tree, files, directories };
  }

  export let hasUnsaved: boolean;

  const runtimeClient = useRuntimeClient();

  $: getFileTree = createRuntimeServiceListFiles(
    runtimeClient,
    {},
    {
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
    },
  );

  $: ({ data: fileTree } = $getFileTree);

  $: displayTree =
    fileTree && $editorMode === "visual"
      ? filterTreeToVisualKinds(fileTree)
      : fileTree;

  let showRenameModelModal = false;
  let renameFilePath: string;
  let renameIsDir: boolean;

  function onRename(filePath: string, isDir: boolean) {
    showRenameModelModal = true;
    renameFilePath = filePath;
    renameIsDir = isDir;
  }

  async function onDuplicate(filePath: string, isDir: boolean) {
    if (isDir) {
      throw new Error("Copying directories is not supported");
    }

    try {
      const newFilePath = await duplicateFileArtifact(runtimeClient, filePath);
      await goto(`/files${newFilePath}`);
    } catch {
      eventBus.emit("notification", {
        message: `Failed to copy ${filePath}`,
      });
    }
  }

  let forceDeletePath: string;
  let showForceDelete = false;

  async function onDelete(filePath: string, isDir: boolean) {
    if (!$getFileTree.data) return;

    if (isDir) {
      const dir = findDirectory($getFileTree.data, filePath);
      if (dir?.directories?.length || dir?.files?.length) {
        forceDeletePath = filePath;
        showForceDelete = true;
        return;
      }
    }
    await deleteFileArtifact(runtimeClient, filePath);
    if (isCurrentActivePage(filePath, isDir)) {
      await goto("/");
    }
  }

  async function onForceDelete() {
    await deleteFileArtifact(runtimeClient, forceDeletePath, true);
    // onForceDelete is only called on folders, so isDir is always true
    if (isCurrentActivePage(forceDeletePath, true)) {
      await goto("/");
    }
  }

  const { dragData, position } = navEntryDragDropStore;

  async function handleDropSuccess(fromPath: string, toDir: string) {
    const isCurrentFile =
      $page.params.file && // handle case when user is on home page
      removeLeadingSlash(fromPath) === removeLeadingSlash($page.params.file);
    const [, srcFile] = splitFolderAndFileName(fromPath);
    const newFilePath = `${toDir === "/" ? toDir : toDir + "/"}${srcFile}`;

    if (fromPath !== newFilePath) {
      const newTopLevelPath = getTopLevelFolder(newFilePath);
      if (PROTECTED_DIRECTORIES.includes(newTopLevelPath)) {
        eventBus.emit("notification", {
          message: "cannot move to restricted directories",
        });
        return;
      }
      await renameFileArtifact(runtimeClient, fromPath, newFilePath);

      if (isCurrentFile) {
        await goto(`/files${newFilePath}`);
      }
    }
  }

  async function saveAll(e: KeyboardEvent) {
    if (e.code === "KeyS" && e.metaKey && e.altKey) {
      e.preventDefault();
      await fileArtifacts.saveAll();
    }
  }
</script>

<svelte:window
  onbeforeunload={(event) => {
    if (hasUnsaved) {
      event.preventDefault();
      return confirm(
        "Are you sure you want to leave? Unsaved changes will be lost.",
      );
    }
  }}
  onmousemove={(e) => navEntryDragDropStore.onMouseMove(e)}
  onmouseup={(e) => navEntryDragDropStore.onMouseUp(e, handleDropSuccess)}
  onkeydown={saveAll}
/>

<!-- File tree -->
<ul class="flex flex-col w-full items-start justify-start overflow-auto">
  {#if displayTree}
    <NavDirectory
      directory={displayTree}
      {onRename}
      {onDuplicate}
      {onDelete}
      onMouseDown={(e, dragData) =>
        navEntryDragDropStore.onMouseDown(e, dragData)}
    />
  {/if}
</ul>

{#if showRenameModelModal}
  <RenameAssetModal
    closeModal={() => (showRenameModelModal = false)}
    filePath={renameFilePath}
    isDir={renameIsDir}
  />
{/if}

{#if $dragData}
  <NavEntryPortal position={$position} dragData={$dragData} />
{/if}

<ForceDeleteConfirmation bind:open={showForceDelete} onDelete={onForceDelete} />

<QuickView />
