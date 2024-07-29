<script lang="ts">
  import {
    createAdminServiceGetProject,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { GithubConnectionUpdater } from "@rilldata/web-admin/features/projects/github/GithubConnectionUpdater";
  import { getGithubData } from "@rilldata/web-admin/features/projects/github/GithubData";
  import GithubOverwriteConfirmationDialog from "@rilldata/web-admin/features/projects/github/GithubOverwriteConfirmationDialog.svelte";
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

  const githubConnectionUpdater = new GithubConnectionUpdater();
  const connectToGithubMutation =
    githubConnectionUpdater.connectToGithubMutation;
  const showOverwriteConfirmation =
    githubConnectionUpdater.showOverwriteConfirmation;
  async function updateGithubUrl(force: boolean) {
    if (
      !(await githubConnectionUpdater.update({
        organization,
        project,
        githubUrl,
        subpath,
        branch,
        force,
        instanceId: $projectQuery.data?.prodDeployment?.runtimeInstanceId ?? "",
      }))
    ) {
      return;
    }

    eventBus.emit("notification", {
      message: `Set github repo to ${githubUrl}`,
      type: "success",
    });
    open = false;
    advancedOpened = false;
  }

  $: error = ($status.error ??
    $connectToGithubMutation.error) as unknown as AxiosError<RpcStatus>;
  $: errorMessage = error ? error.response?.data?.message ?? error.message : "";
</script>

<Dialog bind:open>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <div class="flex flex-row gap-x-2 items-center">
        <Github size="40px" />
        <div class="flex flex-col gap-y-1">
          <DialogTitle>Select Github repository</DialogTitle>
          <DialogDescription>
            Choose a GitHub repo to house this project.
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
          bind:value={githubUrl}
          options={repoSelections}
          on:change={({ detail: newUrl }) => onRepoChange(newUrl)}
        />
        <span class="text-gray-500 mt-1">
          <span class="font-semibold">Note:</span> Contents of this repo will replace
          your current Rill project.
        </span>
      {/if}
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
            placeholder="subdirectory_path"
            bind:value={subpath}
            optional
          />
          <Input id="branch" label="Branch" bind:value={branch} optional />
        </CollapsibleContent>
      </Collapsible>
      {#if error}
        <div class="text-red-500 text-sm py-px">
          {errorMessage}
        </div>
      {/if}
    </DialogHeader>
    <DialogFooter class="mt-3">
      <Button
        outline={false}
        type="link"
        on:click={() => githubData.reselectRepos()}
      >
        Choose other repos
      </Button>
      <Button type="secondary" on:click={() => (open = false)}>Cancel</Button>
      <Button
        type="primary"
        loading={$connectToGithubMutation.isLoading}
        on:click={() => updateGithubUrl(false)}>Continue</Button
      >
    </DialogFooter>
  </DialogContent>
</Dialog>

<GithubOverwriteConfirmationDialog
  bind:open={$showOverwriteConfirmation}
  loading={$connectToGithubMutation.isLoading}
  error={errorMessage}
  {githubUrl}
  {subpath}
  onConfirm={() => updateGithubUrl(true)}
/>
