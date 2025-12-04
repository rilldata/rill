<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import GithubConnectionDialog from "@rilldata/web-admin/features/projects/github/GithubConnectionDialog.svelte";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import {
    getRepoNameFromGitRemote,
    getGitUrlFromRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let organization: string;
  export let project: string;

  $: ({ instanceId } = $runtime);

  $: proj = createAdminServiceGetProject(organization, project);
  $: ({
    project: { gitRemote, managedGitId, subpath, prodBranch },
  } = $proj.data);

  $: isGithubConnected = !!gitRemote;
  $: isManagedGit = !!managedGitId;
  $: repoName = getRepoNameFromGitRemote(gitRemote);
  $: githubLastSynced = useGithubLastSynced(instanceId);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    instanceId,
    organization,
    project,
  );
  // Github last synced might not always be available for projects not updated since we added commitedOn
  // So fallback to old way of approximating the last updated.
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;
</script>

{#if $proj.data}
  <div class="flex flex-col gap-y-1 max-w-[400px]">
    <span
      class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
    >
      GitHub
    </span>
    <div class="flex flex-col gap-x-1">
      {#if isGithubConnected && !isManagedGit}
        <div class="flex flex-row gap-x-1 items-center">
          <Github className="w-4 h-4" />
          <a
            href={getGitUrlFromRemote($proj.data?.project?.gitRemote)}
            class="text-gray-800 text-[12px] font-semibold font-mono leading-5 truncate"
            target="_blank"
            rel="noreferrer noopener"
          >
            {repoName}
          </a>
        </div>
        {#if subpath}
          <div class="flex items-center">
            <span class="font-mono">subpath</span>
            <span class="text-gray-800">
              : /{subpath}
            </span>
          </div>
        {/if}
        <div class="flex items-center">
          <span class="font-mono">branch</span>
          <span class="text-gray-800">
            : {prodBranch}
          </span>
        </div>
        {#if lastUpdated}
          <span class="text-gray-500 text-[11px] leading-4">
            Synced {lastUpdated.toLocaleString(undefined, {
              month: "short",
              day: "numeric",
              hour: "numeric",
              minute: "numeric",
            })}
          </span>
        {/if}
      {:else}
        <span class="my-1">
          Unlock the power of BI-as-code with GitHub-backed collaboration,
          version control, and approval workflows.
          <a
            href="https://docs.rilldata.com/deploy/deploy-dashboard/github-101"
            target="_blank"
            class="text-primary-600"
          >
            Learn more ->
          </a>
        </span>
        <GithubConnectionDialog {organization} {project} />
      {/if}
    </div>
  </div>
{/if}
