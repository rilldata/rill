<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { createRuntimeServiceCreateTrigger } from "@rilldata/web-common/runtime-client";
  import { RefreshCcwIcon } from "lucide-svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";

  export let resourceKind: string;
  export let resourceName: string;
  export let canRefresh: boolean;
  export let triggerRefresh: () => void;

  let isDropdownOpen = false;

  const createTrigger = createRuntimeServiceCreateTrigger();

  function refresh(resourceKind: string, resourceName: string) {
    void $createTrigger.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        resources: [
          {
            kind: resourceKind,
            name: resourceName,
          },
        ],
      },
    });

    triggerRefresh();
  }
</script>

{#if canRefresh}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={() => refresh(resourceKind, resourceName)}
      >
        <Tooltip location="left" alignment="middle" distance={16}>
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Refresh</span>
          </div>
          <TooltipContent maxWidth="400px" slot="tooltip-content">
            Refreshing this resource will update all dependent resources.
          </TooltipContent>
        </Tooltip>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
