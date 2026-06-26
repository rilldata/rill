<script lang="ts">
  import {
    GitDiffResponse_GitFileStatus,
    type GitDiffResponse_GitFileChange,
  } from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";
  import type { PartialMessage } from "@bufbuild/protobuf";
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

  const badges: Record<
    number,
    { letter: string; class: string; label: string }
  > = {
    [GitDiffResponse_GitFileStatus.ADDED]: {
      letter: "A",
      class: "badge-added",
      label: "Added",
    },
    [GitDiffResponse_GitFileStatus.MODIFIED]: {
      letter: "M",
      class: "badge-modified",
      label: "Modified",
    },
    [GitDiffResponse_GitFileStatus.DELETED]: {
      letter: "D",
      class: "badge-deleted",
      label: "Deleted",
    },
    [GitDiffResponse_GitFileStatus.RENAMED]: {
      letter: "R",
      class: "badge-renamed",
      label: "Renamed",
    },
  };

  function isRenamed(file: PartialMessage<GitDiffResponse_GitFileChange>) {
    return (
      file.status === GitDiffResponse_GitFileStatus.RENAMED && !!file.oldPath
    );
  }

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
          {@const badge =
            badges[file.status ?? GitDiffResponse_GitFileStatus.UNSPECIFIED] ??
            badges[GitDiffResponse_GitFileStatus.MODIFIED]}
          <li class="file-row">
            <span class="badge {badge.class}" title={badge.label}
              >{badge.letter}</span
            >
            {#if isRenamed(file)}
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
    @apply bg-green-100 text-green-700;
  }

  .badge-modified {
    @apply bg-yellow-100 text-yellow-700;
  }

  .badge-deleted {
    @apply bg-red-100 text-red-700;
  }

  .badge-renamed {
    @apply bg-blue-100 text-blue-700;
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
