<script lang="ts">
  import {
    NavDragData,
    navEntryDragDropStore,
  } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import NavDirectoryEntry from "@rilldata/web-common/features/file-explorer/NavDirectoryEntry.svelte";
  import NavFile from "./NavFile.svelte";
  import { directoryState } from "./directory-store";
  import type { Directory } from "./transform-file-list";

  export let directory: Directory;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string) => void;
  export let onGenerateChart: (data: {
    table?: string;
    connector?: string;
    metricsView?: string;
  }) => void;
  export let onMouseDown: (e: MouseEvent, dragData: NavDragData) => void;
  export let onMouseUp: (e: MouseEvent, dragData: NavDragData) => void;

  $: expanded = $directoryState[directory.path];
  const { dragData, dropFolders } = navEntryDragDropStore;
  $: isDragDropHover =
    $dragData && $dropFolders[$dropFolders.length - 1] === directory.path;
</script>

<div
  class="w-full"
  class:bg-slate-100={isDragDropHover}
  on:mouseenter={() => navEntryDragDropStore.onMouseEnter(directory.path)}
  on:mouseleave={() => navEntryDragDropStore.onMouseLeave()}
  role="directory"
>
  {#if directory.path !== "/"}
    <NavDirectoryEntry
      dir={directory}
      {onDelete}
      {onMouseDown}
      {onMouseUp}
      {onRename}
    />
  {/if}

  {#if expanded}
    {#if directory?.directories}
      {#each directory.directories as dir}
        <!-- Recursive call to display subdirectories -->
        <svelte:self
          directory={dir}
          {onRename}
          {onDelete}
          {onGenerateChart}
          {onMouseDown}
          {onMouseUp}
        />
      {/each}
    {/if}

    {#each directory.files as file}
      {@const filePath =
        directory.path === "/" ? `/${file}` : `${directory.path}/${file}`}
      <NavFile
        {filePath}
        {onRename}
        {onDelete}
        {onGenerateChart}
        {onMouseDown}
        {onMouseUp}
      />
    {/each}
  {/if}
</div>
