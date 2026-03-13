<script context="module" lang="ts">
  // Track the currently open dropdown so right-clicking another node
  // closes the previous one instead of stacking.
  let closeActive: (() => void) | null = null;
</script>

<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
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

  const graphNav = getGraphNavigation();

  export let data: ResourceNodeData;

  let isOpen = false;
  let fullRefreshConfirmOpen = false;
  let specDialogOpen = false;

  function close() {
    isOpen = false;
  }

  export function open() {
    if (closeActive && closeActive !== close) closeActive();
    closeActive = close;
    isOpen = true;
  }

  // Keep module-level tracker in sync when dropdown closes via outside click
  $: if (!isOpen && closeActive === close) closeActive = null;

  const runtimeClient = useRuntimeClient();
  $: resource = data?.resource;
  $: kind = data?.kind;
  $: resourceName = resource?.meta?.name?.name ?? "";
  $: filePath = resource?.meta?.filePaths?.[0];
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
    isOpen = false;
    refreshModel(false);
  }

  function handleFullRefreshClick() {
    isOpen = false;
    fullRefreshConfirmOpen = true;
  }

  function confirmFullRefresh() {
    fullRefreshConfirmOpen = false;
    refreshModel(true);
  }

  function openFile() {
    if (!filePath) return;
    isOpen = false;
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
    isOpen = false;
    specDialogOpen = true;
  }

  function handleCopyError() {
    isOpen = false;
    navigator.clipboard.writeText(reconcileError).catch((err) => {
      console.error("Failed to copy error:", err);
    });
  }

  function viewNodeTree() {
    isOpen = false;
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

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="actions-root" on:click|stopPropagation on:mousedown|stopPropagation>
  <DropdownMenu.Root bind:open={isOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isOpen} size={28}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content side="right" align="start">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={handleViewSpec}
      >
        <div class="flex items-center gap-x-2">
          <Info size="12px" />
          <span>Describe</span>
        </div>
      </DropdownMenu.Item>
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={viewNodeTree}
      >
        <div class="flex items-center gap-x-2">
          <GitBranch size="12px" />
          <span>View Lineage</span>
        </div>
      </DropdownMenu.Item>
      {#if filePath}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={openFile}
        >
          <div class="flex items-center gap-x-2">
            <ExternalLink size="12px" />
            <span>Go to Resource</span>
          </div>
        </DropdownMenu.Item>
      {/if}
      {#if reconcileError}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={handleCopyError}
        >
          <div class="flex items-center gap-x-2">
            <Copy size="12px" />
            <span>Copy Error Message</span>
          </div>
        </DropdownMenu.Item>
      {/if}
      {#if canRefresh}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={handleFullRefreshClick}
        >
          <div class="flex items-center gap-x-2">
            <RefreshCw size="12px" />
            <span>Full Refresh</span>
          </div>
        </DropdownMenu.Item>
        {#if isIncremental}
          <DropdownMenu.Item
            class="font-normal flex items-center"
            on:click={handleIncrementalRefresh}
          >
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
  .actions-root {
    @apply flex items-center;
  }
  .spec-container {
    @apply overflow-auto flex-1 min-h-0;
  }
  .spec-content {
    @apply text-xs font-mono whitespace-pre-wrap bg-surface-subtle rounded-md p-4;
  }
</style>
