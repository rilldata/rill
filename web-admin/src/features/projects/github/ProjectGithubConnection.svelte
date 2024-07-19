<script lang="ts">
  import ConnectToGithubConfirmDialog from "@rilldata/web-admin/features/projects/github/ConnectToGithubConfirmDialog.svelte";
  import { GithubConnection } from "@rilldata/web-admin/features/projects/github/GithubConnection";
  import GithubRepoSelectionDialog from "@rilldata/web-admin/features/projects/github/GithubRepoSelectionDialog.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
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

  let confirmDialogOpen = false;
  let githubSelectionOpen = false;

  const githubConnection = new GithubConnection(() => {
    githubSelectionOpen = true;
  });
  const userStatus = githubConnection.userStatus;

  function connectToGithub() {
    confirmDialogOpen = true;
  }

  function confirmConnectToGithub() {
    confirmDialogOpen = false;
    void githubConnection.check();
  }

  function editGithubConnection() {
    // keep the github selection open while checking for user access
    githubSelectionOpen = true;
    void githubConnection.check();
  }

  function handleVisibilityChange() {
    if (document.visibilityState !== "visible") return;
    void githubConnection.refetch();
  }
</script>

<svelte:window on:visibilitychange={handleVisibilityChange} />

{#if $proj.data}
  <div class="flex flex-col gap-y-1 max-w-[400px]">
    <span class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
      >Github</span
    >
    <div class="flex items-start gap-x-1">
      <div class="py-0.5 mt-1">
        <Github className="w-4 h-4" />
      </div>
      <div class="flex flex-col">
        {#if isGithubConnected}
          <div class="flex flex-row gap-x-1 items-center">
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
  onContinue={confirmConnectToGithub}
/>

<GithubRepoSelectionDialog
  bind:open={githubSelectionOpen}
  currentUrl={$proj.data?.project?.githubUrl}
  {organization}
  {project}
/>
