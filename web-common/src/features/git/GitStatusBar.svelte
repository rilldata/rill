<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import {
    createLocalServiceGitStatus,
    createLocalServiceGitPull,
    createLocalServiceGitPush,
    createLocalServiceGitCommit,
    createLocalServicePublishBranch,
    createLocalServiceDiscardChanges,
    getLocalServiceGitStatusQueryKey,
    getLocalServiceListBranchesQueryKey,
    getLocalServiceGetCommitHistoryQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    ArrowUp,
    ArrowDown,
    RefreshCw,
    Check,
    Upload,
    Undo2,
  } from "lucide-svelte";

  const statusQuery = createLocalServiceGitStatus();
  const pullMutation = createLocalServiceGitPull();
  const pushMutation = createLocalServiceGitPush();
  const commitMutation = createLocalServiceGitCommit();
  const publishMutation = createLocalServicePublishBranch();
  const discardMutation = createLocalServiceDiscardChanges();

  $: status = $statusQuery.data;
  $: localCommits = status?.localCommits ?? 0;
  $: remoteCommits = status?.remoteCommits ?? 0;
  $: hasLocalChanges = status?.localChanges ?? false;
  $: hasUpstream = status?.hasUpstream ?? false;
  $: isPulling = $pullMutation.isPending;
  $: isPushing = $pushMutation.isPending;
  $: isCommitting = $commitMutation.isPending;
  $: isPublishing = $publishMutation.isPending;
  $: isDiscarding = $discardMutation.isPending;

  let showCommitDialog = false;
  let showDiscardDialog = false;
  let commitMessage = "";

  function invalidateQueries() {
    void queryClient.invalidateQueries({
      queryKey: getLocalServiceGitStatusQueryKey(),
    });
    void queryClient.invalidateQueries({
      queryKey: getLocalServiceListBranchesQueryKey(),
    });
    void queryClient.invalidateQueries({
      queryKey: getLocalServiceGetCommitHistoryQueryKey(),
    });
  }

  async function handlePull() {
    try {
      await $pullMutation.mutateAsync({ discardLocal: false });
      invalidateQueries();
      eventBus.emit("notification", {
        message: "Pulled latest changes from remote",
      });
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to pull: ${(e as Error).message}`,
      });
    }
  }

  async function handlePush() {
    try {
      await $pushMutation.mutateAsync({});
      invalidateQueries();
      eventBus.emit("notification", {
        message: "Pushed changes to remote",
      });
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to push: ${(e as Error).message}`,
      });
    }
  }

  async function handlePublish() {
    try {
      const result = await $publishMutation.mutateAsync();
      invalidateQueries();
      eventBus.emit("notification", {
        message: `Branch "${result.branch}" published to remote`,
      });
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to publish: ${(e as Error).message}`,
      });
    }
  }

  async function handleCommit() {
    if (!commitMessage.trim()) return;

    try {
      await $commitMutation.mutateAsync({ message: commitMessage });
      invalidateQueries();
      eventBus.emit("notification", {
        message: "Changes committed successfully",
      });
      showCommitDialog = false;
      commitMessage = "";
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to commit: ${(e as Error).message}`,
      });
    }
  }

  async function handleDiscard() {
    try {
      await $discardMutation.mutateAsync({});
      invalidateQueries();
      eventBus.emit("notification", {
        message: "Changes discarded",
      });
      showDiscardDialog = false;
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to discard: ${(e as Error).message}`,
      });
    }
  }

  async function handleSync() {
    void queryClient.invalidateQueries({
      queryKey: getLocalServiceGitStatusQueryKey(),
    });
  }
</script>

<div class="flex items-center gap-x-2 text-xs">
  <!-- Sync status indicators -->
  {#if !hasUpstream}
    <span class="text-blue-600 text-[10px]" title="Local branch only"
      >• local</span
    >
  {:else if localCommits > 0 || remoteCommits > 0}
    <div class="flex items-center gap-x-1 text-slate-500">
      {#if localCommits > 0}
        <span
          class="flex items-center gap-x-0.5"
          title="{localCommits} commit(s) ahead of remote"
        >
          <ArrowUp size={12} />
          {localCommits}
        </span>
      {/if}
      {#if remoteCommits > 0}
        <span
          class="flex items-center gap-x-0.5"
          title="{remoteCommits} commit(s) behind remote"
        >
          <ArrowDown size={12} />
          {remoteCommits}
        </span>
      {/if}
    </div>
  {/if}

  {#if hasLocalChanges}
    <span class="text-amber-600 text-[10px]">• modified</span>
  {/if}

  <!-- Discard button (when there are uncommitted changes) -->
  {#if hasLocalChanges}
    <Button
      type="toolbar"
      compact
      small
      onClick={() => (showDiscardDialog = true)}
    >
      <Undo2 size={12} />
      Discard
    </Button>
  {/if}

  <!-- Commit button (when there are uncommitted changes) -->
  {#if hasLocalChanges}
    <Button
      type="toolbar"
      compact
      small
      onClick={() => (showCommitDialog = true)}
    >
      <Check size={12} />
      Commit
    </Button>
  {/if}

  <!-- Pull button (only when branch has upstream and behind remote) -->
  {#if hasUpstream && remoteCommits > 0}
    <Button
      type="toolbar"
      compact
      small
      disabled={isPulling}
      onClick={handlePull}
    >
      <ArrowDown size={12} />
      {isPulling ? "Pulling..." : "Pull"}
    </Button>
  {/if}

  <!-- Push button (only when branch has upstream and ahead of remote) -->
  {#if hasUpstream && localCommits > 0}
    <Button
      type="toolbar"
      compact
      small
      disabled={isPushing || remoteCommits > 0}
      onClick={handlePush}
    >
      <ArrowUp size={12} />
      {isPushing ? "Pushing..." : "Push"}
    </Button>
  {/if}

  <!-- Publish button (when branch doesn't have upstream) -->
  {#if !hasUpstream && localCommits > 0}
    <Button
      type="toolbar"
      compact
      small
      disabled={isPublishing}
      onClick={handlePublish}
    >
      <Upload size={12} />
      {isPublishing ? "Publishing..." : "Publish"}
    </Button>
  {/if}

  <!-- Refresh status -->
  <button
    class="p-1 hover:bg-slate-100 rounded text-slate-400 hover:text-slate-600"
    title="Refresh git status"
    on:click={handleSync}
  >
    <RefreshCw size={12} />
  </button>
</div>

<!-- Commit Dialog -->
<Dialog.Root bind:open={showCommitDialog}>
  <Dialog.Content class="max-w-md">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-x-2">
        <Check size={18} />
        Commit Changes
      </Dialog.Title>
      <Dialog.Description>
        Create a commit with your current changes
      </Dialog.Description>
    </Dialog.Header>

    <div class="py-4">
      <label class="block text-sm font-medium text-slate-700 mb-2">
        Commit message
      </label>
      <textarea
        bind:value={commitMessage}
        placeholder="Describe your changes..."
        class="w-full px-3 py-2 text-sm border border-slate-300 rounded focus:outline-none focus:ring-1 focus:ring-primary-500 resize-none"
        rows="3"
      />
    </div>

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (showCommitDialog = false)}>
        Cancel
      </Button>
      <Button
        type="primary"
        disabled={isCommitting || !commitMessage.trim()}
        onClick={handleCommit}
      >
        {isCommitting ? "Committing..." : "Commit"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<!-- Discard Dialog -->
<Dialog.Root bind:open={showDiscardDialog}>
  <Dialog.Content class="max-w-md">
    <Dialog.Header>
      <Dialog.Title class="flex items-center gap-x-2 text-red-600">
        <Undo2 size={18} />
        Discard Changes
      </Dialog.Title>
      <Dialog.Description>
        This will permanently discard all uncommitted changes
      </Dialog.Description>
    </Dialog.Header>

    <div class="py-4">
      <div
        class="p-3 bg-red-50 border border-red-200 rounded text-sm text-red-700"
      >
        <strong>Warning:</strong> This action cannot be undone. All modified and
        untracked files will be reverted to their last committed state.
      </div>
    </div>

    <Dialog.Footer>
      <Button type="secondary" onClick={() => (showDiscardDialog = false)}>
        Cancel
      </Button>
      <Button type="primary" disabled={isDiscarding} onClick={handleDiscard}>
        {isDiscarding ? "Discarding..." : "Discard All Changes"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
