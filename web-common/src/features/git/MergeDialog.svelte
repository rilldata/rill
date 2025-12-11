<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import * as Select from "@rilldata/web-common/components/select";
  import {
    createLocalServiceListBranches,
    createLocalServiceGitMerge,
    getLocalServiceListBranchesQueryKey,
    getLocalServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { GitMerge, AlertTriangle, ExternalLink } from "lucide-svelte";

  export let open = false;

  const branchesQuery = createLocalServiceListBranches();
  const mergeMutation = createLocalServiceGitMerge();

  $: branches = $branchesQuery.data?.branches ?? [];
  $: currentBranch = $branchesQuery.data?.currentBranch ?? "";
  $: otherBranches = branches.filter((b) => b.name !== currentBranch);

  let sourceBranch = "";
  let mergeResult: {
    success: boolean;
    hasConflicts: boolean;
    conflictingFiles: string[];
    pullRequestUrl: string;
  } | null = null;

  $: isMerging = $mergeMutation.isPending;

  function reset() {
    sourceBranch = "";
    mergeResult = null;
  }

  async function handleMerge() {
    if (!sourceBranch) return;

    try {
      const result = await $mergeMutation.mutateAsync({
        sourceBranch,
      });

      mergeResult = {
        success: result.success,
        hasConflicts: result.hasConflicts,
        conflictingFiles: result.conflictingFiles,
        pullRequestUrl: result.pullRequestUrl,
      };

      if (result.success) {
        void queryClient.invalidateQueries({
          queryKey: getLocalServiceListBranchesQueryKey(),
        });
        void queryClient.invalidateQueries({
          queryKey: getLocalServiceGitStatusQueryKey(),
        });

        eventBus.emit("notification", {
          message: `Successfully merged "${sourceBranch}" into "${currentBranch}"`,
        });

        setTimeout(() => {
          open = false;
          reset();
        }, 1500);
      }
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Merge failed: ${(e as Error).message}`,
      });
    }
  }

  function handleOpenPR() {
    if (mergeResult?.pullRequestUrl) {
      window.open(mergeResult.pullRequestUrl, "_blank");
    }
  }

  function handleClose() {
    open = false;
    reset();
  }
</script>

<Dialog.Root bind:open onOpenChange={(isOpen) => !isOpen && reset()}>
  <Dialog.Content class="max-w-md">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-x-2">
        <GitMerge size={18} />
        Merge Branch
      </Dialog.Title>
      <Dialog.Description>
        Merge changes from another branch into <strong>{currentBranch}</strong>
      </Dialog.Description>
    </Dialog.Header>

    <div class="py-4 space-y-4">
      {#if !mergeResult}
        <div class="space-y-2">
          <!-- svelte-ignore a11y-label-has-associated-control -->
          <label class="text-sm font-medium text-slate-700">Source branch</label
          >
          <Select.Root
            selected={{ value: sourceBranch, label: sourceBranch }}
            onSelectedChange={(s) => (sourceBranch = s?.value ?? "")}
          >
            <Select.Trigger class="w-full">
              <Select.Value placeholder="Select a branch to merge..." />
            </Select.Trigger>
            <Select.Content>
              {#each otherBranches as branch}
                <Select.Item value={branch.name}>{branch.name}</Select.Item>
              {/each}
            </Select.Content>
          </Select.Root>
        </div>

        {#if sourceBranch}
          <div class="p-3 bg-slate-50 rounded text-sm text-slate-600">
            <div class="font-medium">Preview:</div>
            <div class="mt-1">
              Merge <span class="font-mono text-primary-600"
                >{sourceBranch}</span
              >
              → <span class="font-mono text-primary-600">{currentBranch}</span>
            </div>
          </div>
        {/if}
      {:else if mergeResult.success}
        <div
          class="p-4 bg-green-50 border border-green-200 rounded text-center"
        >
          <div class="text-green-700 font-medium">Merge successful!</div>
          <div class="text-sm text-green-600 mt-1">
            Changes from "{sourceBranch}" have been merged into "{currentBranch}"
          </div>
        </div>
      {:else if mergeResult.hasConflicts}
        <div class="space-y-3">
          <div class="p-4 bg-amber-50 border border-amber-200 rounded">
            <div class="flex items-center gap-x-2 text-amber-700 font-medium">
              <AlertTriangle size={16} />
              Merge Conflicts Detected
            </div>
            <div class="text-sm text-amber-600 mt-1">
              The merge cannot be completed automatically due to conflicts in
              the following files:
            </div>
          </div>

          {#if mergeResult.conflictingFiles.length > 0}
            <div class="border rounded overflow-hidden">
              <div
                class="px-3 py-2 bg-slate-50 text-xs font-medium text-slate-600"
              >
                Conflicting files ({mergeResult.conflictingFiles.length})
              </div>
              <div class="max-h-32 overflow-y-auto">
                {#each mergeResult.conflictingFiles as file}
                  <div
                    class="px-3 py-1.5 text-sm font-mono text-slate-700 border-t"
                  >
                    {file}
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          {#if mergeResult.pullRequestUrl}
            <div class="text-sm text-slate-600">
              To resolve these conflicts, you can create a pull request on
              GitHub:
            </div>
            <Button type="secondary" onClick={handleOpenPR} class="w-full">
              <ExternalLink size={14} />
              Open Pull Request on GitHub
            </Button>
          {/if}
        </div>
      {/if}
    </div>

    <Dialog.Footer>
      <Button type="secondary" onClick={handleClose}>
        {mergeResult ? "Close" : "Cancel"}
      </Button>
      {#if !mergeResult}
        <Button
          type="primary"
          disabled={!sourceBranch || isMerging}
          onClick={handleMerge}
        >
          {isMerging ? "Merging..." : "Merge"}
        </Button>
      {/if}
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
