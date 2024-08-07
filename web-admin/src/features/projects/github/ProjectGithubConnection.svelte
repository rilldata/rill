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
  import { getRepoNameFromGithubUrl } from "@rilldata/web-common/features/project/github-utils";

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

  let hovered = false;

  const githubData = new GithubData();
  setGithubData(githubData);
  const userStatus = githubData.userStatus;
  const repoSelectionOpen = githubData.repoSelectionOpen;

  function confirmConnectToGithub() {
    // prompt reselection repos since a new repo might be created here.
    repoSelectionOpen.set(true);
    void githubData.reselectRepos();
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
  <div
    class="flex flex-col gap-y-1 max-w-[400px]"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
    role="region"
  >
    <span
      class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
    >
      GitHub
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
            <div class="min-w-4">
              {#if hovered}
                <EditIcon size="16px" />
              {/if}
            </div>
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
        <span class="my-1">
          Unlock the power of BI-as-code with GitHub-backed collaboration,
          version control, and approval workflows.
          <a
            href="https://docs.rilldata.com/deploy/existing-project/github-101"
            target="_blank"
            class="text-primary-600"
          >
            Learn more ->
          </a>
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

<GithubRepoSelectionDialog
  bind:open={$repoSelectionOpen}
  currentUrl={$proj.data?.project?.githubUrl}
  currentSubpath={$proj.data?.project?.subpath}
  currentBranch={$proj.data?.project?.prodBranch}
  {organization}
  {project}
/>
