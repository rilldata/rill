<script lang="ts">
  import {
    V1GitStatusResponseGitFileStatus,
    type V1GitStatusResponseGitFileChange,
  } from "@rilldata/web-common/runtime-client";
  import { ChevronDown, ChevronRight } from "lucide-svelte";

  export let changedFiles: V1GitStatusResponseGitFileChange[];

  let expanded = false;

  const badges: Record<
    string,
    { letter: string; class: string; label: string }
  > = {
    [V1GitStatusResponseGitFileStatus.GIT_FILE_STATUS_ADDED]: {
      letter: "A",
      class: "badge-added",
      label: "Added",
    },
    [V1GitStatusResponseGitFileStatus.GIT_FILE_STATUS_MODIFIED]: {
      letter: "M",
      class: "badge-modified",
      label: "Modified",
    },
    [V1GitStatusResponseGitFileStatus.GIT_FILE_STATUS_DELETED]: {
      letter: "D",
      class: "badge-deleted",
      label: "Deleted",
    },
    [V1GitStatusResponseGitFileStatus.GIT_FILE_STATUS_RENAMED]: {
      letter: "R",
      class: "badge-renamed",
      label: "Renamed",
    },
  };

  function isRenamed(file: V1GitStatusResponseGitFileChange) {
    return (
      file.status ===
        V1GitStatusResponseGitFileStatus.GIT_FILE_STATUS_RENAMED &&
      !!file.oldPath
    );
  }

  $: count = changedFiles.length;
</script>

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
          (file.status && badges[file.status]) ??
          badges[V1GitStatusResponseGitFileStatus.GIT_FILE_STATUS_MODIFIED]}
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

<style lang="postcss">
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
