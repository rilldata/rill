<script lang="ts">
  import ContextButton from "@rilldata/web-common/components/column-profile/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import NavFile from "./NavFile.svelte";
  import { directoryState } from "./directory-store";
  import type { Directory } from "./transform-file-list";

  export let directory: Directory;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string) => void;

  function toggleDirectory(directory: Directory): void {
    directoryState.toggle(directory.path);
  }

  function getDirectoryLevelFromPath(path: string) {
    // Root level is 0; each "/" in the path represents a level deeper
    return path === "" ? 0 : path.split("/").length;
  }

  let contextMenuOpen = false;
</script>

{#if directory?.directories}
  {#each directory.directories as dir}
    {@const expanded = $directoryState[dir.path]}
    {@const directoryLevel = getDirectoryLevelFromPath(dir.path)}
    <button
      style:padding-left="{8 + (directoryLevel - 1) * 14}px"
      class="pr-2 w-full text-left flex justify-between group gap-x-1 items-center text-gray-900 font-medium hover:text-gray-900 hover:bg-slate-100"
      on:click={() => toggleDirectory(dir)}
    >
      <CaretDownIcon
        className="text-gray-400 {expanded ? '' : 'transform -rotate-90'}"
      />
      <span class="truncate w-full">{dir.name}</span>
      <DropdownMenu.Root bind:open={contextMenuOpen}>
        <DropdownMenu.Trigger asChild let:builder>
          <ContextButton
            builders={[builder]}
            id="more-actions-{dir.name}"
            label="{dir.name} actions menu trigger"
            suppressTooltip={contextMenuOpen}
            tooltipText="More actions"
          >
            <MoreHorizontal />
          </ContextButton>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content
          align="start"
          class="border-none bg-gray-800 text-white min-w-60"
          side="right"
          sideOffset={16}
        >
          <NavigationMenuItem on:click={() => onRename(dir.path, true)}>
            <EditIcon slot="icon" />
            Rename...
          </NavigationMenuItem>
          <NavigationMenuItem on:click={() => onDelete(dir.path)}>
            <Cancel slot="icon" />
            Delete
          </NavigationMenuItem>
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </button>

    {#if expanded}
      <!-- Recursive call to display subdirectories -->
      <svelte:self directory={dir} {onRename} {onDelete} />
    {/if}
  {/each}
{/if}

{#each directory.files as file}
  {@const filePath = directory.path ? `${directory.path}/${file}` : file}
  <NavFile {filePath} {onRename} {onDelete} />
{/each}
