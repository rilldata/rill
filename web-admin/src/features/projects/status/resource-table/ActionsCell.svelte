<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { refreshableTypes } from "@rilldata/web-common/features/resources/resource-filter-utils";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import {
    RefreshCcwIcon,
    LayoutGridIcon,
    CodeIcon,
    ScrollTextIcon,
    AlertCircleIcon,
  } from "lucide-svelte";
  import { getAvailableModelActions } from "@rilldata/web-common/features/projects/status/tables/model-actions";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let resourceKind: string;
  export let resourceName: string;
  export let resource: V1Resource;
  export let isReconciling: boolean = false;
  export let onClickRefreshDialog: (
    resourceName: string,
    resourceKind: string,
    refreshType: "refresh" | "full" | "incremental",
  ) => void;
  export let onClickRefreshErroredPartitions: (resourceName: string) => void;
  export let onClickViewSpec: (
    resourceName: string,
    resourceKind: string,
    resource: V1Resource,
  ) => void;
  export let onViewLogsClick: (name: string) => void = () => {};
  export let onViewPartitionsClick:
    | ((resource: V1Resource) => void)
    | undefined = undefined;
  export let isDropdownOpen: boolean;
  export let onDropdownOpenChange: (isOpen: boolean) => void;

  $: isModel = resourceKind === ResourceKind.Model;
  $: canRefresh = refreshableTypes.includes(resourceKind as ResourceKind);

  $: actions = isModel ? getAvailableModelActions(resource) : [];
  $: isPartitioned = actions.includes("viewPartitions");
  $: isIncremental = actions.includes("incrementalRefresh");
  $: hasErroredPartitions = actions.includes("refreshErrored");

  $: refreshDisabled = isReconciling;
  $: refreshTooltip = isReconciling ? m.status_resource_reconciling() : "";
</script>

<DropdownMenu.Root open={isDropdownOpen} onOpenChange={onDropdownOpenChange}>
  <DropdownMenu.Trigger
    class="flex-none"
    aria-label={m.status_open_resource_actions()}
  >
    <IconButton rounded active={isDropdownOpen} size={20}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    <!-- Describe (always available) -->
    <DropdownMenu.Item
      class="font-normal flex items-center"
      onclick={() => onClickViewSpec(resourceName, resourceKind, resource)}
    >
      <div class="flex items-center">
        <CodeIcon size="12px" />
        <span class="ml-2">{m.status_action_describe()}</span>
      </div>
    </DropdownMenu.Item>

    <!-- View Logs (always available) -->
    <DropdownMenu.Item
      class="font-normal flex items-center"
      onclick={() => onViewLogsClick(resourceName)}
    >
      <div class="flex items-center">
        <ScrollTextIcon size="12px" />
        <span class="ml-2">{m.status_action_view_logs()}</span>
      </div>
    </DropdownMenu.Item>

    <!-- View Partitions (models only, if partitioned) -->
    {#if isPartitioned && onViewPartitionsClick}
      <DropdownMenu.Item
        class="font-normal flex items-center"
        onclick={() => onViewPartitionsClick?.(resource)}
      >
        <div class="flex items-center">
          <LayoutGridIcon size="12px" />
          <span class="ml-2">{m.status_action_view_partitions()}</span>
        </div>
      </DropdownMenu.Item>
    {/if}

    {#if isModel}
      <DropdownMenu.Separator />

      <!-- Refresh Errored Partitions (models with errors) -->
      {#if hasErroredPartitions}
        <Tooltip distance={8} suppress={!refreshDisabled}>
          <DropdownMenu.Item
            class="font-normal flex items-center"
            disabled={refreshDisabled}
            onclick={() => onClickRefreshErroredPartitions(resourceName)}
          >
            <div class="flex items-center">
              <AlertCircleIcon size="12px" />
              <span class="ml-2"
                >{m.status_action_refresh_errored_partitions()}</span
              >
            </div>
          </DropdownMenu.Item>
          <TooltipContent slot="tooltip-content"
            >{refreshTooltip}</TooltipContent
          >
        </Tooltip>
      {/if}

      <!-- Full Refresh -->
      <Tooltip distance={8} suppress={!refreshDisabled}>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          disabled={refreshDisabled}
          onclick={() =>
            onClickRefreshDialog(resourceName, resourceKind, "full")}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">{m.status_action_full_refresh()}</span>
          </div>
        </DropdownMenu.Item>
        <TooltipContent slot="tooltip-content">{refreshTooltip}</TooltipContent>
      </Tooltip>

      <!-- Incremental Refresh (incremental models only) -->
      {#if isIncremental}
        <Tooltip distance={8} suppress={!refreshDisabled}>
          <DropdownMenu.Item
            class="font-normal flex items-center"
            disabled={refreshDisabled}
            onclick={() =>
              onClickRefreshDialog(resourceName, resourceKind, "incremental")}
          >
            <div class="flex items-center">
              <RefreshCcwIcon size="12px" />
              <span class="ml-2">{m.status_action_incremental_refresh()}</span>
            </div>
          </DropdownMenu.Item>
          <TooltipContent slot="tooltip-content"
            >{refreshTooltip}</TooltipContent
          >
        </Tooltip>
      {/if}
    {:else if canRefresh}
      <DropdownMenu.Separator />

      <!-- Other resources only support a standard refresh. -->
      <Tooltip distance={8} suppress={!refreshDisabled}>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          disabled={refreshDisabled}
          onclick={() =>
            onClickRefreshDialog(resourceName, resourceKind, "refresh")}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">{m.status_action_refresh()}</span>
          </div>
        </DropdownMenu.Item>
        <TooltipContent slot="tooltip-content">{refreshTooltip}</TooltipContent>
      </Tooltip>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
