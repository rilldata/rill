<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetGithubUserStatusQueryKey,
    getAdminServiceGetProjectQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { getGithubData } from "@rilldata/web-admin/features/projects/github/GithubData";
  import UpdateGithubRepoButton from "@rilldata/web-admin/features/projects/github/UpdateGithubRepoButton.svelte";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import {
    Collapsible,
    CollapsibleTrigger,
    CollapsibleContent,
  } from "@rilldata/web-common/components/collapsible";
  import { Button } from "@rilldata/web-common/components/button";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { invalidateRuntimeQueries } from "@rilldata/web-common/runtime-client/invalidation";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import type { AxiosError } from "axios";

  export let open = false;
  export let currentUrl: string;
  export let currentSubpath: string;
  export let currentBranch: string;
  export let project: string;
  export let organization: string;

  let githubUrl = currentUrl;
  let subpath = currentSubpath;
  let branch = currentBranch;
  let advancedOpened = false;

  const githubData = getGithubData();
  const userRepos = githubData.userRepos;
  const status = githubData.status;
  const projectQuery = createAdminServiceGetProject(organization, project);

  $: repoSelections =
    $userRepos.data?.repos?.map((r) => ({
      value: r.url,
      label: `${r.owner}/${r.name}`,
    })) ?? [];
  function onRepoChange(newUrl: string) {
    const repo = $userRepos.data?.repos?.find((r) => r.url === newUrl);
    if (!repo) return; // shouldnt happen

    subpath = "";
    branch = repo.defaultBranch;
  }
  $: if (!githubUrl && repoSelections.length === 1) {
    onRepoChange(repoSelections[0].value);
  }

  const updateProject = createAdminServiceUpdateProject();
  async function updateGithubUrl() {
    const repo = $userRepos.data?.repos?.find((r) => r.url === githubUrl);
    if (!repo) return; // shouldnt happen

    await $updateProject.mutateAsync({
      name: project,
      organizationName: organization,
      data: {
        githubUrl,
        subpath,
        prodBranch: branch,
      },
    });
    eventBus.emit("notification", {
      message: `Set github repo to ${githubUrl}`,
      type: "success",
    });
    void queryClient.refetchQueries(
      getAdminServiceGetProjectQueryKey(organization, project),
      {
        // avoid refetching createAdminServiceGetProjectWithBearerToken
        exact: true,
      },
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

  $: error = ($status.error ??
    $updateProject.error) as unknown as AxiosError<RpcStatus>;
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent>
    <div class="flex flex-row gap-x-2">
      <Github size="28px" />
      <div class="flex flex-col">
        <DialogHeader>
          <DialogTitle>Select Github repository</DialogTitle>
          <DialogDescription class="flex flex-col gap-y-2">
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
                on:change={({ detail: newUrl }) => onRepoChange(newUrl)}
              />
            {/if}
            <span>
              <span class="font-semibold">Note:</span> Contents of this repo will
              replace your current Rill project.
            </span>
            <Collapsible bind:open={advancedOpened}>
              <CollapsibleTrigger asChild let:builder>
                <Button builders={[builder]} type="text">
                  {#if advancedOpened}
                    <CaretUpIcon size="16px" />
                  {:else}
                    <CaretDownIcon size="16px" />
                  {/if}
                  <span class="text-sm">Advanced options</span>
                </Button>
              </CollapsibleTrigger>
              <CollapsibleContent class="ml-6 flex flex-col gap-y-2">
                <Input
                  id="subpath"
                  label="Subpath"
                  placeholder="/subdirectory_path"
                  bind:value={subpath}
                  optional
                />
                <Input
                  id="branch"
                  label="Branch"
                  bind:value={branch}
                  optional
                />
              </CollapsibleContent>
            </Collapsible>
            {#if error}
              <div class="text-red-500 text-sm py-px">
                {error.response?.data?.message ?? error.message}
              </div>
            {/if}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter class="mt-5">
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
          <UpdateGithubRepoButton
            loading={$updateProject.isLoading}
            onConnect={updateGithubUrl}
          />
        </DialogFooter>
      </div>
    </div>
  </DialogContent>
</Dialog>
