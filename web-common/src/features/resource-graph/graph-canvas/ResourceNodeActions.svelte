<script context="module" lang="ts">
  // Track the currently open dropdown so right-clicking another node
  // closes the previous one instead of stacking.
  let closeActive: (() => void) | null = null;

  export function closeActiveDropdown() {
    closeActive?.();
  }
</script>

<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    RefreshCw,
    ExternalLink,
    Info,
    GitBranch,
    Copy,
  } from "lucide-svelte";
  import {
    createRuntimeServiceCreateTriggerMutation,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { goto } from "$app/navigation";
  import type { ResourceNodeData } from "../shared/types";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { tokenForKind } from "../navigation/seed-parser";
  import { getGraphNavigation } from "../shared/graph-navigation-context";
  import { onDestroy } from "svelte";
  import FullRefreshConfirmDialog from "../shared/FullRefreshConfirmDialog.svelte";
  import ResourceSpecDialog from "../shared/ResourceSpecDialog.svelte";


  const graphNav = getGraphNavigation();

  // Portal the trigger to body so it escapes SvelteFlow transforms
  function portalToBody(node: HTMLElement) {
    document.body.appendChild(node);
    return {
      destroy() {
        node.remove();
      },
    };
  }

  export let data: ResourceNodeData;

  let isOpen = false;
  let fullRefreshConfirmOpen = false;
  let specDialogOpen = false;
  let menuX = 0;
  let menuY = 0;

  function close() {
    isOpen = false;
    if (closeActive === close) closeActive = null;
  }

  export function open(e?: MouseEvent) {
    if (closeActive && closeActive !== close) closeActive();
    closeActive = close;
    if (e) {
      menuX = e.clientX;
      menuY = e.clientY;
    }
    isOpen = true;
  }

  function handleOpenChange(next: boolean) {
    if (!next) close();
  }

  onDestroy(() => {
    if (closeActive === close) closeActive = null;
  });

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  $: resource = data?.resource;
  $: kind = data?.kind;
  $: resourceName = resource?.meta?.name?.name ?? "";
  $: filePath = resource?.meta?.filePaths?.[0];
  $: canOpenFile = !!filePath && (!!graphNav?.openFile || !graphNav);
  $: canRefresh =
    (kind === ResourceKind.Model || kind === ResourceKind.Source) &&
    !!resourceName;
  $: isIncremental = !!resource?.model?.spec?.incremental;
  $: reconcileError = resource?.meta?.reconcileError ?? "";

  const triggerMutation =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  // IMPORTANT: coerceResourceKind classifies root models as ResourceKind.Source,
  // but the trigger API still expects the "models" mutation key because the
  // underlying runtime resource is always a Model. Do not change this to "sources".
  function refreshModel(full: boolean) {
    if (!resourceName) return;
    $triggerMutation.mutate(
      {
        models: [{ model: resourceName, full }],
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries({
            queryKey: getRuntimeServiceListResourcesQueryKey(
              runtimeClient.instanceId,
              undefined,
            ),
          });
        },
        onError: (err) => {
          console.error(`Failed to refresh ${resourceName}:`, err);
          eventBus.emit("notification", {
            message: `Failed to refresh ${resourceName}`,
            type: "error",
          });
        },
      },
    );
  }

  function handleIncrementalRefresh() {
    close();
    refreshModel(false);
  }

  function handleFullRefreshClick() {
    close();
    fullRefreshConfirmOpen = true;
  }

  function confirmFullRefresh() {
    fullRefreshConfirmOpen = false;
    refreshModel(true);
  }

  function openFile() {
    if (!filePath) return;
    close();
    if (graphNav?.openFile) {
      graphNav.openFile(filePath);
      return;
    }
    try {
      const prefs = JSON.parse(localStorage.getItem(filePath) || "{}");
      localStorage.setItem(
        filePath,
        JSON.stringify({ ...prefs, view: "code" }),
      );
    } catch (error) {
      console.warn(`Failed to save file view preference:`, error);
    }
    goto(`/files${filePath}`);
  }

  function handleViewSpec() {
    close();
    specDialogOpen = true;
  }

  function handleCopyError() {
    close();
    navigator.clipboard.writeText(reconcileError).catch((err) => {
      console.error("Failed to copy error:", err);
    });
  }

  function viewNodeTree() {
    close();
    const kindToken = tokenForKind(kind);
    if (graphNav?.viewLineage) {
      graphNav.viewLineage(kindToken, resourceName);
      return;
    }
    const params = new URLSearchParams();
    if (kindToken) params.set("kind", kindToken);
    if (resourceName) params.set("resource", resourceName);
    goto(`/graph?${params.toString()}`);
  }
</script>

<!--
  The trigger+anchor must be portaled to <body> so it escapes SvelteFlow's
  CSS transform. The DropdownMenu.Content already portals via bits-ui,
  but it positions relative to the trigger's bounding rect — if the trigger
  is inside a transformed container, getBoundingClientRect returns wrong coords.
-->
<div
  use:portalToBody
  class="trigger-anchor"
  style="left: {menuX}px; top: {menuY}px;"
>
  <DropdownMenu.Root open={isOpen} onOpenChange={handleOpenChange}>
    <DropdownMenu.Trigger class="trigger-btn" />
    <DropdownMenu.Content side="bottom" align="start" sideOffset={0}>
      <DropdownMenu.Item onclick={handleViewSpec}>
        <div class="flex items-center gap-x-2">
          <Info size="12px" />
          <span>Describe</span>
        </div>
      </DropdownMenu.Item>
      <DropdownMenu.Item onclick={viewNodeTree}>
        <div class="flex items-center gap-x-2">
          <GitBranch size="12px" />
          <span>View Lineage</span>
        </div>
      </DropdownMenu.Item>
      {#if canOpenFile}
        <DropdownMenu.Item onclick={openFile}>
          <div class="flex items-center gap-x-2">
            <ExternalLink size="12px" />
            <span>Go to Resource</span>
          </div>
        </DropdownMenu.Item>
      {/if}
      {#if reconcileError}
        <DropdownMenu.Item onclick={handleCopyError}>
          <div class="flex items-center gap-x-2">
            <Copy size="12px" />
            <span>Copy Error Message</span>
          </div>
        </DropdownMenu.Item>
      {/if}
      {#if canRefresh}
        <DropdownMenu.Separator />
        <DropdownMenu.Item onclick={handleFullRefreshClick}>
          <div class="flex items-center gap-x-2">
            <RefreshCw size="12px" />
            <span>Full Refresh</span>
          </div>
        </DropdownMenu.Item>
        {#if isIncremental}
          <DropdownMenu.Item onclick={handleIncrementalRefresh}>
            <div class="flex items-center gap-x-2">
              <RefreshCw size="12px" />
              <span>Incremental Refresh</span>
            </div>
          </DropdownMenu.Item>
        {/if}
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>

<FullRefreshConfirmDialog
  bind:open={fullRefreshConfirmOpen}
  {resourceName}
  onConfirm={confirmFullRefresh}
/>

<ResourceSpecDialog
  bind:open={specDialogOpen}
  {resourceName}
  {kind}
  {resource}
/>

<style lang="postcss">
  .trigger-anchor {
    position: fixed;
    width: 1px;
    height: 1px;
    pointer-events: none;
    z-index: 50;
  }
  .trigger-anchor :global(.trigger-btn) {
    width: 1px;
    height: 1px;
    padding: 0;
    margin: 0;
    border: none;
    background: transparent;
    pointer-events: none;
    opacity: 0;
  }
</style>
