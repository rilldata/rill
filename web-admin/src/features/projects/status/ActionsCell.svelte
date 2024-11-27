<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { RefreshCcwIcon } from "lucide-svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let resourceKind: string;
  export let resourceName: string;
  export let isSource: boolean;

  let isDropdownOpen = false;

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  // TODO: start refetching interval for sources
  function refreshSources(resourceKind: string, resourceName: string) {
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

    // FIXME: Specify invalidating the resource kind
    // TODO: only accept name and path in the query params, so we can't really invalidate a single resource
    void queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey(
        $runtime.instanceId,
        // All resource "kinds"
        undefined,
      ),
    );
  }
</script>

{#if isSource}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen}>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      <DropdownMenu.Item
        class="font-normal flex items-center"
        on:click={() => refreshSources(resourceKind, resourceName)}
      >
        <RefreshCcwIcon size="12px" />
        <span class="ml-2">Refresh</span>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
