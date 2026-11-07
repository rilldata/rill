<script lang="ts">
  import { createRuntimeServiceGitDiff } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ChevronDown, ChevronRight, Eye, RotateCcw } from "lucide-svelte";
  import FileChangeBadge from "./FileChangeBadge.svelte";
  import RevertConfirmDialog from "./RevertConfirmDialog.svelte";
  import { revertFiles } from "./revert";

  // remoteBranch is the branch to compare against; open gates the query so the
  // changed-files list is only fetched while the popover is open, not on page load.
  // onViewDiff opens the full diff. The host (popover) owns the dialog so it survives the popover
  // closing. Passing a path requests scrolling to that file.
  let {
    remoteBranch,
    open,
    onViewDiff = () => {},
  }: {
    remoteBranch: string | undefined;
    open: boolean;
    onViewDiff?: (path?: string) => void;
  } = $props();

  const client = useRuntimeClient();
  // fetch: true updates the remote-tracking ref so the list reflects the latest remote; the diff
  // dialog reuses this fetch. refetchOnMount "always" overrides the global default (false) so the
  // list is re-fetched every time the popover reopens and this component remounts, rather than
  // serving a stale cache from a previous session.
  let changesQuery = $derived(
    createRuntimeServiceGitDiff(
      client,
      { remoteBranch, fetch: true },
      { query: { enabled: open && !!remoteBranch, refetchOnMount: "always" } },
    ),
  );
  let changedFiles = $derived($changesQuery.data?.changedFiles ?? []);
  let isFetching = $derived($changesQuery.isFetching);

  let expanded = $state(false);

  let count = $derived(changedFiles.length);

  // Revert state. revertTargets holds the subpath-relative paths queued for the confirm dialog;
  // a single-file revert queues one path, "Discard all" queues every changed path.
  let revertOpen = $state(false);
  let revertTargets = $state<string[]>([]);
  let isReverting = $state(false);

  function openRevert(paths: string[]) {
    if (paths.length === 0) return;
    revertTargets = paths;
    revertOpen = true;
  }

  function allPaths(): string[] {
    return changedFiles.map((f) => f.path).filter((p): p is string => !!p);
  }

  async function handleRevert() {
    isReverting = true;
    try {
      const reverted = await revertFiles(client, remoteBranch, revertTargets);
      eventBus.emit("notification", {
        type: "success",
        message: m.edit_changes_reverted({ count: reverted }),
      });
      revertOpen = false;
    } catch {
      eventBus.emit("notification", {
        type: "error",
        message: m.edit_changes_revert_failed(),
      });
    } finally {
      isReverting = false;
    }
  }
</script>

{#if isFetching}
  <div class="loading">
    <DelayedSpinner isLoading={true} size="14px" />
    <span>{m.edit_changes_checking()}</span>
  </div>
{:else if count > 0}
  <div class="changed-files">
    <div class="header-row">
      <button
        type="button"
        class="header"
        onclick={() => (expanded = !expanded)}
      >
        {#if expanded}
          <ChevronDown size="14" />
        {:else}
          <ChevronRight size="14" />
        {/if}
        <span>{m.edit_changes_files_changed({ count })}</span>
      </button>
      <div class="header-actions">
        <button
          type="button"
          class="text-action"
          onclick={() => openRevert(allPaths())}
        >
          <RotateCcw size="13" />
          {m.edit_changes_discard_all()}
        </button>
        <button type="button" class="text-action" onclick={() => onViewDiff()}>
          <Eye size="13" />
          {m.edit_changes_view_diff()}
        </button>
      </div>
    </div>
    {#if expanded}
      <ul class="file-list">
        {#each changedFiles as file (file.path)}
          {@const status = file.status as unknown as string}
          <li class="file-row">
            <button
              type="button"
              class="file-info"
              onclick={() => onViewDiff(file.path)}
            >
              <FileChangeBadge status={file.status} />
              {#if status === "GIT_FILE_STATUS_RENAMED"}
                <span class="file-path" title="{file.oldPath} → {file.path}">
                  <span class="old-path">{file.oldPath}</span>
                  <span class="arrow">→</span>{file.path}
                </span>
              {:else}
                <span class="file-path" title={file.path}>{file.path}</span>
              {/if}
            </button>
            <button
              type="button"
              class="revert-btn"
              title={m.edit_changes_revert()}
              aria-label={m.edit_changes_revert()}
              onclick={() => openRevert(file.path ? [file.path] : [])}
            >
              <RotateCcw size="13" />
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<RevertConfirmDialog
  bind:open={revertOpen}
  files={revertTargets}
  loading={isReverting}
  onConfirm={handleRevert}
/>

<style lang="postcss">
  .loading {
    @apply flex items-center gap-x-2 text-xs text-fg-secondary;
  }

  .changed-files {
    @apply flex flex-col gap-y-1;
  }

  .header-row {
    @apply flex items-center justify-between;
  }

  .header {
    @apply flex items-center gap-x-1 text-xs text-fg-secondary;
    @apply hover:text-fg-primary;
  }

  .header-actions {
    @apply flex items-center gap-x-3;
  }

  .text-action {
    @apply flex items-center gap-x-1 text-xs font-medium text-primary-600;
    @apply hover:text-primary-700;
  }

  .file-list {
    @apply flex flex-col gap-y-0.5 max-h-40 overflow-y-auto;
  }

  .file-row {
    @apply flex items-center gap-x-1 w-full;
    @apply rounded hover:bg-gray-100;
  }

  .file-info {
    @apply flex items-center gap-x-2 flex-1 min-w-0 text-left text-xs;
    @apply px-1 py-0.5;
  }

  .file-path {
    @apply text-fg-secondary truncate;
  }

  .file-row:hover .file-path {
    @apply text-fg-primary;
  }

  .revert-btn {
    @apply flex-none p-1 rounded text-fg-secondary opacity-0;
    @apply hover:bg-gray-200 hover:text-fg-primary;
  }

  .file-row:hover .revert-btn {
    @apply opacity-100;
  }

  .old-path {
    @apply text-fg-disabled;
  }

  .arrow {
    @apply px-0.5 text-fg-disabled;
  }
</style>
