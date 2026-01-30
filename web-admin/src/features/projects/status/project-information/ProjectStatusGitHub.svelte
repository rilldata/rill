<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import GithubConnectionDialog from "@rilldata/web-admin/features/projects/github/GithubConnectionDialog.svelte";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    getRepoNameFromGitRemote,
    getGitUrlFromRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let organization: string;
  export let project: string;

  $: ({ instanceId } = $runtime);

  // Project data
  $: proj = createAdminServiceGetProject(organization, project);
  $: projectData = $proj.data?.project;
  $: gitRemote = projectData?.gitRemote;
  $: managedGitId = projectData?.managedGitId;
  $: primaryBranch = projectData?.primaryBranch;
  $: subpath = projectData?.subpath;
  $: isGithubConnected = !!gitRemote;
  $: isManagedGit = !!managedGitId;
  $: repoName = gitRemote ? getRepoNameFromGitRemote(gitRemote) : "";

  // Last synced
  $: githubLastSynced = useGithubLastSynced(instanceId);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    instanceId,
    organization,
    project,
  );
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;
</script>

<div class="info-cell github-cell">
  <div class="cell-header">
    <Github size="16px" color="#6b7280" />
    <span class="cell-label">GitHub</span>
    <Tooltip distance={8}>
      <a
        href="https://docs.rilldata.com/developers/deploy/deploy-dashboard/github-101"
        target="_blank"
        rel="noreferrer noopener"
        class="info-link"
      >
        <InfoCircle size="14px" color="#9ca3af" />
      </a>
      <TooltipContent slot="tooltip-content">
        <p class="tooltip-text">
          Unlock BI-as-code with GitHub-backed collaboration, version control,
          and approval workflows.
        </p>
      </TooltipContent>
    </Tooltip>
  </div>
  <div class="cell-content">
    {#if isGithubConnected && !isManagedGit}
      <div class="flex items-center gap-1">
        <Github size="16px" />
        <a
          href={getGitUrlFromRemote(gitRemote)}
          class="repo-link"
          target="_blank"
          rel="noreferrer noopener"
        >
          {repoName}
        </a>
      </div>
      {#if subpath}
        <div class="github-detail">
          <span class="detail-label">subpath</span>: /{subpath}
        </div>
      {/if}
      <div class="github-detail">
        <span class="detail-label">branch</span>: {primaryBranch}
      </div>
      {#if lastUpdated}
        <span class="synced-text">
          Synced {lastUpdated.toLocaleString(undefined, {
            month: "short",
            day: "numeric",
            hour: "numeric",
            minute: "numeric",
          })}
        </span>
      {/if}
    {:else}
      <GithubConnectionDialog {organization} {project} />
    {/if}
  </div>
</div>

<style lang="postcss">
  .info-cell {
    @apply flex flex-col gap-1.5;
  }

  .github-cell {
    @apply gap-2;
  }

  .cell-header {
    @apply flex items-center gap-1.5;
  }

  .cell-label {
    @apply text-xs font-medium text-gray-500 uppercase tracking-wide;
  }

  .cell-content {
    @apply flex flex-col gap-1;
  }

  .repo-link {
    @apply text-sm font-semibold text-gray-900;
    @apply font-mono truncate;
  }

  .repo-link:hover {
    @apply text-primary-600;
  }

  .github-detail {
    @apply text-sm text-gray-700;
  }

  .detail-label {
    @apply font-mono text-gray-500;
  }

  .synced-text {
    @apply text-xs text-gray-500;
  }

  .info-link {
    @apply flex items-center;
  }

  .info-link:hover :global(svg) {
    color: #6b7280;
  }

  .tooltip-text {
    @apply text-sm max-w-[200px];
  }
</style>
