<script lang="ts">
  import ContextButton from "@rilldata/web-common/components/column-profile/ContextButton.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import { directoryState } from "@rilldata/web-common/features/file-explorer/directory-store";
  import { NavDragData } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { getPaddingFromPath } from "@rilldata/web-common/features/file-explorer/nav-tree-spacing";
  import { Directory } from "@rilldata/web-common/features/file-explorer/transform-file-list";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  export let dir: Directory;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string) => void;
  export let onMouseDown: (e: MouseEvent, dragData: NavDragData) => void;
  export let onMouseUp: (e: MouseEvent, dragData: NavDragData) => void;

  let contextMenuOpen = false;
  $: expanded = $directoryState[dir.path];

  $: padding = getPaddingFromPath(dir.path);

  $: id = `${dir.path}-nav-entry`;

  function toggleDirectory(directory: Directory): void {
    directoryState.toggle(directory.path);
  }
</script>

<button
  class="pr-2 w-full text-left flex justify-between group gap-x-1 items-center text-gray-900 font-medium hover:text-gray-900 hover:bg-slate-100"
  {id}
  on:click={() => toggleDirectory(dir)}
  on:mousedown={(e) => onMouseDown(e, { id, filePath: dir.path, isDir: true })}
  on:mouseup={(e) => onMouseUp(e, { id, filePath: dir.path, isDir: true })}
  style:padding-left="{padding}px"
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
