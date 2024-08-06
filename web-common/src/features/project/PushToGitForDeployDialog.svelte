<script lang="ts">
  import { getRepoNameFromGithubUrl } from "@rilldata/web-common/features/project/github-utils";
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
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";

  export let open = false;
  export let githubUrl: string;
  export let subpath: string;

  $: repoName = getRepoNameFromGithubUrl(githubUrl);
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>To update this deployment, use GitHub</AlertDialogTitle>
      <AlertDialogDescription class="flex flex-col gap-y-2 pt-2">
        <div>
          This project has already been connected to a GitHub repo. Please use
          the command line to push changes to GitHub, which will update Rill
          Cloud.
        </div>
        <div class="w-fit mx-auto">
          <div class="flex flex-row gap-x-1 items-center">
            <Github className="w-4 h-4" />
            <a
              href={githubUrl}
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
        <CLICommandDisplay command="git push" className="w-fit mx-auto" />
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button on:click={() => (open = false)} type="secondary">Close</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
