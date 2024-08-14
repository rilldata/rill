<script lang="ts">
  import ConnectToGithubButton from "@rilldata/web-admin/features/projects/github/ConnectToGithubButton.svelte";
  import DisconnectProjectConfirmDialog from "@rilldata/web-admin/features/projects/github/DisconnectProjectConfirmDialog.svelte";
  import {
    GithubData,
    setGithubData,
  } from "@rilldata/web-admin/features/projects/github/GithubData";
  import GithubRepoSelectionDialog from "@rilldata/web-admin/features/projects/github/GithubRepoSelectionDialog.svelte";
  import { Button, IconButton } from "@rilldata/web-common/components/button";
  import DisconnectIcon from "@rilldata/web-common/components/icons/DisconnectIcon.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetProject,
    createAdminServiceUploadProjectAssets,
  } from "@rilldata/web-admin/client";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { getRepoNameFromGithubUrl } from "@rilldata/web-common/features/project/github-utils";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { PencilIcon } from "lucide-svelte";

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
  let editDropdownOpen = false;

  const githubData = new GithubData();
  setGithubData(githubData);
  const userStatus = githubData.userStatus;
  const repoSelectionOpen = githubData.repoSelectionOpen;

  const deleteProjectConnection = createAdminServiceUploadProjectAssets();
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

  function editGithubConnection() {
    void githubData.startRepoSelection();
    behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubConnectStart,
      {
        is_fresh_connection: isGithubConnected,
      },
    );
  }

  function disconnectGithubConnection() {
    void behaviourEvent?.fireGithubIntentEvent(
      BehaviourEventAction.GithubDisconnect,
      {
        is_fresh_connection: isGithubConnected,
      },
    );
    void $deleteProjectConnection.mutateAsync({
      organization,
      project,
      data: {},
    });
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
          <DropdownMenu.Root bind:open={editDropdownOpen}>
            <DropdownMenu.Trigger>
              <IconButton>
                {#if hovered || editDropdownOpen}
                  <ThreeDot size="16px" />
                {/if}
              </IconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              <DropdownMenu.Item class="px-1 py-1">
                <Button on:click={editGithubConnection} type="text" compact>
                  <div class="flex flex-row items-center gap-x-2">
                    <EditIcon size="14px" />
                    <span class="text-xs">Edit</span>
                  </div>
                </Button>
              </DropdownMenu.Item>
              <DropdownMenu.Item class="px-1 py-1">
                <Button
                  on:click={() => (disconnectConfirmOpen = true)}
                  type="text"
                  compact
                >
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

<DisconnectProjectConfirmDialog
  bind:open={disconnectConfirmOpen}
  loading={$deleteProjectConnection.isLoading}
  onDisconnect={disconnectGithubConnection}
/>
