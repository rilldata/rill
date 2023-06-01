<script lang="ts">
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { directoryState } from "./directory-store";
  import NavFile from "./NavFile.svelte";
  import type { Directory } from "./transform-file-list";

  export let directory: Directory;

  function toggleDirectory(directory: Directory): void {
    directoryState.toggle(directory.name);
  }

  function getLeftPaddingFromDirectoryPath(path: string) {
    const dirLevel = path.split("/").length;
    console.log(path, "dirLevel", dirLevel);
    return 4 + (1 - dirLevel) * 2;
  }

  $: console.log("directories", directory?.directories);
</script>

{#if directory?.directories}
  {#each directory.directories as dir}
    {@const expanded = $directoryState[dir.name]}
    <button
      class="pl-{getLeftPaddingFromDirectoryPath(
        dir.path
      )} pr-2 py-0.5 w-full flex gap-x-1 items-center text-gray-500 hover:text-gray-900 hover:bg-gray-100"
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
  <NavFile filePath={directory.path + "/" + file} />
{/each}
