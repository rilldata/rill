<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    RefreshCw,
    RotateCcw,
    ExternalLink,
    GitFork,
  } from "lucide-svelte";
  import { createRuntimeServiceCreateTrigger } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { goto } from "$app/navigation";
  import type { ResourceNodeData } from "../shared/types";

  export let data: ResourceNodeData;

  let isOpen = false;
  let fullRefreshConfirmOpen = false;

  $: ({ instanceId } = $runtime);
  $: resource = data?.resource;
  $: kind = data?.kind;
  $: resourceName = resource?.meta?.name?.name ?? "";
  $: filePath = resource?.meta?.filePaths?.[0];
  $: canRefresh =
    (kind === ResourceKind.Model || kind === ResourceKind.Source) &&
    !!resourceName;

  const triggerMutation = createRuntimeServiceCreateTrigger();

  function refreshModel(full: boolean) {
    if (!resourceName) return;
    $triggerMutation.mutate({
      instanceId,
      data: {
        models: [{ model: resourceName, full }],
      },
    });
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
    goto(`/files${filePath}`);
    try {
      const prefs = JSON.parse(localStorage.getItem(filePath) || "{}");
      localStorage.setItem(
        filePath,
        JSON.stringify({ ...prefs, view: "code" }),
      );
    } catch (error) {
      console.warn(`Failed to save file view preference:`, error);
    }
  }

  function handleViewLineage() {
    if (!resource?.meta?.name) return;
    isOpen = false;
    const resourceKindName = resource.meta.name.kind;
    const resourceNameValue = resource.meta.name.name;
    const resourceId = encodeURIComponent(
      `${resourceKindName}:${resourceNameValue}`,
    );
    goto(`/graph?resource=${resourceId}&expanded=${resourceId}`);
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="actions-root" on:click|stopPropagation on:mousedown|stopPropagation>
  <DropdownMenu.Root bind:open={isOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isOpen} size={20}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end">
      {#if filePath}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={openFile}
        >
          <div class="flex items-center gap-x-2">
            <ExternalLink size="12px" />
            <span>Edit File</span>
          </div>
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={handleViewLineage}
      >
        <div class="flex items-center gap-x-2">
          <GitFork size="12px" />
          <span>View Lineage</span>
        </div>
      </DropdownMenu.Item>
      {#if canRefresh}
        <DropdownMenu.Separator />
        {#if kind === ResourceKind.Model}
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
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={handleFullRefreshClick}
        >
          <div class="flex items-center gap-x-2">
            <RotateCcw size="12px" />
            <span>Full Refresh</span>
          </div>
        </DropdownMenu.Item>
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

<style lang="postcss">
  .actions-root {
    @apply flex items-center;
  }
</style>
