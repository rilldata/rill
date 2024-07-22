<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetGithubUserStatusQueryKey,
    getAdminServiceGetProjectQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { getGithubData } from "@rilldata/web-admin/features/projects/github/GithubData";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { invalidateRuntimeQueries } from "@rilldata/web-common/runtime-client/invalidation";
  import type { AxiosError } from "axios";

  export let open = false;
  export let currentUrl: string;
  export let project: string;
  export let organization: string;

  let githubUrl = currentUrl;
  const githubData = getGithubData();
  const userRepos = githubData.userRepos;
  const status = githubData.status;
  const projectQuery = createAdminServiceGetProject(organization, project);

  $: repoSelections =
    $userRepos.data?.repos?.map((r) => ({
      value: r.url,
      label: `${r.owner}/${r.name}`,
    })) ?? [];

  const updateProject = createAdminServiceUpdateProject();
  async function updateGithubUrl() {
    const repo = $userRepos.data?.repos?.find((r) => r.url === githubUrl);
    if (!repo) return; // shouldnt happen

    await $updateProject.mutateAsync({
      name: project,
      organizationName: organization,
      data: {
        githubUrl,
        prodBranch: repo.defaultBranch,
      },
    });
    eventBus.emit("notification", {
      message: `Set github repo to ${githubUrl}`,
      type: "success",
    });
    void queryClient.refetchQueries(
      getAdminServiceGetProjectQueryKey(organization, project),
    );
    void queryClient.refetchQueries(
      getAdminServiceGetGithubUserStatusQueryKey(),
    );
    void invalidateRuntimeQueries(
      queryClient,
      $projectQuery.data.prodDeployment.runtimeInstanceId,
    );
    open = false;
  }

  function handleVisibilityChange() {
    if (document.visibilityState !== "visible") return;
    void githubData.refetch();
  }

  $: error = ($status.error ??
    $updateProject.error) as unknown as AxiosError<RpcStatus>;
</script>

<svelte:window on:visibilitychange={handleVisibilityChange} />

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <div class="flex flex-row gap-x-2">
      <Github size="28px" />
      <div class="flex flex-col">
        <AlertDialogHeader>
          <AlertDialogTitle>Select Github repository</AlertDialogTitle>
          <AlertDialogDescription class="flex flex-col gap-y-2">
            <span>
              Which Github repo would you like to connect to this Rill project?
            </span>
            {#if $status.isFetching}
              <div class="flex flex-row items-center ml-5 h-8">
                <Spinner status={EntityStatus.Running} />
              </div>
            {:else}
              <Select
                id="emails"
                label=""
                bind:value={githubUrl}
                options={repoSelections}
              />
            {/if}
            <span class="font-semibold">
              Note: Contents of this repo will replace your current Rill
              project.
            </span>
            {#if error}
              <div class="text-red-500 text-sm py-px">
                {error.response?.data?.message ?? error.message}
              </div>
            {/if}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter class="mt-5">
          <Button
            outline={false}
            type="link"
            on:click={() => githubData.reselectRepos()}
          >
            Choose other repos
          </Button>
          <Button type="secondary" on:click={() => (open = false)}>
            Cancel
          </Button>
          <Button
            type="primary"
            on:click={() => updateGithubUrl()}
            loading={$updateProject.isLoading}
          >
            Continue
          </Button>
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>
