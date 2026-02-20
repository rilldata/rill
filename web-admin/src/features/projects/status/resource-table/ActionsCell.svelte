<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { RefreshCcwIcon, CodeIcon } from "lucide-svelte";

  export let resourceKind: string;
  export let resourceName: string;
  export let canRefresh: boolean;
  export let resource: V1Resource;
  export let onClickRefreshDialog: (
    resourceName: string,
    resourceKind: string,
    refreshType: "full" | "incremental",
  ) => void;
  export let onClickViewSpec: (
    resourceName: string,
    resourceKind: string,
    resource: V1Resource,
  ) => void;
  export let isDropdownOpen: boolean;
  export let onDropdownOpenChange: (isOpen: boolean) => void;
</script>

<DropdownMenu.Root open={isDropdownOpen} onOpenChange={onDropdownOpenChange}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen} size={20}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    <DropdownMenu.Item
      class="font-normal flex items-center"
      on:click={() => onClickViewSpec(resourceName, resourceKind, resource)}
    >
      <div class="flex items-center">
        <CodeIcon size="12px" />
        <span class="ml-2">Describe</span>
      </div>
    </DropdownMenu.Item>
    {#if canRefresh}
      {#if resourceKind === ResourceKind.Model}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() =>
            onClickRefreshDialog(resourceName, resourceKind, "full")}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Full Refresh</span>
          </div>
        </DropdownMenu.Item>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() =>
            onClickRefreshDialog(resourceName, resourceKind, "incremental")}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Incremental Refresh</span>
          </div>
        </DropdownMenu.Item>
      {:else}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() =>
            onClickRefreshDialog(resourceName, resourceKind, "full")}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Full Refresh</span>
          </div>
        </DropdownMenu.Item>
      {/if}
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
