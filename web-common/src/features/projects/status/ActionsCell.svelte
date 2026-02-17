<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { RefreshCcwIcon } from "lucide-svelte";

  export let resourceKind: string;
  export let resourceName: string;
  export let canRefresh: boolean;
  export let onRefresh: () => void;
  export let isDropdownOpen: boolean;
  export let onDropdownOpenChange: (isOpen: boolean) => void;
  export let resource: V1Resource | undefined = undefined;

  let isLoading = false;

  $: supportsIncremental =
    resourceKind === ResourceKind.Model &&
    resource?.model?.spec?.incremental === true;

  async function handleRefresh(refreshType: "full" | "incremental") {
    if (isLoading) return;
    isLoading = true;

    try {
      if (!$runtime?.instanceId || !$runtime?.host) {
        throw new Error("Runtime not initialized");
      }

      const triggerBody =
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

      const response = await fetch(
        `${$runtime.host}/v1/instances/${$runtime.instanceId}/trigger`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(triggerBody),
        },
      );

      if (!response.ok) {
        throw new Error(`Failed to refresh resource: ${response.statusText}`);
      }

      onRefresh();
      onDropdownOpenChange(false);
    } catch (error) {
      console.error("Failed to refresh resource:", error);
    } finally {
      isLoading = false;
    }
  }
</script>

{#if canRefresh}
  <DropdownMenu.Root open={isDropdownOpen} onOpenChange={onDropdownOpenChange}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen} size={20}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        disabled={isLoading}
        on:click={() => {
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
          on:click={() => {
            handleRefresh("incremental");
          }}
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
