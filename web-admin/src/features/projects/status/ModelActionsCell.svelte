<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    RefreshCcwIcon,
    FileTextIcon,
    LayoutGridIcon,
    AlertCircleIcon,
    GitBranch,
  } from "lucide-svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let resource: V1Resource | undefined;
  export let isDropdownOpen: boolean;
  export let onDropdownOpenChange: (isOpen: boolean) => void;
  export let onModelInfoClick: (resource: V1Resource) => void;
  export let onViewPartitionsClick: (resource: V1Resource) => void;
  export let onRefreshErroredClick: (resource: V1Resource) => void;
  export let onIncrementalRefreshClick: (resource: V1Resource) => void;
  export let onFullRefreshClick: (resource: V1Resource) => void;

  $: isPartitioned = !!resource?.model?.spec?.partitionsResolver;
  $: isIncremental = !!resource?.model?.spec?.incremental;
  $: hasErroredPartitions =
    !!resource?.model?.state?.partitionsModelId &&
    !!resource?.model?.state?.partitionsHaveErrors;

  function handleViewInDAG() {
    const org = $page.params.organization;
    const proj = $page.params.project;
    const resourceName = resource?.meta?.name?.name;
    // resource param creates the graph, expanded param opens it in expanded view
    goto(`/${org}/${proj}/-/status/dag-viewer?resource=model:${resourceName}&expanded=${ResourceKind.Model}:${resourceName}`);
  }
</script>

{#if resource}
  <DropdownMenu.Root open={isDropdownOpen} onOpenChange={onDropdownOpenChange}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen} size={20}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={() => onModelInfoClick(resource)}
      >
        <div class="flex items-center">
          <FileTextIcon size="12px" />
          <span class="ml-2">Model Information</span>
        </div>
      </DropdownMenu.Item>

      {#if isPartitioned}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => onViewPartitionsClick(resource)}
        >
          <div class="flex items-center">
            <LayoutGridIcon size="12px" />
            <span class="ml-2">View Partitions</span>
          </div>
        </DropdownMenu.Item>
      {/if}

      {#if hasErroredPartitions}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => onRefreshErroredClick(resource)}
        >
          <div class="flex items-center">
            <AlertCircleIcon size="12px" />
            <span class="ml-2">Refresh Errored Partitions</span>
          </div>
        </DropdownMenu.Item>
      {/if}

      <DropdownMenu.Separator />

      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={() => onFullRefreshClick(resource)}
      >
        <div class="flex items-center">
          <RefreshCcwIcon size="12px" />
          <span class="ml-2">Full Refresh</span>
        </div>
      </DropdownMenu.Item>

      {#if isIncremental}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => onIncrementalRefreshClick(resource)}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Incremental Refresh</span>
          </div>
        </DropdownMenu.Item>
      {/if}

      <DropdownMenu.Separator />

      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={handleViewInDAG}
      >
        <div class="flex items-center">
          <GitBranch size="12px" />
          <span class="ml-2">View in DAG</span>
        </div>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
