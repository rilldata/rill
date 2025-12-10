<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { GithubAccessManager } from "@rilldata/web-admin/features/projects/github/GithubAccessManager";
  import GithubConnectionDialog from "@rilldata/web-admin/features/projects/github/GithubConnectionDialog.svelte";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    getGitUrlFromRemote,
    getRepoNameFromGitRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let organization: string;
  export let project: string;

  let dialogOpen = false;

  const githubAccessManager = new GithubAccessManager();

  $: ({ instanceId } = $runtime);

  $: proj = createAdminServiceGetProject(organization, project);
  $: ({ isLoading, error } = $proj);

  $: gitRemote = $proj.data?.project?.gitRemote;
  $: managedGitId = $proj.data?.project?.managedGitId;
  $: subpath = $proj.data?.project?.subpath;
  $: prodBranch = $proj.data?.project?.prodBranch;

  $: isGithubConnected = !!gitRemote;
  $: isManagedGit = !!managedGitId;
  $: repoName = getRepoNameFromGitRemote(gitRemote);

  $: githubLastSynced = useGithubLastSynced(instanceId);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    instanceId,
    organization,
    project,
  );
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;

  function formatSyncTime(date: Date): string {
    return date.toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    });
  }
</script>

<div class="config-row">
  <div class="config-label">GitHub</div>
  <div class="config-value">
    {#if isLoading}
      <Spinner status={EntityStatus.Running} size="14px" />
    {:else if error}
      <span class="text-red-600 text-sm">Error loading GitHub status</span>
    {:else if isGithubConnected && !isManagedGit}
      <div class="connected-content">
        <Github className="w-4 h-4 flex-shrink-0" />
        <a
          href={getGitUrlFromRemote(gitRemote)}
          class="repo-link"
          target="_blank"
          rel="noreferrer noopener"
        >
          {repoName}
        </a>
        <span class="separator">•</span>
        <span class="detail">branch: {prodBranch}</span>
        {#if subpath}
          <span class="separator">•</span>
          <span class="detail">/{subpath}</span>
        {/if}
        {#if lastUpdated}
          <span class="separator">•</span>
          <span class="sync-time">Synced {formatSyncTime(lastUpdated)}</span>
        {/if}
      </div>
    {:else}
      <div class="not-connected-content">
        <span class="not-connected-text">Not connected</span>
        <button
          class="connect-link"
          on:click={() => {
            void githubAccessManager.ensureGithubAccess();
            dialogOpen = true;
          }}
        >
          Connect →
        </button>
      </div>
    {/if}
  </div>
</div>

<GithubConnectionDialog
  {organization}
  {project}
  bind:open={dialogOpen}
  hideTrigger
/>

<style lang="postcss">
  .config-row {
    @apply flex items-center;
    @apply border-b border-slate-200;
    @apply min-h-[44px];
  }

  .config-row:last-child {
    @apply border-b-0;
  }

  .config-label {
    @apply w-[140px] flex-shrink-0;
    @apply px-4 py-3;
    @apply text-sm font-medium text-gray-600;
    @apply bg-slate-50;
    @apply border-r border-slate-200;
    @apply whitespace-nowrap;
  }

  .config-value {
    @apply flex-1 px-4 py-3;
    @apply text-sm;
  }

  .connected-content {
    @apply flex items-center;
    @apply gap-x-2 flex-wrap;
  }

  .repo-link {
    @apply font-semibold font-mono;
    @apply text-gray-800;
  }

  .repo-link:hover {
    @apply text-primary-600 underline;
  }

  .separator {
    @apply text-gray-400;
  }

  .detail {
    @apply text-gray-600;
  }

  .sync-time {
    @apply text-gray-500 text-xs;
  }

  .not-connected-content {
    @apply flex items-center gap-x-2;
  }

  .not-connected-text {
    @apply text-gray-600;
  }

  .connect-link {
    @apply text-primary-600;
    @apply cursor-pointer;
    @apply bg-transparent border-none p-0;
    @apply text-sm font-medium;
  }

  .connect-link:hover {
    @apply text-primary-700 underline;
  }
</style>
