<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import {
    getRepoNameFromGitRemote,
    getGitUrlFromRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;
  export let project: string;

  const runtimeClient = useRuntimeClient();

  $: proj = createAdminServiceGetProject(organization, project);
  $: ({
    project: { gitRemote, managedGitId, subpath, primaryBranch },
  } = $proj.data);

  $: isGithubConnected = !!gitRemote;
  $: isManagedGit = !!managedGitId;
  $: repoName = getRepoNameFromGitRemote(gitRemote);
  $: githubLastSynced = useGithubLastSynced(runtimeClient);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    runtimeClient,
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
          <span class="font-mono">{m.github_branch()}: {primaryBranch}</span>
          {#if subpath}
            <span class="font-mono">{m.github_subpath()}: /{subpath}</span>
          {/if}
        </div>
        {#if lastUpdated}
          <span class="text-fg-secondary text-[11px] leading-4">
            {m.github_synced()} {lastUpdated.toLocaleString(undefined, {
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
        {m.github_unlock_bi_as_code()}
        <span class="whitespace-nowrap">
          <a
            href="https://docs.rilldata.com/developers/deploy/deploy-dashboard/github-101"
            target="_blank"
            class="text-primary-600"
          >
            {m.common_learn_more()} ->
          </a>
        </span>
      </span>
    {/if}
  </div>
{/if}
