<script lang="ts">
  import { createRuntimeServiceGitDiff } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { ChevronDown, ChevronRight } from "lucide-svelte";

  // remoteBranch is the branch to compare against; open gates the query so the
  // changed-files list is only fetched while the popover is open, not on page load.
  export let remoteBranch: string | undefined;
  export let open: boolean;

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

  // The v2 client serializes proto enums as their JSON string names (e.g. "GIT_FILE_STATUS_ADDED").
  const badges: Record<
    string,
    { letter: string; class: string; label: string }
  > = {
    GIT_FILE_STATUS_ADDED: {
      letter: "A",
      class: "badge-added",
      label: "Added",
    },
    GIT_FILE_STATUS_MODIFIED: {
      letter: "M",
      class: "badge-modified",
      label: "Modified",
    },
    GIT_FILE_STATUS_DELETED: {
      letter: "D",
      class: "badge-deleted",
      label: "Deleted",
    },
    GIT_FILE_STATUS_RENAMED: {
      letter: "R",
      class: "badge-renamed",
      label: "Renamed",
    },
  };

  $: count = changedFiles.length;
</script>

{#if isFetching}
  <div class="loading">
    <DelayedSpinner isLoading={true} size="14px" />
    <span>Checking for changes…</span>
  </div>
{:else if count > 0}
  <div class="changed-files">
    <button type="button" class="header" onclick={() => (expanded = !expanded)}>
      {#if expanded}
        <ChevronDown size="14" />
      {:else}
        <ChevronRight size="14" />
      {/if}
      <span>{count} {count === 1 ? "file" : "files"} changed</span>
    </button>
    {#if expanded}
      <ul class="file-list">
        {#each changedFiles as file (file.path)}
          {@const status = file.status as unknown as string}
          {@const badge = badges[status] ?? badges["GIT_FILE_STATUS_MODIFIED"]}
          <li class="file-row">
            <span class="badge {badge.class}" title={badge.label}
              >{badge.letter}</span
            >
            {#if status === "GIT_FILE_STATUS_RENAMED"}
              <span class="file-path" title="{file.oldPath} → {file.path}">
                <span class="old-path">{file.oldPath}</span>
                <span class="arrow">→</span>{file.path}
              </span>
            {:else}
              <span class="file-path" title={file.path}>{file.path}</span>
            {/if}
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

  .header {
    @apply flex items-center gap-x-1 text-xs text-fg-secondary;
    @apply hover:text-fg-primary;
  }

  .file-list {
    @apply flex flex-col gap-y-1 max-h-40 overflow-y-auto;
  }

  .file-row {
    @apply flex items-center gap-x-2 text-xs;
  }

  .badge {
    @apply flex-none text-[0.625rem] leading-none px-1 py-0.5 rounded;
    @apply font-mono font-medium;
  }

  .badge-added {
    @apply bg-primary-100 text-primary-800;
  }

  .badge-modified {
    @apply bg-yellow-100 text-yellow-700;
  }

  .badge-deleted {
    @apply bg-red-100 text-red-700;
  }

  .badge-renamed {
    @apply bg-secondary-100 text-secondary-800;
  }

  .file-path {
    @apply text-fg-secondary truncate;
  }

  .old-path {
    @apply text-fg-disabled;
  }

  .arrow {
    @apply px-0.5 text-fg-disabled;
  }
</style>
