<script lang="ts">
  import { createRuntimeServiceGitDiff } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { ChevronDown, ChevronRight, Eye } from "lucide-svelte";
  import FileChangeBadge from "./FileChangeBadge.svelte";

  // remoteBranch is the branch to compare against; open gates the query so the
  // changed-files list is only fetched while the popover is open, not on page load.
  export let remoteBranch: string | undefined;
  export let open: boolean;
  // onViewDiff opens the full diff. The host (popover) owns the dialog so it survives the popover
  // closing. Passing a path requests scrolling to that file.
  export let onViewDiff: (path?: string) => void = () => {};

  const client = useRuntimeClient();
  // refetchOnMount "always" overrides the global default (false) so the list is
  // re-fetched every time the popover reopens and this component remounts, rather
  // than serving a stale cache from a previous session.
  $: changesQuery = createRuntimeServiceGitDiff(
    client,
    { remoteBranch },
    { query: { enabled: open && !!remoteBranch, refetchOnMount: "always" } },
  );
  $: changedFiles = $changesQuery.data?.changedFiles ?? [];
  $: isFetching = $changesQuery.isFetching;

  let expanded = false;

  $: count = changedFiles.length;
</script>

{#if isFetching}
  <div class="loading">
    <DelayedSpinner isLoading={true} size="14px" />
    <span>Checking for changes…</span>
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
        <span>{count} {count === 1 ? "file" : "files"} changed</span>
      </button>
      <button type="button" class="view-diff" onclick={() => onViewDiff()}>
        <Eye size="13" />
        View diff
      </button>
    </div>
    {#if expanded}
      <ul class="file-list">
        {#each changedFiles as file (file.path)}
          {@const status = file.status as unknown as string}
          <li>
            <button
              type="button"
              class="file-row"
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
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

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

  .view-diff {
    @apply flex items-center gap-x-1 text-xs font-medium text-primary-600;
    @apply hover:text-primary-700;
  }

  .file-list {
    @apply flex flex-col gap-y-0.5 max-h-40 overflow-y-auto;
  }

  .file-row {
    @apply flex items-center gap-x-2 w-full text-left text-xs;
    @apply rounded px-1 py-0.5 hover:bg-gray-100;
  }

  .file-path {
    @apply text-fg-secondary truncate;
  }

  .file-row:hover .file-path {
    @apply text-fg-primary;
  }

  .old-path {
    @apply text-fg-disabled;
  }

  .arrow {
    @apply px-0.5 text-fg-disabled;
  }
</style>
