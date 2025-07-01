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
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let resourceKind: string;
  export let resourceName: string;
  export let canRefresh: boolean;

  let isConfirmDialogOpen = false;
  let isDropdownOpen = false;
  let refreshType: "full" | "incremental" = "full";

  const createTrigger = createRuntimeServiceCreateTrigger();
  const queryClient = useQueryClient();

  async function refresh() {
    if (resourceKind === ResourceKind.Model) {
      await $createTrigger.mutateAsync({
        instanceId: $runtime.instanceId,
        data: {
          models: [
            {
              model: resourceName,
              full: refreshType === "full",
            },
          ],
        },
      });
    } else {
      await $createTrigger.mutateAsync({
        instanceId: $runtime.instanceId,
        data: {
          resources: [{ kind: resourceKind, name: resourceName }],
        },
      });
    }

    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        $runtime.instanceId,
        undefined,
      ),
    });
  }
</script>

{#if canRefresh}
  <DropdownMenu.Root bind:open={isDropdownOpen}>
    <DropdownMenu.Trigger class="flex-none">
      <IconButton rounded active={isDropdownOpen} compact>
        <ThreeDot size="16px" />
      </IconButton>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      {#if resourceKind === ResourceKind.Model}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            refreshType = "full";
            isConfirmDialogOpen = true;
          }}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Full Refresh</span>
          </div>
        </DropdownMenu.Item>
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            refreshType = "incremental";
            isConfirmDialogOpen = true;
          }}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Incremental Refresh</span>
          </div>
        </DropdownMenu.Item>
      {:else}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            refreshType = "full";
            isConfirmDialogOpen = true;
          }}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Full Refresh</span>
          </div>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<RefreshResourceConfirmDialog
  bind:open={isConfirmDialogOpen}
  name={resourceName}
  onRefresh={() => {
    void refresh();
    isConfirmDialogOpen = false;
  }}
/>
