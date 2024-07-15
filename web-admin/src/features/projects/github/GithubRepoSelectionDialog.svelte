<script lang="ts">
  import {
    createAdminServiceGetGithubUserStatus,
    createAdminServiceListGithubUserRepos,
  } from "@rilldata/web-admin/client";
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

  export let open = false;
  export let onConnect: (url: string) => void;

  let repo = "";
  const githubUserStatus = createAdminServiceGetGithubUserStatus();
  $: githubUserRepos = createAdminServiceListGithubUserRepos({
    query: {
      enabled: !!$githubUserStatus.data?.hasAccess,
      refetchOnWindowFocus: true,
    },
  });
  $: repoSelections =
    $githubUserRepos.data?.repos?.map((r) => ({
      value: r.url,
      label: `${r.owner}/${r.name}`,
    })) ?? [];
</script>

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
          <AlertDialogDescription class="flex flex-col gap-y-1">
            <span>
              Which Github repo would you like to connect to this Rill project?
            </span>
            {#if $githubUserStatus.isLoading || $githubUserRepos.isLoading}
              <div class="flex flex-row items-center ml-5 h-8">
                <Spinner status={EntityStatus.Running} />
              </div>
            {:else}
              <Select
                id="repo-selector"
                bind:value={repo}
                label=""
                options={repoSelections}
              />
            {/if}
            <span class="font-semibold">
              Note: Contents of this repo will replace your current Rill
              project.
            </span>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter class="mt-5">
          <Button
            outline={false}
            type="link"
            href={$githubUserStatus.data?.grantAccessUrl}
            target="_blank"
          >
            Choose other repos
          </Button>
          <Button type="secondary" on:click={() => (open = false)}>
            Cancel
          </Button>
          <Button type="primary" on:click={() => onConnect(repo)}>
            Continue
          </Button>
        </AlertDialogFooter>
      </div>
    </div>
  </AlertDialogContent>
</AlertDialog>
