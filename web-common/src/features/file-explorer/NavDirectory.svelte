<script lang="ts">
  import { NavDragData } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import NavDirectoryEntry from "@rilldata/web-common/features/file-explorer/NavDirectoryEntry.svelte";
  import NavFile from "./NavFile.svelte";
  import { directoryState } from "./directory-store";
  import type { Directory } from "./transform-file-list";

  export let directory: Directory;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string) => void;
  export let onMouseDown: (e: MouseEvent, dragData: NavDragData) => void;
  export let onMouseUp: (e: MouseEvent, dragData: NavDragData) => void;
</script>

{#if directory?.directories}
  {#each directory.directories as dir}
    {@const expanded = $directoryState[dir.path]}
    <NavDirectoryEntry {dir} {onRename} {onDelete} {onMouseDown} {onMouseUp} />

    {#if expanded}
      <!-- Recursive call to display subdirectories -->
      <svelte:self
        directory={dir}
        {onRename}
        {onDelete}
        {onMouseDown}
        {onMouseUp}
      />
    {/if}
  {/each}
{/if}

{#each directory.files as file}
  {@const filePath = directory.path ? `${directory.path}/${file}` : file}
  <NavFile {filePath} {onRename} {onDelete} {onMouseDown} {onMouseUp} />
{/each}
