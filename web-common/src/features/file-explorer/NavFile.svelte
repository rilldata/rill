<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/column-profile/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import { NavDragData } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { getPaddingFromPath } from "@rilldata/web-common/features/file-explorer/nav-tree-spacing";
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
  import { getTopLevelFolder } from "../sources/extract-file-name";
  import SourceMenuItems from "../sources/navigation/SourceMenuItems.svelte";
  import { PROTECTED_DIRECTORIES, PROTECTED_FILES } from "./protected-paths";

  export let filePath: string;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string) => void;
  export let onGenerateChart: (data: {
    table?: string;
    connector?: string;
    metricsView?: string;
  }) => void;
  export let onMouseDown: (e: MouseEvent, dragData: NavDragData) => void;
  export let onMouseUp: (e: MouseEvent, dragData: NavDragData) => void;

  let contextMenuOpen = false;

  $: id = `${filePath}-nav-entry`;
  $: fileName = filePath.split("/").pop();
  $: isCurrentFile = filePath === $page.params.file;
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  let name: Readable<V1ResourceName | undefined>;
  $: name = fileArtifact.name;
  $: resourceKind = $name?.kind as ResourceKind;
  $: padding = getPaddingFromPath(filePath);
  $: topLevelFolder = getTopLevelFolder(filePath);
  $: isProtectedDirectory = PROTECTED_DIRECTORIES.includes(topLevelFolder);
  $: isProtectedFile = PROTECTED_FILES.includes(filePath);

  async function navigate(filePath: string) {
    await goto(`/files/${filePath}`);
  }

  function handleMouseDown(e: MouseEvent) {
    if (fileName === "rill.yaml") return;
    onMouseDown(e, { id, filePath, isDir: false, kind: resourceKind });
  }
</script>

<button
  aria-label="{fileName} Nav Entry"
  class="w-full h-6 group pr-2 text-left flex justify-between gap-x-1 items-center
  {isProtectedDirectory ? 'text-gray-500' : 'text-gray-900 hover:text-gray-900'}
  {isCurrentFile ? 'bg-slate-100' : ''} 
  font-medium hover:bg-slate-100"
  {id}
  on:click={() => navigate(filePath)}
  on:mousedown={handleMouseDown}
  on:mouseup={(e) =>
    onMouseUp(e, { id, filePath, isDir: false, kind: resourceKind })}
  style:padding-left="{padding}px"
>
  <svelte:component
    this={resourceKind ? resourceIconMapping[resourceKind] : File}
    className="text-gray-400"
    size="14px"
  />
  <span class="truncate w-full">{fileName}</span>
  {#if !isProtectedDirectory && !isProtectedFile}
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
            <SourceMenuItems
              {filePath}
              on:generate-chart={({ detail }) => onGenerateChart(detail)}
            />
            <NavigationMenuSeparator />
          {:else if resourceKind === ResourceKind.Model}
            <ModelMenuItems
              {filePath}
              on:generate-chart={({ detail }) => onGenerateChart(detail)}
            />
            <NavigationMenuSeparator />
          {:else if resourceKind === ResourceKind.MetricsView}
            <DashboardMenuItems
              {filePath}
              on:generate-chart={({ detail }) => onGenerateChart(detail)}
            />
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
  {/if}
</button>
