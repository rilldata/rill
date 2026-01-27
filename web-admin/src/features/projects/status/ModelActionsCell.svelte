<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import {
    RefreshCcwIcon,
    FileTextIcon,
    LayoutGridIcon,
    AlertCircleIcon,
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
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
