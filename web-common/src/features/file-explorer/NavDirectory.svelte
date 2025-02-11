<script lang="ts">
  import NavDirectoryEntry from "@rilldata/web-common/features/file-explorer/NavDirectoryEntry.svelte";
  import {
    navEntryDragDropStore,
    type NavDragData,
  } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import NavFile from "./NavFile.svelte";
  import { directoryState } from "./directory-store";
  import type { Directory } from "./transform-file-list";

  export let directory: Directory;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDuplicate: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string, isDir: boolean) => void;
  export let onMouseDown: (e: MouseEvent, dragData: NavDragData) => void;

  $: expanded = $directoryState[directory.path];
  const { dragData, dropDirs } = navEntryDragDropStore;
  $: isDragDropHover =
    $dragData && $dropDirs[$dropDirs.length - 1] === directory.path;
</script>

<ul
  id={`nav-${directory.path}`}
  aria-label={directory.path}
  role="directory"
  class="w-full"
  class:bg-slate-100={isDragDropHover}
  on:mouseenter={() => navEntryDragDropStore.onMouseEnter(directory.path)}
  on:mouseleave={() => navEntryDragDropStore.onMouseLeave()}
  on:contextmenu={() => navEntryDragDropStore.resetDrag()}
>
  {#if directory.path !== "/"}
    <NavDirectoryEntry dir={directory} {onRename} {onDelete} {onMouseDown} />
  {/if}

  {#if expanded}
    {#if directory?.directories}
      {#each directory.directories as dir (dir.path)}
        <!-- Recursive call to display subdirectories -->
        <svelte:self
          directory={dir}
          {onRename}
          {onDuplicate}
          {onDelete}
          {onMouseDown}
        />
      {/each}
    {/if}

    {#each directory.files as file (file)}
      {@const filePath =
        directory.path === "/" ? `/${file}` : `${directory.path}/${file}`}
      <NavFile {filePath} {onRename} {onDuplicate} {onDelete} {onMouseDown} />
    {/each}
  {/if}
</ul>
