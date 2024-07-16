<script lang="ts">
  import ConnectToGithubConfirmDialog from "@rilldata/web-admin/features/projects/github/ConnectToGithubConfirmDialog.svelte";
  import { GithubConnection } from "@rilldata/web-admin/features/projects/github/GithubConnection";
  import GithubRepoSelectionDialog from "@rilldata/web-admin/features/projects/github/GithubRepoSelectionDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
  } from "web-admin/src/client";
  import { useDashboardsLastUpdated } from "web-admin/src/features/dashboards/listing/selectors";
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

  let confirmDialogOpen = false;
  let githubSelectionOpen = false;

  const githubConnection = new GithubConnection(() => {
    if (isGithubConnected) {
      githubSelectionOpen = true;
    } else {
      confirmDialogOpen = true;
    }
  });
  const userStatus = githubConnection.userStatus;

  function connectToGithub() {
    void githubConnection.check();
  }

  const updateProject = createAdminServiceUpdateProject();
  async function updateGithubUrl(url: string) {
    await $updateProject.mutateAsync({
      name: project,
      organizationName: organization,
      data: {
        githubUrl: url,
        archiveAssetId: "",
      },
    });
  }

  function handleVisibilityChange() {
    if (document.visibilityState !== "visible") return;
    githubConnection.focused();
  }
</script>

<svelte:window on:visibilitychange={handleVisibilityChange} />

{#if $proj.data}
  <div class="flex flex-col gap-y-1 max-w-[400px]">
    <span class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
      >Github</span
    >
    <div class="flex items-start gap-x-1">
      <div class="py-0.5">
        <Github className="w-4 h-4" />
      </div>
      <div class="flex flex-col">
        {#if isGithubConnected}
          <div class="flex flex-row items-center">
            <a
              href={$proj.data?.project?.githubUrl}
              class="text-gray-800 text-[12px] font-semibold font-mono leading-5 truncate"
              target="_blank"
              rel="noreferrer noopener"
            >
              {repoName}
            </a>
            <Button
              on:click={() => (githubSelectionOpen = true)}
              type="ghost"
              compact
            >
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
          <span>
            Unlock the power of BI-as-code with Github-backed collaboration,
            version control, and approval workflows.
          </span>
          <Button
            type="primary"
            class="w-fit mt-1"
            loading={$userStatus.isFetching}
            on:click={connectToGithub}
          >
            Connect to Github
          </Button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<ConnectToGithubConfirmDialog
  bind:open={confirmDialogOpen}
  onContinue={() => {
    confirmDialogOpen = false;
    githubSelectionOpen = true;
  }}
/>

<GithubRepoSelectionDialog
  bind:open={githubSelectionOpen}
  onConnect={updateGithubUrl}
/>
