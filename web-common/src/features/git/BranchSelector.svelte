<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Popover from "@rilldata/web-common/components/popover";
  import {
    createLocalServiceListBranches,
    createLocalServiceCheckoutBranch,
    createLocalServiceCreateBranch,
    getLocalServiceListBranchesQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    GitBranch,
    ChevronDown,
    Plus,
    Check,
    ExternalLink,
  } from "lucide-svelte";

  let open = false;
  let showCreateBranch = false;
  let newBranchName = "";
  let creating = false;

  const branchesQuery = createLocalServiceListBranches();
  const checkoutMutation = createLocalServiceCheckoutBranch();
  const createBranchMutation = createLocalServiceCreateBranch();

  $: branches = $branchesQuery.data?.branches ?? [];
  $: currentBranch = $branchesQuery.data?.currentBranch ?? "main";
  $: hasUncommittedChanges =
    $branchesQuery.data?.hasUncommittedChanges ?? false;

  async function handleCheckout(branchName: string) {
    if (branchName === currentBranch) {
      open = false;
      return;
    }

    if (hasUncommittedChanges) {
      const confirmed = window.confirm(
        "You have uncommitted changes. Switching branches will discard them. Continue?",
      );
      if (!confirmed) return;
    }

    try {
      await $checkoutMutation.mutateAsync({
        branch: branchName,
        force: hasUncommittedChanges,
      });

      void queryClient.invalidateQueries({
        queryKey: getLocalServiceListBranchesQueryKey(),
      });

      eventBus.emit("notification", {
        message: `Switched to branch "${branchName}"`,
      });
      open = false;
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to switch branch: ${(e as Error).message}`,
      });
    }
  }

  async function handleCreateBranch() {
    if (!newBranchName.trim()) return;

    creating = true;
    try {
      await $createBranchMutation.mutateAsync({
        name: newBranchName.trim(),
        checkout: true,
      });

      void queryClient.invalidateQueries({
        queryKey: getLocalServiceListBranchesQueryKey(),
      });

      eventBus.emit("notification", {
        message: `Created and switched to branch "${newBranchName}"`,
      });
      newBranchName = "";
      showCreateBranch = false;
      open = false;
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to create branch: ${(e as Error).message}`,
      });
    } finally {
      creating = false;
    }
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter" && showCreateBranch) {
      void handleCreateBranch();
    } else if (event.key === "Escape") {
      if (showCreateBranch) {
        showCreateBranch = false;
        newBranchName = "";
      } else {
        open = false;
      }
    }
  }

  function handleDeploymentClick(
    event: MouseEvent,
    deploymentUrl: string | undefined,
  ) {
    if (!deploymentUrl) return;
    event.stopPropagation();
    window.open(deploymentUrl, "_blank");
  }

  function getDeploymentBadgeClass(environment: string | undefined): string {
    if (environment === "prod") {
      return "bg-blue-100 text-blue-700 hover:bg-blue-200";
    } else if (environment === "dev") {
      return "bg-amber-100 text-amber-700 hover:bg-amber-200";
    } else if (environment === "preview") {
      return "bg-purple-100 text-purple-700 hover:bg-purple-200";
    }
    return "bg-slate-100 text-slate-700 hover:bg-slate-200";
  }
</script>

<Popover.Root bind:open>
  <Popover.Trigger asChild let:builder>
    <Button type="toolbar" builders={[builder]} class="gap-x-1.5">
      <GitBranch size={14} class="text-slate-500" />
      <span class="text-xs font-medium max-w-[120px] truncate">
        {currentBranch}
      </span>
      {#if hasUncommittedChanges}
        <span
          class="w-1.5 h-1.5 rounded-full bg-amber-500"
          title="Uncommitted changes"
        ></span>
      {/if}
      <ChevronDown size={12} class="text-slate-400" />
    </Button>
  </Popover.Trigger>

  <Popover.Content align="start" class="w-[280px] p-0" sideOffset={4}>
    <div class="p-2 border-b border-slate-200">
      <div class="text-xs font-medium text-slate-500 mb-1">Switch branch</div>
      {#if hasUncommittedChanges}
        <div class="text-[10px] text-amber-600 flex items-center gap-x-1">
          <span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
          You have uncommitted changes
        </div>
      {/if}
    </div>

    <div class="max-h-[240px] overflow-y-auto">
      {#if $branchesQuery.isLoading}
        <div class="p-3 text-xs text-slate-500 text-center">
          Loading branches...
        </div>
      {:else if branches.length === 0}
        <div class="p-3 text-xs text-slate-500 text-center">
          No branches found
        </div>
      {:else}
        {#each branches as branch}
          <button
            class="w-full px-3 py-2 text-left hover:bg-slate-50 flex items-center gap-x-2 text-sm"
            class:bg-primary-50={branch.isCurrent}
            on:click={() => handleCheckout(branch.name)}
          >
            <div class="w-4 flex-shrink-0">
              {#if branch.isCurrent}
                <Check size={14} class="text-primary-600" />
              {/if}
            </div>
            <div class="flex-1 min-w-0">
              <div
                class="font-medium truncate"
                class:text-primary-700={branch.isCurrent}
              >
                {branch.name}
              </div>
              {#if branch.lastCommitMessage}
                <div class="text-xs text-slate-400 truncate">
                  {branch.lastCommitMessage}
                </div>
              {/if}
            </div>
            <div class="flex items-center gap-x-1 text-[10px] flex-shrink-0">
              <!-- Show deployment badge if available -->
              {#if branch.deploymentEnvironment && branch.deploymentUrl}
                <button
                  class="px-1.5 py-0.5 rounded flex items-center gap-x-1 transition-colors {getDeploymentBadgeClass(
                    branch.deploymentEnvironment,
                  )}"
                  on:click={(e) =>
                    handleDeploymentClick(e, branch.deploymentUrl)}
                  title="Open deployment in new tab"
                >
                  <span class="font-medium">{branch.deploymentEnvironment}</span
                  >
                  <ExternalLink size={10} />
                </button>
              {/if}
              <!-- Always show local/remote status -->
              {#if branch.isLocal && branch.isRemote}
                <span class="px-1 py-0.5 bg-slate-100 rounded text-slate-400"
                  >synced</span
                >
              {:else if branch.isLocal}
                <span class="px-1 py-0.5 bg-slate-100 rounded text-slate-400"
                  >local</span
                >
              {:else}
                <span class="px-1 py-0.5 bg-slate-100 rounded text-slate-400"
                  >remote</span
                >
              {/if}
            </div>
          </button>
        {/each}
      {/if}
    </div>

    <div class="border-t border-slate-200 p-2">
      {#if showCreateBranch}
        <div class="flex items-center gap-x-2">
          <!-- svelte-ignore a11y-autofocus -->
          <input
            type="text"
            bind:value={newBranchName}
            placeholder="Branch name..."
            class="flex-1 px-2 py-1.5 text-sm border border-slate-300 rounded focus:outline-none focus:ring-1 focus:ring-primary-500"
            on:keydown={handleKeyDown}
            autofocus
          />
          <Button
            type="primary"
            small
            disabled={!newBranchName.trim() || creating}
            onClick={handleCreateBranch}
          >
            {creating ? "..." : "Create"}
          </Button>
        </div>
      {:else}
        <button
          class="w-full px-2 py-1.5 text-left text-sm text-slate-600 hover:bg-slate-50 rounded flex items-center gap-x-2"
          on:click={() => (showCreateBranch = true)}
        >
          <Plus size={14} />
          Create new branch
        </button>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>
