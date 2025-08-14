<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import ConnectToGithubButton from "@rilldata/web-admin/features/projects/github/ConnectToGithubButton.svelte";
  import DisconnectProjectConfirmDialog from "@rilldata/web-admin/features/projects/github/DisconnectProjectConfirmDialog.svelte";
  import {
    GithubData,
    setGithubData,
  } from "@rilldata/web-admin/features/projects/github/GithubData";
  import GithubRepoSelectionDialog from "@rilldata/web-admin/features/projects/github/GithubRepoSelectionDialog.svelte";
  import NewGithubConnectionDialog from "@rilldata/web-admin/features/projects/github/NewGithubConnectionDialog.svelte";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import { Button, IconButton } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import DisconnectIcon from "@rilldata/web-common/components/icons/DisconnectIcon.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import {
    getRepoNameFromGitRemote,
    getGitUrlFromRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
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
  // So fallback to old way of aproximating the last updated.
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;

  let hovered = false;
  let editDropdownOpen = false;

  const githubData = new GithubData();
  setGithubData(githubData);
  const userStatus = githubData.userStatus;
  const repoSelectionOpen = githubData.repoSelectionOpen;

  let disconnectConfirmOpen = false;

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

  function disconnectGithubConnect() {
    disconnectConfirmOpen = true;
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
          <DropdownMenu.Root bind:open={editDropdownOpen}>
            <DropdownMenu.Trigger>
              <IconButton>
                {#if hovered || editDropdownOpen}
                  <ThreeDot size="16px" />
                {/if}
              </IconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              <!-- Disabling for now, until we figure out how to do this  -->
              <!--              <DropdownMenu.Item class="px-1 py-1">-->
              <!--                <Button onClick={editGithubConnection} type="text" compact>-->
              <!--                  <div class="flex flex-row items-center gap-x-2">-->
              <!--                    <EditIcon size="14px" />-->
              <!--                    <span class="text-xs">Edit</span>-->
              <!--                  </div>-->
              <!--                </Button>-->
              <!--              </DropdownMenu.Item>-->
              <DropdownMenu.Item class="px-1 py-1">
                <Button onClick={disconnectGithubConnect} type="text" compact>
                  <div class="flex flex-row items-center gap-x-2">
                    <DisconnectIcon size="14px" />
                    <span class="text-xs">Disconnect</span>
                  </div>
                </Button>
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
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
        <NewGithubConnectionDialog {organization} {project} />
      {/if}
    </div>
  </div>
{/if}

<DisconnectProjectConfirmDialog
  bind:open={disconnectConfirmOpen}
  {organization}
  {project}
/>
