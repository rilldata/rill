<script lang="ts">
  import {
    getGitUrlFromRemote,
    getRepoNameFromGitRemote,
  } from "@rilldata/web-common/features/project/deploy/github-utils";
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

  export let open = false;
  export let gitRemote: string;
  export let subpath: string;

  $: repoName = getRepoNameFromGitRemote(gitRemote);
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>To update this deployment, use GitHub</AlertDialogTitle>
      <AlertDialogDescription class="flex flex-col gap-y-4 pt-2">
        <div>
          This project has already been connected to a GitHub repo. Please push
          changes directly to GitHub and the project in Rill Cloud will
          automatically be updated.
          <a
            href="https://docs.rilldata.com/deploy/deploy-dashboard/github-101"
            target="_blank"
          >
            Learn more ->
          </a>
        </div>
        <div class="w-fit mx-auto">
          <div class="flex flex-row gap-x-1 items-center">
            <Github className="w-4 h-4" />
            <a
              href={getGitUrlFromRemote(gitRemote)}
              class="text-gray-800 text-[12px] font-semibold font-mono leading-5 truncate"
              target="_blank"
              rel="noreferrer noopener"
            >
              {repoName}
            </a>
          </div>
          {#if subpath}
            <div class="flex items-center">
              <span class="font-mono">subpath</span>
              <span class="text-gray-800">
                : /{subpath}
              </span>
            </div>
          {/if}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button onClick={() => (open = false)} type="secondary">Close</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
