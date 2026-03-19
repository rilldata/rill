<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    RefreshCcwIcon,
    LayoutGridIcon,
    AlertCircleIcon,
    CodeIcon,
    ScrollTextIcon,
  } from "lucide-svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { getAvailableModelActions } from "@rilldata/web-common/features/projects/status/tables/model-actions";

  export let resource: V1Resource | undefined;
  export let isReconciling: boolean = false;
  export let isDropdownOpen: boolean;
  export let onDropdownOpenChange: (isOpen: boolean) => void;
  export let onModelInfoClick: (resource: V1Resource) => void;
  export let onViewPartitionsClick: (resource: V1Resource) => void;
  export let onRefreshErroredClick: (resource: V1Resource) => void;
  export let onIncrementalRefreshClick: (resource: V1Resource) => void;
  export let onFullRefreshClick: (resource: V1Resource) => void;
  export let onViewLogsClick: ((name: string) => void) | undefined = undefined;

  $: actions = getAvailableModelActions(resource);
  $: isPartitioned = actions.includes("viewPartitions");
  $: isIncremental = actions.includes("incrementalRefresh");
  $: hasErroredPartitions = actions.includes("refreshErrored");

  $: refreshDisabled = isReconciling;
  $: refreshTooltip = isReconciling
    ? "Model is currently being reconciled"
    : "";
</script>

{#if resource}
  <DropdownMenu.Root open={isDropdownOpen} onOpenChange={onDropdownOpenChange}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen} size={20}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <!-- Describe (always available) -->
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={() => onModelInfoClick(resource)}
      >
        <div class="flex items-center">
          <CodeIcon size="12px" />
          <span class="ml-2">Describe</span>
        </div>
      </DropdownMenu.Item>

      <!-- View Logs (always available, optional) -->
      {#if onViewLogsClick}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => onViewLogsClick?.(resource.meta?.name?.name ?? "")}
        >
          <div class="flex items-center">
            <ScrollTextIcon size="12px" />
            <span class="ml-2">View Logs</span>
          </div>
        </DropdownMenu.Item>
      {/if}

      <!-- View Partitions (if partitioned, always available) -->
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

      <DropdownMenu.Separator />

      <!-- Refresh Errored Partitions (disabled when reconciling) -->
      {#if hasErroredPartitions}
        <Tooltip distance={8} suppress={!refreshDisabled}>
          <DropdownMenu.Item
            class="font-normal flex items-center"
            disabled={refreshDisabled}
            on:click={() => onRefreshErroredClick(resource)}
          >
            <div class="flex items-center">
              <AlertCircleIcon size="12px" />
              <span class="ml-2">Refresh Errored Partitions</span>
            </div>
          </DropdownMenu.Item>
          <TooltipContent slot="tooltip-content"
            >{refreshTooltip}</TooltipContent
          >
        </Tooltip>
      {/if}

      <!-- Full Refresh (disabled when reconciling) -->
      <Tooltip distance={8} suppress={!refreshDisabled}>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          disabled={refreshDisabled}
          on:click={() => onFullRefreshClick(resource)}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Full Refresh</span>
          </div>
        </DropdownMenu.Item>
        <TooltipContent slot="tooltip-content">{refreshTooltip}</TooltipContent>
      </Tooltip>

      <!-- Incremental Refresh (incremental models only, disabled when reconciling) -->
      {#if isIncremental}
        <Tooltip distance={8} suppress={!refreshDisabled}>
          <DropdownMenu.Item
            class="font-normal flex items-center"
            disabled={refreshDisabled}
            on:click={() => onIncrementalRefreshClick(resource)}
          >
            <div class="flex items-center">
              <RefreshCcwIcon size="12px" />
              <span class="ml-2">Incremental Refresh</span>
            </div>
          </DropdownMenu.Item>
          <TooltipContent slot="tooltip-content"
            >{refreshTooltip}</TooltipContent
          >
        </Tooltip>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
