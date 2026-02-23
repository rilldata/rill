<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
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
    project: { gitRemote, managedGitId, subpath, primaryBranch },
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
  <div class="flex flex-row gap-x-1 w-full">
    {#if isGithubConnected && !isManagedGit}
      <div class="flex flex-col gap-y-1">
        <div class="flex items-center gap-x-1">
          <Github className="shrink-0 h-4 w-4" />
          <a
            href={getGitUrlFromRemote($proj.data?.project?.gitRemote)}
            class="text-fg-primary text-[12px] font-semibold font-mono leading-5 truncate"
            target="_blank"
            rel="noreferrer noopener"
          >
            {repoName}
          </a>
        </div>
        <div class="flex flex-col text-[12px]">
          <span class="font-mono">branch: {primaryBranch}</span>
          {#if subpath}
            <span class="font-mono">subpath: /{subpath}</span>
          {/if}
        </div>
        {#if lastUpdated}
          <span class="text-fg-secondary text-[11px] leading-4">
            Synced {lastUpdated.toLocaleString(undefined, {
              month: "short",
              day: "numeric",
              hour: "numeric",
              minute: "numeric",
            })}
          </span>
        {/if}
      </div>
    {:else}
      <span class="my-1 text-fg-tertiary">
        Unlock the power of BI-as-code with GitHub-backed collaboration, version
        control, and approval workflows.
        <a
          href="https://docs.rilldata.com/developers/deploy/deploy-dashboard/github-101"
          target="_blank"
          class="text-primary-600"
        >
          Learn more ->
        </a>
      </span>
    {/if}
  </div>
{/if}
