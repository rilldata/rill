<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    createRuntimeServiceCreateTriggerMutation,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { RefreshCcwIcon, CodeIcon } from "lucide-svelte";

  export let resourceKind: string;
  export let resourceName: string;
  export let canRefresh: boolean;
  export let onRefresh: () => void;
  export let isDropdownOpen: boolean;
  export let onDropdownOpenChange: (isOpen: boolean) => void;
  export let resource: V1Resource | undefined = undefined;
  /** Callback to show the resource spec/describe dialog */
  export let onDescribe:
    | ((name: string, kind: string, resource: V1Resource) => void)
    | undefined = undefined;

  const runtimeClient = useRuntimeClient();
  const triggerMutation =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  $: isLoading = $triggerMutation.isPending;

  $: supportsIncremental =
    resourceKind === ResourceKind.Model &&
    resource?.model?.spec?.incremental === true;

  async function handleRefresh(refreshType: "full" | "incremental") {
    if (isLoading) return;

    try {
      const body =
        resourceKind === ResourceKind.Model
          ? {
              models: [
                {
                  model: resourceName,
                  full: refreshType === "full",
                },
              ],
            }
          : {
              resources: [
                {
                  kind: resourceKind,
                  name: resourceName,
                },
              ],
            };

      await $triggerMutation.mutateAsync(body);

      onRefresh();
      onDropdownOpenChange(false);
    } catch (error) {
      console.error("Failed to refresh resource:", error);
    }
  }
</script>

<DropdownMenu.Root open={isDropdownOpen} onOpenChange={onDropdownOpenChange}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen} size={20}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#if onDescribe && resource}
      <DropdownMenu.Item
        class="font-normal flex items-center"
        onclick={() => onDescribe?.(resourceName, resourceKind, resource)}
      >
        <div class="flex items-center">
          <CodeIcon size="12px" />
          <span class="ml-2">Describe</span>
        </div>
      </DropdownMenu.Item>
    {/if}
    {#if canRefresh}
      <DropdownMenu.Item
        class="font-normal flex items-center"
        disabled={isLoading}
        onclick={() => {
          handleRefresh("full");
        }}
      >
        <div class="flex items-center">
          <RefreshCcwIcon size="12px" />
          <span class="ml-2"
            >{isLoading ? "Refreshing..." : "Full Refresh"}</span
          >
        </div>
      </DropdownMenu.Item>
      {#if supportsIncremental}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          disabled={isLoading}
          onclick={() => {
            handleRefresh("incremental");
          }}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Incremental Refresh</span>
          </div>
        </DropdownMenu.Item>
      {/if}
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
