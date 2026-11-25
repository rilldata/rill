<script lang="ts">
  import { page } from "$app/stores";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/actions/RenameAssetModal.svelte";
  import {
    navigateToFile,
    navigateToHome,
  } from "@rilldata/web-common/layout/navigation/editor-routing";
  import {
    deleteFileArtifact,
    duplicateFileArtifact,
    renameFileArtifact,
  } from "@rilldata/web-common/features/entity-management/actions/actions.ts";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    getTopLevelFolder,
    splitFolderAndFileName,
  } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import DeleteConfirmationDialog from "@rilldata/web-common/features/file-explorer/DeleteConfirmationDialog.svelte";
  import NavEntryPortal from "@rilldata/web-common/features/file-explorer/NavEntryPortal.svelte";
  import { navEntryDragDropStore } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { isCurrentActivePage } from "@rilldata/web-common/features/file-explorer/utils";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { eventBus } from "../../lib/event-bus/event-bus";
  import { fileArtifacts } from "../entity-management/file-artifacts";
  import { fetchResources } from "../entity-management/resource-selectors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { ChevronsDownUp, ChevronsUpDown } from "lucide-svelte";
  import { get } from "svelte/store";
  import NavDirectory from "./NavDirectory.svelte";
  import {
    type DeleteTarget,
    getFileDeleteTarget,
    getFolderDeleteTarget,
  } from "./delete-target";
  import { directoryState, isDirectoryExpanded } from "./directory-store";
  import { useProjectScopeId } from "./project-scope";
  import {
    collectDirectoryPaths,
    transformFileList,
  } from "./transform-file-list";
  import QuickView from "@rilldata/web-common/features/resource-graph/quick-view/QuickView.svelte";
  import {
    isPinned,
    isProtectedDirectory,
    isManaged,
  } from "@rilldata/web-common/features/entity-management/actions/protected-files.ts";

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

  $: projectTitleQuery = useProjectTitle(runtimeClient);
  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Rill Project";

  // Scope the persisted expand/collapse state per project.
  const projectScopeId = useProjectScopeId(runtimeClient.instanceId);
  $: if ($projectScopeId) directoryState.setProjectScope($projectScopeId);

  $: directoryPaths = fileTree ? collectDirectoryPaths(fileTree) : [];
  $: allDirectoriesCollapsed =
    directoryPaths.length > 0 &&
    directoryPaths.every((path) => !isDirectoryExpanded($directoryState, path));
  $: toggleAllDirectoriesLabel = allDirectoriesCollapsed
    ? "Expand all folders"
    : "Collapse all folders";

  function toggleAllDirectories() {
    if (!fileTree) return;
    if (allDirectoriesCollapsed) {
      directoryState.expandAll(directoryPaths);
    } else {
      directoryState.collapseAll(directoryPaths);
    }
  }

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
      await navigateToFile(newFilePath);
    } catch {
      eventBus.emit("notification", {
        message: `Failed to copy ${filePath}`,
      });
    }
  }

  // The delete awaiting confirmation. It stays set after the dialog closes, so the dialog keeps its text while it animates out.
  let deleteTarget: DeleteTarget | undefined;
  let showDeleteConfirmation = false;

  // Deleting a file or a non-empty folder loses work, so it asks for confirmation first.
  // An empty folder holds nothing that could be lost and is deleted right away.
  async function onDelete(filePath: string, isDir: boolean) {
    if (!$getFileTree.data) return;

    let target: DeleteTarget | undefined;
    if (isDir) {
      target = getFolderDeleteTarget($getFileTree.data, filePath);
    } else {
      const resourceName = get(
        fileArtifacts.getFileArtifact(filePath).resourceName,
      );
      // The dialog lists the files that depend on this one; if the lookup fails, it opens without that list.
      const resources = resourceName
        ? await fetchResources(queryClient, runtimeClient).catch(() => [])
        : [];
      target = getFileDeleteTarget(filePath, resourceName, resources);
    }
    if (!target) {
      await deleteEntry(filePath, isDir, false);
      return;
    }

    deleteTarget = target;
    showDeleteConfirmation = true;
  }

  async function onConfirmDelete(target: DeleteTarget) {
    const isDir = target.type === "folder";
    // Only non-empty folders are confirmed, and the runtime refuses to delete those without force.
    await deleteEntry(target.path, isDir, isDir);
  }

  async function deleteEntry(filePath: string, isDir: boolean, force: boolean) {
    await deleteFileArtifact(runtimeClient, filePath, force);
    if (isCurrentActivePage(filePath, isDir)) {
      await navigateToHome();
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
      if (isProtectedDirectory(newTopLevelPath)) {
        eventBus.emit("notification", {
          message: "cannot move to restricted directories",
        });
        return;
      }
      if (isPinned(newFilePath) || isManaged(newFilePath)) {
        eventBus.emit("notification", {
          message: "cannot rename to a restricted file",
        });
        return;
      }
      await renameFileArtifact(runtimeClient, fromPath, newFilePath);

      if (isCurrentFile) {
        await navigateToFile(newFilePath);
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

<!-- Project header with a collapse-all control -->
<div class="project-header">
  <h3 class="truncate" title={projectTitle}>{projectTitle}</h3>
  {#if directoryPaths.length}
    <button
      type="button"
      class="collapse-all"
      aria-label={toggleAllDirectoriesLabel}
      title={toggleAllDirectoriesLabel}
      onclick={toggleAllDirectories}
    >
      {#if allDirectoriesCollapsed}
        <ChevronsUpDown size="14px" />
      {:else}
        <ChevronsDownUp size="14px" />
      {/if}
    </button>
  {/if}
</div>

<!-- File tree -->
<ul class="flex flex-col w-full items-start justify-start overflow-auto">
  {#if fileTree}
    <NavDirectory
      directory={fileTree}
      {onRename}
      {onDuplicate}
      {onDelete}
      onMouseDown={(e, dragData) =>
        navEntryDragDropStore.onMouseDown(e, dragData)}
    />
  {:else if $getFileTree.isLoading}
    <div class="flex flex-col gap-y-1.5 w-full px-2 py-2">
      {#each [0.7, 0.5, 0.8, 0.6, 0.55, 0.65] as width}
        <div
          class="h-5 bg-gray-200 animate-pulse rounded"
          style:width="{width * 100}%"
        ></div>
      {/each}
    </div>
  {:else if $getFileTree.isError}
    <div class="px-2 py-3 text-xs text-fg-muted">Failed to load files</div>
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

<DeleteConfirmationDialog
  bind:open={showDeleteConfirmation}
  target={deleteTarget}
  onConfirm={onConfirmDelete}
/>

<QuickView />

<style lang="postcss">
  .project-header {
    @apply sticky top-0 z-10 bg-surface-base;
    @apply flex items-center justify-between gap-x-1;
    @apply h-7 w-full pl-2 pr-1.5;
  }

  h3 {
    @apply truncate font-semibold text-[10px] uppercase text-fg-muted;
  }

  .collapse-all {
    @apply flex flex-none items-center justify-center;
    @apply size-5 rounded text-fg-secondary;
  }

  .collapse-all:hover,
  .collapse-all:focus-visible {
    @apply bg-surface-hover text-fg-primary;
  }
</style>
