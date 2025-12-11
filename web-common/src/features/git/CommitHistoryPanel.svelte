<script lang="ts">
  import { createLocalServiceGetCommitHistory } from "@rilldata/web-common/runtime-client/local-service";
  import { Clock, User } from "lucide-svelte";

  export let branch = "";
  export let limit = 50;

  const historyQuery = createLocalServiceGetCommitHistory({ branch, limit });

  $: commits = $historyQuery.data?.commits ?? [];
  $: totalCount = $historyQuery.data?.totalCount ?? 0;

  function formatTimestamp(
    timestamp: { seconds?: bigint; nanos?: number } | undefined,
  ): string {
    if (!timestamp?.seconds) return "";
    const date = new Date(Number(timestamp.seconds) * 1000);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days} day${days !== 1 ? "s" : ""} ago`;
    if (hours > 0) return `${hours} hour${hours !== 1 ? "s" : ""} ago`;
    if (minutes > 0) return `${minutes} minute${minutes !== 1 ? "s" : ""} ago`;
    return "just now";
  }
</script>

<div class="flex flex-col h-full bg-white">
  <div class="px-4 py-3 border-b border-slate-200">
    <div class="font-medium text-sm flex items-center gap-x-2">
      <Clock size={14} class="text-slate-500" />
      Commit History
    </div>
    {#if totalCount > 0}
      <div class="text-xs text-slate-400 mt-0.5">
        {totalCount} commit{totalCount !== 1 ? "s" : ""}
      </div>
    {/if}
  </div>

  <div class="flex-1 overflow-y-auto">
    {#if $historyQuery.isLoading}
      <div class="p-4 text-center text-sm text-slate-500">
        Loading commit history...
      </div>
    {:else if commits.length === 0}
      <div class="p-4 text-center text-sm text-slate-500">No commits found</div>
    {:else}
      {#each commits as commit}
        <div class="px-4 py-3 border-b border-slate-100 hover:bg-slate-50">
          <div class="flex items-start gap-x-3">
            <div class="flex-shrink-0 mt-0.5">
              <div class="w-2 h-2 rounded-full bg-primary-500"></div>
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium text-slate-800 break-words">
                {commit.message}
              </div>
              <div
                class="flex items-center gap-x-3 mt-1 text-xs text-slate-400"
              >
                <span class="font-mono text-slate-500">{commit.shortHash}</span>
                <span class="flex items-center gap-x-1">
                  <User size={10} />
                  {commit.authorName}
                </span>
                <span>{formatTimestamp(commit.timestamp)}</span>
              </div>
            </div>
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>
