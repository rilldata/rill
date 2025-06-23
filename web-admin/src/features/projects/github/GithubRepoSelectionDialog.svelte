<script lang="ts">
  import {
    createAdminServiceGetProject,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { extractGithubConnectError } from "@rilldata/web-admin/features/projects/github/github-errors";
  import { getGithubData } from "@rilldata/web-admin/features/projects/github/GithubData";
  import GithubOverwriteConfirmDialog from "@rilldata/web-admin/features/projects/github/GithubOverwriteConfirmDialog.svelte";
  import { ProjectGithubConnectionUpdater } from "@rilldata/web-admin/features/projects/github/ProjectGithubConnectionUpdater";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
  } from "@rilldata/web-common/components/collapsible";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CaretDownFilledIcon from "@rilldata/web-common/components/icons/CaretDownFilledIcon.svelte";
  import CaretRightFilledIcon from "@rilldata/web-common/components/icons/CaretRightFilledIcon.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { AxiosError } from "axios";

  export let open = false;
  export let organization: string;
  export let project: string;
  export let currentGithubRemote: string;
  export let currentSubpath: string;
  export let currentBranch: string;

  let advancedOpened = false;

  const githubData = getGithubData();
  const userRepos = githubData.userRepos;
  const status = githubData.status;
  const projectQuery = createAdminServiceGetProject(organization, project);

  $: repoSelections =
    $userRepos.data?.repos?.map((r) => ({
      value: r.remote,
      label: `${r.owner}/${r.name}`,
    })) ?? [];

  // update data from project, this is needed if the user never leaves the status page and this component is not unmounted
  $: githubConnectionUpdater = new ProjectGithubConnectionUpdater(
    organization,
    project,
    currentGithubRemote,
    currentSubpath,
    currentBranch,
  );

  $: connectToGithubMutation = githubConnectionUpdater.connectToGithubMutation;
  $: showOverwriteConfirmation =
    githubConnectionUpdater.showOverwriteConfirmation;
  $: githubRemote = githubConnectionUpdater.githubRemote;
  $: subpath = githubConnectionUpdater.subpath;
  $: branch = githubConnectionUpdater.branch;
  $: disableContinue = !$githubRemote || !$branch || $status.isFetching;

  function onSelectedRepoChange(newRemote: string) {
    const repo = $userRepos.data?.repos?.find((r) => r.remote === newRemote);
    if (!repo) return; // shouldnt happen

    githubConnectionUpdater.onSelectedRepoChange(repo);
  }
  $: if (!$githubRemote && $userRepos.data?.repos?.length) {
    githubConnectionUpdater.onSelectedRepoChange($userRepos.data.repos[0]);
  }

  $: if ($showOverwriteConfirmation) {
    // hide the selection dialog if overwrite confirmation dialog is open
    open = false;
  }

  async function updateGithubUrl(force: boolean) {
    let url = $githubRemote;
    const updateSucceeded = await githubConnectionUpdater.update({
      instanceId: $projectQuery.data?.prodDeployment?.runtimeInstanceId ?? "",
      force,
    });
    if (!updateSucceeded) return;

    eventBus.emit("notification", {
      message: `Set github repo to ${getGitUrlFromRemote(url)}`,
      type: "success",
    });
    open = false;
    advancedOpened = false;
  }

  $: error = extractGithubConnectError(
    ($status.error ??
      $connectToGithubMutation.error) as unknown as AxiosError<RpcStatus>,
  );

  function handleDialogClose() {
    githubConnectionUpdater.reset();
  }
</script>

<Dialog
  bind:open
  onOpenChange={(o) => {
    if (!o) handleDialogClose();
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <div class="flex flex-row gap-x-2 items-center">
        <Github size="40px" />
        <div class="flex flex-col gap-y-1">
          <DialogTitle>Select GitHub repository</DialogTitle>
          <DialogDescription>
            Choose a GitHub repo to push this project to.
          </DialogDescription>
        </div>
      </div>

      {#if $status.isFetching}
        <div class="flex flex-row items-center ml-5 h-20 w-full">
          <div class="m-auto w-10">
            <Spinner size="18px" status={EntityStatus.Running} />
          </div>
        </div>
      {:else}
        <Select
          id="emails"
          label="Repo"
          bind:value={$githubRemote}
          options={repoSelections}
          on:change={({ detail: newRemote }) => onSelectedRepoChange(newRemote)}
        />
        <span class="text-gray-500 mt-1">
          <span class="font-semibold">Note:</span> This current project will replace
          contents of the selected repo.
        </span>
      {/if}
      <Collapsible bind:open={advancedOpened}>
        <CollapsibleTrigger asChild let:builder>
          <Button builders={[builder]} type="text">
            {#if advancedOpened}
              <CaretDownFilledIcon size="12px" />
            {:else}
              <CaretRightFilledIcon size="12px" />
            {/if}
            <span class="text-sm">Advanced options</span>
          </Button>
        </CollapsibleTrigger>
        <CollapsibleContent class="ml-6 flex flex-col gap-y-2">
          <Input
            id="subpath"
            label="Subpath"
            placeholder="subdirectory_path"
            bind:value={$subpath}
            optional
          />
          <Input id="branch" label="Branch" bind:value={$branch} optional />
        </CollapsibleContent>
      </Collapsible>
      {#if error?.message}
        <div class="text-red-500 text-sm py-px">
          {error.message}
        </div>
      {/if}
    </DialogHeader>
    <DialogFooter class="mt-3">
      <!-- temporarily show this only during edit. in the long run we will not have edit -->
      {#if $githubRemote}
        <Button type="link" onClick={() => githubData.reselectRepos()}>
          Choose other repos
        </Button>
      {/if}
      <Button
        type="secondary"
        onClick={() => {
          open = false;
          handleDialogClose();
        }}>Cancel</Button
      >
      <Button
        type="primary"
        loading={$connectToGithubMutation.isPending}
        disabled={disableContinue}
        onClick={() => updateGithubUrl(false)}>Continue</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>

<GithubOverwriteConfirmDialog
  bind:open={$showOverwriteConfirmation}
  loading={$connectToGithubMutation.isPending}
  {error}
  githubRemote={$githubRemote}
  subpath={$subpath}
  onConfirm={() => updateGithubUrl(true)}
  onCancel={() => (open = true)}
/>
