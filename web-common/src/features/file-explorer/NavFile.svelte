<script lang="ts">
  import { page } from "$app/stores";
  import ContextButton from "@rilldata/web-common/components/button/ContextButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Alert from "@rilldata/web-common/components/icons/Alert.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import MoreHorizontal from "@rilldata/web-common/components/icons/MoreHorizontal.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { NavDragData } from "@rilldata/web-common/features/file-explorer/nav-entry-drag-drop-store";
  import { getPaddingFromPath } from "@rilldata/web-common/features/file-explorer/nav-tree-spacing";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import NavigationMenuItem from "@rilldata/web-common/layout/navigation/NavigationMenuItem.svelte";
  import NavigationMenuSeparator from "@rilldata/web-common/layout/navigation/NavigationMenuSeparator.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
    ResourceKindToScreenMap,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { Save } from "lucide-svelte";
  import type { Readable } from "svelte/store";
  import CopyIcon from "../../components/icons/CopyIcon.svelte";
  import CanvasMenuItems from "../canvas/CanvasMenuItems.svelte";
  import { fileArtifacts } from "../entity-management/file-artifacts";
  import { getTopLevelFolder } from "../entity-management/file-path-utils";
  import { getIconComponent } from "../entity-management/resource-icon-mapping";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import ExploreMenuItems from "../explores/ExploreMenuItems.svelte";
  import MetricsViewMenuItems from "../metrics-views/MetricsViewMenuItems.svelte";
import ModelMenuItems from "../models/navigation/ModelMenuItems.svelte";
import { PROTECTED_DIRECTORIES, PROTECTED_FILES } from "./protected-paths";

  export let filePath: string;
  export let onRename: (filePath: string, isDir: boolean) => void;
  export let onDuplicate: (filePath: string, isDir: boolean) => void;
  export let onDelete: (filePath: string, isDir: boolean) => void;
  export let onMouseDown: (e: MouseEvent, dragData: NavDragData) => void;

  let contextMenuOpen = false;
  let resourceName: Readable<V1ResourceName | undefined>;

  $: id = `${filePath}-nav-link`;
  $: fileName = filePath.split("/").pop();
  $: isCurrentFile =
    removeLeadingSlash(filePath) ===
    removeLeadingSlash($page.params.file ?? "");
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  $: ({
    resourceName,
    hasUnsavedChanges,
    saveLocalContent,
    inferredResourceKind,
    saveState: { saving, error },
  } = fileArtifact);

  $: ({ instanceId } = $runtime);

  // Normalize Source to Model (Source is deprecated)
  $: rawResourceKind = ($resourceName?.kind ??
    $inferredResourceKind) as ResourceKind;
  $: resourceKind = rawResourceKind === ResourceKind.Source ? ResourceKind.Model : rawResourceKind;
  $: padding = getPaddingFromPath(filePath);
  $: topLevelFolder = getTopLevelFolder(filePath);
  $: isProtectedDirectory = PROTECTED_DIRECTORIES.includes(topLevelFolder);
  $: isDotFile = fileName && fileName.startsWith(".");
  $: isProtectedFile = PROTECTED_FILES.includes(filePath);

  $: hasErrors = fileArtifact.getHasErrors(queryClient, instanceId);

  function fireTelemetry() {
    const previousScreenName = getScreenNameFromPage();
    behaviourEvent
      ?.fireNavigationEvent(
        filePath,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        previousScreenName,
        ResourceKindToScreenMap[resourceKind] ?? MetricsEventScreenName.Unknown,
      )
      .catch(console.error);
  }

  function handleMouseDown(e: MouseEvent) {
    if (PROTECTED_FILES.includes(filePath)) return;
    onMouseDown(e, { id, filePath, isDir: false, kind: resourceKind });
  }
</script>

<li
  aria-label="{filePath} Nav Entry"
  class="w-full text-left pr-2 h-6 group flex justify-between gap-x-1 items-center hover:bg-surface-hover"
  class:bg-surface-active={isCurrentFile}
  class:opacity-50={$hasUnsavedChanges || $saving}
>
  <a
    class="w-full truncate flex items-center gap-x-1 font-medium {isProtectedDirectory ||
    isDotFile
      ? 'hover:text-fg-secondary text-fg-muted '
      : 'text-fg-primary hover:text-fg-primary'}"
    href="/files{filePath}"
    {id}
    class:italic={$hasUnsavedChanges || $saving}
    on:click={fireTelemetry}
    on:mousedown={handleMouseDown}
    style:padding-left="{padding}px"
  >
    <div class="flex-none">
      {#if $saving}
        <LoadingSpinner size="14px" />
      {:else if $error}
        <Alert size="14px" color="red" />
      {:else}
        <svelte:component
          this={getIconComponent(resourceKind, filePath)}
          size="14px"
        />
      {/if}
    </div>
    <span class="truncate w-full" class:text-red-600={$hasErrors}>
      {fileName}
    </span>
  </a>
  {#if !isProtectedDirectory && !isProtectedFile}
    <DropdownMenu.Root bind:open={contextMenuOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <ContextButton
          builders={[builder]}
          id="more-actions-{filePath}"
          label="{filePath} actions menu trigger"
          suppressTooltip={contextMenuOpen}
          tooltipText="More actions"
        >
          <MoreHorizontal />
        </ContextButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        align="start"
        class="min-w-60"
        side="right"
        sideOffset={16}
      >
        {#if $hasUnsavedChanges}
          <NavigationMenuItem on:click={saveLocalContent}>
            <Save slot="icon" size="12px" />
            Save file
          </NavigationMenuItem>
        {/if}
        <NavigationMenuItem on:click={() => onRename(filePath, false)}>
          <EditIcon slot="icon" />
          Rename
        </NavigationMenuItem>
        <NavigationMenuItem on:click={() => onDuplicate(filePath, false)}>
          <CopyIcon slot="icon" />
          Duplicate
        </NavigationMenuItem>
        {#if resourceKind}
          {#if resourceKind === ResourceKind.Model}
            <ModelMenuItems {filePath} />
          {:else if resourceKind === ResourceKind.MetricsView}
            <MetricsViewMenuItems {filePath} />
          {:else if resourceKind === ResourceKind.Explore}
            <ExploreMenuItems {filePath} />
          {:else if resourceKind === ResourceKind.Canvas}
            <CanvasMenuItems {filePath} />
          {/if}
        {/if}
        <NavigationMenuSeparator />
        <NavigationMenuItem on:click={() => onDelete(filePath, false)}>
          <Trash slot="icon" />
          Delete
        </NavigationMenuItem>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  {/if}
</li>
