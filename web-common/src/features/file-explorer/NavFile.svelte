<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/column-profile/ContextButton.svelte";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import File from "../../components/icons/File.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  export let filePath: string;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string) => void;

  $: fileName = filePath.split("/").pop();
  $: fileLevel = getDirectoryLevelFromPath(filePath);
  $: isCurrentFile = filePath === $page.params.file;

  async function navigate(filePath: string) {
    await goto(`/files/${filePath}`);
  }

  function getDirectoryLevelFromPath(path: string) {
    // Root level is 0; each "/" in the path represents a level deeper
    return path === "" ? 0 : path.split("/").length;
  }

  let contextMenuOpen = false;
</script>

<button
  class="w-full group pr-4 text-left py-1 flex justify-between gap-x-1 items-center text-gray-900 font-medium hover:text-gray-900 hover:bg-slate-100 {isCurrentFile
    ? 'bg-slate-100 text-gray-900'
    : ''}"
  on:click={() => navigate(filePath)}
  style:padding-left="{5 + fileLevel * 14}px"
>
  <File className="shrink-0" size="14px" />
  <span class="truncate w-full">{fileName}</span>
  <DropdownMenu.Root bind:open={contextMenuOpen}>
    <DropdownMenu.Trigger asChild let:builder>
      <ContextButton
        builders={[builder]}
        id="more-actions-{fileName}"
        label="{fileName} actions menu trigger"
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
      <NavigationMenuItem on:click={() => onRename(filePath, false)}>
        <EditIcon slot="icon" />
        Rename...
      </NavigationMenuItem>
      <NavigationMenuItem on:click={() => onDelete(filePath)}>
        <Cancel slot="icon" />
        Delete
      </NavigationMenuItem>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</button>
