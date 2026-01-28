<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { GitBranch, RefreshCcwIcon } from "lucide-svelte";

  export let resourceKind: string;
  export let resourceName: string;
  export let canRefresh: boolean;
  export let isIncremental: boolean = false;
  export let hasErroredPartitions: boolean = false;
  export let onClickRefreshDialog: (
    resourceName: string,
    resourceKind: string,
    refreshType: "full" | "incremental" | "errored-partitions",
  ) => void;
  export let isDropdownOpen: boolean;
  export let onDropdownOpenChange: (isOpen: boolean) => void;

  // Convert full ResourceKind to short name for URL (e.g., "rill.runtime.v1.Model" -> "model")
  function kindToShortName(kind: string): string {
    const parts = kind.split(".");
    return parts[parts.length - 1].toLowerCase();
  }

  function handleViewInDAG() {
    const org = $page.params.organization;
    const proj = $page.params.project;
    const shortKind = kindToShortName(resourceKind);
    // resource param creates the graph, expanded param opens it in expanded view
    goto(`/${org}/${proj}/-/status/dag-viewer?resource=${shortKind}:${resourceName}&expanded=${resourceKind}:${resourceName}`);
  }
</script>

<DropdownMenu.Root open={isDropdownOpen} onOpenChange={onDropdownOpenChange}>
  <DropdownMenu.Trigger class="flex-none">
    <IconButton rounded active={isDropdownOpen} size={20}>
      <ThreeDot size="16px" />
    </IconButton>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#if canRefresh}
      {#if resourceKind === ResourceKind.Model}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            onClickRefreshDialog(resourceName, resourceKind, "full");
          }}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Full Refresh</span>
          </div>
        </DropdownMenu.Item>
        {#if isIncremental}
          <DropdownMenu.Item
            class="font-normal flex items-center"
            on:click={() => {
              onClickRefreshDialog(resourceName, resourceKind, "incremental");
            }}
          >
            <div class="flex items-center">
              <RefreshCcwIcon size="12px" />
              <span class="ml-2">Incremental Refresh</span>
            </div>
          </DropdownMenu.Item>
        {/if}
        {#if hasErroredPartitions}
          <DropdownMenu.Item
            class="font-normal flex items-center"
            on:click={() => {
              onClickRefreshDialog(
                resourceName,
                resourceKind,
                "errored-partitions",
              );
            }}
          >
            <div class="flex items-center">
              <RefreshCcwIcon size="12px" />
              <span class="ml-2">Refresh Errored Partitions</span>
            </div>
          </DropdownMenu.Item>
        {/if}
      {:else}
        <DropdownMenu.Item
          class="font-normal flex items-center"
          on:click={() => {
            onClickRefreshDialog(resourceName, resourceKind, "full");
          }}
        >
          <div class="flex items-center">
            <RefreshCcwIcon size="12px" />
            <span class="ml-2">Full Refresh</span>
          </div>
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Separator />
    {/if}
    <DropdownMenu.Item
      class="font-normal flex items-center"
      on:click={handleViewInDAG}
    >
      <div class="flex items-center">
        <GitBranch size="12px" />
        <span class="ml-2">View in DAG</span>
      </div>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>
