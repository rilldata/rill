<script lang="ts">
  import ContextButton from "@rilldata/web-common/components/column-profile/ContextButton.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import { Directory } from "@rilldata/web-common/features/file-explorer/transform-file-list";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  export let dir: Directory;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string) => void;

  let contextMenuOpen = false;
  $: expanded = $directoryState[dir.path];

  // Root level is 0; each "/" in the path represents a level deeper
  $: dirLevel = dir.path === "" ? 0 : dir.path.split("/").length;

  function toggleDirectory(directory: Directory): void {
    directoryState.toggle(directory.path);
  }
</script>

<button
  class="pr-2 w-full text-left flex justify-between group gap-x-1 items-center text-gray-900 font-medium hover:text-gray-900 hover:bg-slate-100"
  on:click={() => toggleDirectory(dir)}
  style:padding-left="{8 + (dirLevel - 1) * 14}px"
>
  <CaretDownIcon
    className="text-gray-400 {expanded ? '' : 'transform -rotate-90'}"
  />
  <span class="truncate w-full">{dir.name}</span>
  <DropdownMenu.Root bind:open={contextMenuOpen}>
    <DropdownMenu.Trigger asChild let:builder>
      <ContextButton
        builders={[builder]}
        id="more-actions-{dir.path}"
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
