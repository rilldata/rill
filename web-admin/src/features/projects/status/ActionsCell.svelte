<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { createRuntimeServiceCreateTrigger } from "@rilldata/web-common/runtime-client";
  import { RefreshCcwIcon } from "lucide-svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import RefreshResourceConfirmDialog from "./RefreshResourceConfirmDialog.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";

  export let resourceKind: string;
  export let resourceName: string;
  export let canRefresh: boolean;

  let isConfirmDialogOpen = false;
  let isDropdownOpen = false;

  const createTrigger = createRuntimeServiceCreateTrigger();
  const queryClient = useQueryClient();

  async function refresh(resourceKind: string, resourceName: string) {
    await $createTrigger.mutateAsync({
      instanceId: $runtime.instanceId,
      data: {
        resources: [{ kind: resourceKind, name: resourceName }],
      },
    });

    await queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey($runtime.instanceId, undefined),
    );
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
        on:click={() => {
          isConfirmDialogOpen = true;
        }}
      >
        <div class="flex items-center">
          <RefreshCcwIcon size="12px" />
          <span class="ml-2">Refresh</span>
        </div>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<RefreshResourceConfirmDialog
  bind:open={isConfirmDialogOpen}
  name={resourceName}
  onRefresh={() => {
    void refresh(resourceKind, resourceName);
    isConfirmDialogOpen = false;
  }}
/>
