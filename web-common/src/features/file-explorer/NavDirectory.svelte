<script lang="ts">
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import { directoryState } from "./directory-store";
  import NavFile from "./NavFile.svelte";
  import type { Directory } from "./transform-file-list";

  export let directory: Directory;

  $: console.log("directories", directory?.directories);

  function toggleDirectory(directory: Directory): void {
    directoryState.toggle(directory.name);
  }

  function getDirectoryLevelFromPath(path: string) {
    // Root level is 0; each "/" in the path represents a level deeper
    return path === "" ? 0 : path.split("/").length;
  }

  function getLeftPaddingForDirectoryLevel(dirLevel: number) {
    return 4 + (1 - dirLevel) * 2;
  }
</script>

{#if directory?.directories}
  {#each directory.directories as dir}
    {@const expanded = $directoryState[dir.name]}
    {@const directoryLevel = getDirectoryLevelFromPath(dir.path)}
    <button
      class="pl-{getLeftPaddingForDirectoryLevel(
        directoryLevel,
      )} pr-2 py-0.5 w-full flex gap-x-1 items-center text-gray-900 font-medium hover:text-gray-900 hover:bg-slate-100"
      on:click={() => toggleDirectory(dir)}
    >
      {#if expanded}
        <CaretDownIcon />
      {:else}
        <CaretDownIcon className="transform -rotate-90" />
      {/if}

      {dir.name}
    </button>

    {#if expanded}
      <!-- Recursive call to display subdirectories -->
      <svelte:self directory={dir} />
    {/if}
  {/each}
{/if}

{#each directory.files as file}
  {@const filePath = directory.path ? `${directory.path}/${file}` : file}
  <NavFile {filePath} />
{/each}
