<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/column-profile/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import { V1ResourceName } from "@rilldata/web-common/runtime-client";
  import { Readable } from "svelte/store";
  import File from "../../components/icons/File.svelte";
  import NavigationMenuSeparator from "../../layout/navigation/NavigationMenuSeparator.svelte";
  import DashboardMenuItems from "../dashboards/DashboardMenuItems.svelte";
  import { fileArtifacts } from "../entity-management/file-artifacts";
  import { resourceIconMapping } from "../entity-management/resource-icon-mapping";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import ModelMenuItems from "../models/navigation/ModelMenuItems.svelte";
  import SourceMenuItems from "../sources/navigation/SourceMenuItems.svelte";

  export let filePath: string;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string) => void;

  let contextMenuOpen = false;

  $: fileName = filePath.split("/").pop();
  $: directoryLevel = getDirectoryLevelFromPath(filePath);
  $: isCurrentFile = filePath === $page.params.file;
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  let name: Readable<V1ResourceName | undefined>;
  $: name = fileArtifact.name;
  $: resourceKind = $name?.kind;

  async function navigate(filePath: string) {
    await goto(`/files/${filePath}`);
  }

  function getDirectoryLevelFromPath(path: string) {
    // Root level is 0; each "/" in the path represents a level deeper
    return path === "" ? 0 : path.split("/").length;
  }
</script>

<button
  class="w-full group pr-2 text-left flex justify-between gap-x-1 items-center text-gray-900 font-medium hover:text-gray-900 hover:bg-slate-100 {isCurrentFile
    ? 'bg-slate-100 text-gray-900'
    : ''}"
  on:click={() => navigate(filePath)}
  style:padding-left="{8 + (directoryLevel - 1) * 14}px"
>
  <svelte:component
    this={resourceKind ? resourceIconMapping[resourceKind] : File}
    className="text-gray-400"
    size="14px"
  />
  <span class="truncate w-full">{fileName}</span>
  <DropdownMenu.Root bind:open={contextMenuOpen}>
    <DropdownMenu.Trigger asChild let:builder>
      <ContextButton
        builders={[builder]}
        id="more-actions-{filePath}"
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
      {#if resourceKind}
        {#if resourceKind === ResourceKind.Source}
          <SourceMenuItems {filePath} />
          <NavigationMenuSeparator />
        {:else if resourceKind === ResourceKind.Model}
          <ModelMenuItems {filePath} />
          <NavigationMenuSeparator />
        {:else if resourceKind === ResourceKind.Dashboard}
          <DashboardMenuItems {filePath} />
          <NavigationMenuSeparator />
        {/if}
      {/if}
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
