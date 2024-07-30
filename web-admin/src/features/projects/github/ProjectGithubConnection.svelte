<script lang="ts">
  import ConnectToGithubButton from "@rilldata/web-admin/features/projects/github/ConnectToGithubButton.svelte";
  import {
    GithubData,
    setGithubData,
  } from "@rilldata/web-admin/features/projects/github/GithubData";
  import GithubRepoSelectionDialog from "@rilldata/web-admin/features/projects/github/GithubRepoSelectionDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { getRepoNameFromGithubUrl } from "@rilldata/web-admin/features/projects/github/github-utils";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isGithubConnected = !!$proj.data?.project?.githubUrl;
  $: repoName =
    $proj.data?.project?.githubUrl &&
    getRepoNameFromGithubUrl($proj.data.project.githubUrl);
  $: subpath = $proj.data?.project?.subpath;
  $: githubLastSynced = useDashboardsLastUpdated(
    $runtime.instanceId,
    organization,
    project,
  );

  const githubData = new GithubData();
  setGithubData(githubData);
  const userStatus = githubData.userStatus;
  const repoSelectionOpen = githubData.repoSelectionOpen;

  function confirmConnectToGithub() {
    void githubData.startRepoSelection();
    behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubConnectStart,
      {
        is_fresh_connection: isGithubConnected,
      },
    );
  }

  function editGithubConnection() {
    void githubData.startRepoSelection();
    behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubConnectStart,
      {
        is_fresh_connection: isGithubConnected,
      },
    );
  }
</script>

{#if $proj.data}
  <div class="flex flex-col gap-y-1 max-w-[400px]">
    <span
      class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
    >
      Github
    </span>
    <div class="flex flex-col gap-x-1">
      {#if isGithubConnected}
        <div class="flex flex-row gap-x-1 items-center">
          <Github className="w-4 h-4" />
          <a
            href={$proj.data?.project?.githubUrl}
            class="text-gray-800 text-[12px] font-semibold font-mono leading-5 truncate"
            target="_blank"
            rel="noreferrer noopener"
          >
            {repoName}
          </a>
          <Button on:click={editGithubConnection} type="ghost" compact>
            <EditIcon size="16px" />
          </Button>
        </div>
        {#if subpath}
          <div class="flex items-center">
            <span class="font-mono">subpath</span>
            <span class="text-gray-800">
              : /{subpath}
            </span>
          </div>
        {/if}
        {#if $githubLastSynced}
          <span class="text-gray-500 text-[11px] leading-4">
            Synced {$githubLastSynced.toLocaleString(undefined, {
              month: "short",
              day: "numeric",
              hour: "numeric",
              minute: "numeric",
            })}
          </span>
        {/if}
      {:else}
        <span class="mt-1">
          Unlock the power of BI-as-code with Github-backed collaboration,
          version control, and approval workflows.<br />
          <a href="https://docs.rilldata.com" target="_blank">Learn more</a>
        </span>
        <ConnectToGithubButton
          onContinue={confirmConnectToGithub}
          loading={$userStatus.isFetching}
          {isGithubConnected}
        />
      {/if}
    </div>
  </div>
{/if}

{#if $repoSelectionOpen}
  <!-- unmount to make sure state is reset -->
  <GithubRepoSelectionDialog
    bind:open={$repoSelectionOpen}
    currentUrl={$proj.data?.project?.githubUrl}
    currentSubpath={$proj.data?.project?.subpath}
    currentBranch={$proj.data?.project?.prodBranch}
    {organization}
    {project}
  />
{/if}
