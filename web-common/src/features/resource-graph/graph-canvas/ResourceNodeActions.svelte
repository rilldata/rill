<script context="module" lang="ts">
  // Track the currently open dropdown so right-clicking another node
  // closes the previous one instead of stacking.
  let closeActive: (() => void) | null = null;

  export function closeActiveDropdown() {
    closeActive?.();
  }
</script>

<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    ResourceKind,
    displayResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    RefreshCw,
    ExternalLink,
    Info,
    GitBranch,
    Copy,
  } from "lucide-svelte";
  import {
    createRuntimeServiceCreateTriggerMutation,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { goto } from "$app/navigation";
  import type { ResourceNodeData } from "../shared/types";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { tokenForKind } from "../navigation/seed-parser";
  import { getGraphNavigation } from "../shared/graph-navigation-context";
  import { onDestroy } from "svelte";

  const graphNav = getGraphNavigation();

  function portal(node: HTMLElement) {
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

  // Keep module-level tracker in sync when dropdown closes
  $: if (!isOpen && closeActive === close) closeActive = null;

  // Close on outside click
  function handleWindowClick() {
    if (isOpen) close();
  }

  // Close on Escape
  function handleWindowKeydown(e: KeyboardEvent) {
    if (e.key === "Escape" && isOpen) {
      e.stopPropagation();
      close();
    }
  }

  onDestroy(() => {
    if (closeActive === close) closeActive = null;
  });

  const runtimeClient = useRuntimeClient();
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

  // Raw spec content (same pattern as project tables ResourceSpecDialog)
  function getResourceSpec(res: V1Resource | undefined): string {
    if (!res) return "";
    const kindKeys = [
      "source",
      "model",
      "metricsView",
      "explore",
      "theme",
      "component",
      "canvas",
      "api",
      "connector",
      "report",
      "alert",
    ] as const;
    for (const key of kindKeys) {
      if (res[key]) {
        return JSON.stringify(res[key], null, 2);
      }
    }
    const rest = Object.fromEntries(
      Object.entries(res).filter(([k]) => k !== "meta"),
    );
    return JSON.stringify(rest, null, 2);
  }

  $: specContent = getResourceSpec(resource);

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

<svelte:window on:click={handleWindowClick} on:keydown={handleWindowKeydown} />

{#if isOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    use:portal
    class="context-menu"
    style="left: {menuX}px; top: {menuY}px;"
    on:click|stopPropagation
    on:mousedown|stopPropagation
  >
    <button class="menu-item" on:click={handleViewSpec}>
      <Info size="12px" />
      <span>Describe</span>
    </button>
    <button class="menu-item" on:click={viewNodeTree}>
      <GitBranch size="12px" />
      <span>View Lineage</span>
    </button>
    {#if canOpenFile}
      <button class="menu-item" on:click={openFile}>
        <ExternalLink size="12px" />
        <span>Go to Resource</span>
      </button>
    {/if}
    {#if reconcileError}
      <button class="menu-item" on:click={handleCopyError}>
        <Copy size="12px" />
        <span>Copy Error Message</span>
      </button>
    {/if}
    {#if canRefresh}
      <div class="menu-separator"></div>
      <button class="menu-item" on:click={handleFullRefreshClick}>
        <RefreshCw size="12px" />
        <span>Full Refresh</span>
      </button>
      {#if isIncremental}
        <button class="menu-item" on:click={handleIncrementalRefresh}>
          <RefreshCw size="12px" />
          <span>Incremental Refresh</span>
        </button>
      {/if}
    {/if}
  </div>
{/if}

<AlertDialog.Root bind:open={fullRefreshConfirmOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Full Refresh {resourceName}?</AlertDialog.Title>
      <AlertDialog.Description>
        <div class="mt-1">
          A full refresh will re-ingest ALL data from scratch. This operation
          can take a significant amount of time and will update all dependent
          resources. Only proceed if you're certain this is necessary.
        </div>
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button
        type="secondary"
        onClick={() => {
          fullRefreshConfirmOpen = false;
        }}>Cancel</Button
      >
      <Button type="primary" onClick={confirmFullRefresh}>Yes, refresh</Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<Dialog.Root bind:open={specDialogOpen}>
  <Dialog.Content class="max-w-2xl max-h-[80vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title>
        {resourceName}
        <span class="text-fg-tertiary font-normal text-sm ml-2"
          >{kind ? displayResourceKind(kind) : ""}</span
        >
      </Dialog.Title>
    </Dialog.Header>
    <div class="spec-container">
      {#if !resource}
        <p class="text-sm text-fg-secondary">No resource data available</p>
      {:else}
        <pre class="spec-content">{specContent}</pre>
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .context-menu {
    @apply fixed z-50 min-w-[160px] rounded-md border bg-surface-base shadow-lg py-1;
  }

  .menu-item {
    @apply flex items-center gap-x-2 w-full px-3 py-1.5 text-xs text-left;
  }

  .menu-item:hover {
    @apply bg-surface-hover;
  }

  .menu-separator {
    @apply border-t my-1;
  }

  .spec-container {
    @apply overflow-auto flex-1 min-h-0;
  }
  .spec-content {
    @apply text-xs font-mono whitespace-pre-wrap bg-surface-subtle rounded-md p-4;
  }
</style>
